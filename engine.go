package nora

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/maja42/gl"
	"github.com/maja42/gl/render"
	"github.com/maja42/glfw"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
	"github.com/sirupsen/logrus"
)

// Engine represents the user interface of this library.
// It's a singleton that can be used to access all functionality.
type Engine struct {
	window            *glfw.Window
	windowTitleUpdate time.Time
	windowTitle       string

	resizePolicy       ResizePolicy
	desiredAspectRatio float32
	windowResized      bool

	rendering sync.Mutex // ensures that only one render-function can be executed at once

	vSyncDelay  time.Duration // duration of a single frame (depends on monitor frequency)
	fps         FPSCounter    // measures frame number and fps
	renderStats atomic.Value  // atomically stores render statistics

	glSync                        // synchronization of OpenGL resources like buffer targets
	samplerManager samplerManager // manages samplers (=texture targets)

	// The following members members must not be overwritten directly:
	Camera   Camera
	Shaders  ShaderStore
	Textures TextureStore

	InteractionSystem InteractionSystem // user interaction (mouse, keyboard, ...)
}

var initLock sync.Mutex
var renderThread *render.RenderThread

// Init initializes OpenGL, GLFW and the underlying render thread.
func Init() error {
	initLock.Lock()
	renderThread = render.New()

	gl.Init(renderThread)

	if err := glfw.Init(renderThread, gl.ContextWatcher); err != nil {
		renderThread.Terminate()
		initLock.Unlock()
		return fmt.Errorf("init glfw: %w", err)
	}
	return nil
}

// Destroy destroys all remaining windows, frees any allocated resources and de-initializes the OpenGL and GLFW libraries.
// Stops the render thread afterwards.
func Destroy() {
	if engine != nil {
		logrus.Warnf("Engine was not destroyed before shutting down OpenGL.")
		// if the engine was not destroyed (eg. due to a panic), stop it now
		engine.Destroy()
	}
	glfw.Terminate()
	renderThread.Terminate()
	initLock.Unlock()
}

type ResizePolicy int

const (
	ResizeAdjustViewport  = ResizePolicy(iota) // viewport == window size; can lead to distortions
	ResizeKeepViewport                         // viewport stays the same; can lead to stripped content or unrenderable areas
	ResizeForbid                               // the window can't be resized
	ResizeKeepAspectRatio                      // the window will be resized to keep the original aspect ratio
)

// Currently, only single-window applications are supported.
//     Multiple windows have multiple contexts. This either requires automatic context-switching, or a context-object
//     instead of accessing gl-functions directly. --> Find out if multiple contexts/windows can be used simultaneously,
//     when having multiple renderThreads executed from different OS-threads.
var engineLock sync.Mutex
var engine *Engine // For global access

// CreateWindow opens a new window and initializes the engine.
// Must be called after the library is initialized.
// Call Render() on the returned engine object to start rendering. The window will stay hidden until Render() is called for the first time
// Call Destroy() to close the window and free resources.
func CreateWindow(settings Settings) (*Engine, error) {
	engineLock.Lock() // There can only be one window at a time (context-switching not implemented yet)

	resizeable := gl.TRUE
	if settings.ResizePolicy == ResizeForbid {
		resizeable = gl.FALSE
	}
	glfw.WindowHint(glfw.Resizable, resizeable)

	glfw.WindowHint(glfw.Samples, settings.Samples)

	glfw.WindowHint(glfw.Visible, 0) // Hide window until the render-function is called the first time

	if settings.WindowSize.IsZero() {
		settings.WindowSize = vmath.Vec2i{1280, 720} // default resolution
	}

	window, err := glfw.CreateWindow(settings.WindowSize[0], settings.WindowSize[1], settings.WindowTitle, settings.Monitor, nil)
	if err != nil {
		engineLock.Unlock()
		return nil, err
	}
	window.MakeContextCurrent()

	monitor := glfw.GetPrimaryMonitor()
	vidmode := monitor.GetVideoMode()

	var windowSize vmath.Vec2i
	windowSize[0], windowSize[1] = window.GetSize()

	var framebufferSize vmath.Vec2i
	framebufferSize[0], framebufferSize[1] = window.GetFramebufferSize()

	logrus.Infof("OpenGL version:   %s", gl.GetString(gl.VERSION))
	logrus.Infof("GLSL version:     %s", gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	logrus.Infof("Vendor:           %s", gl.GetString(gl.VENDOR))
	logrus.Infof("Renderer:         %s", gl.GetString(gl.RENDERER))
	logrus.Infof("Monitor:          %d x %d @ %dHz (%s)", vidmode.Width, vidmode.Height, vidmode.RefreshRate, monitor.GetName())
	logrus.Infof("Window size:      %s", windowSize.Format("%d x %d"))
	logrus.Infof("Framebuffer size: %s", framebufferSize.Format("%d x %d"))
	logrus.Infof("")

	cursorX, cursorY := window.GetCursorPos()

	engine = &Engine{
		window:             window,
		windowTitle:        settings.WindowTitle,
		resizePolicy:       settings.ResizePolicy,
		desiredAspectRatio: float32(settings.WindowSize[0]) / float32(settings.WindowSize[1]),

		vSyncDelay: time.Second / time.Duration(vidmode.RefreshRate),
		fps:        NewFPSCounter(),

		Camera:   NewOrthoCamera(),
		Shaders:  newShaderStore(),
		Textures: newTextureStore(),

		InteractionSystem: newInteractionSystem(windowSize, vmath.Vec2i{int(cursorX), int(cursorY)}),
	}

	engine.configureOpenGL()

	engine.samplerManager = newSamplerManager(&engine.Textures)
	engine.renderStats.Store(RenderStats{})

	// wire resize configuration
	engine.Camera.(*OrthoCamera).SetAspectRatio(engine.desiredAspectRatio, engine.desiredAspectRatio > 1)
	window.SetSizeCallback(engine.resizeCallback)
	window.SetMaximizeCallback(engine.maximizeCallback)

	// wire interaction system
	window.SetCursorPosCallback(engine.InteractionSystem.cursorPosCallback)
	window.SetMouseButtonCallback(engine.InteractionSystem.mouseButtonCallback)
	window.SetScrollCallback(engine.InteractionSystem.scrollCallback)
	window.SetKeyCallback(engine.InteractionSystem.keyCallback)

	assert.NoGLError("Engine setup")
	return engine, nil
}

func (n *Engine) configureOpenGL() {
	logrus.Info("Configuring OpenGL...")

	// OpenGL context configuration
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL)
	gl.ClearDepthf(1)

	gl.Enable(gl.CULL_FACE)

	// enable transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

type DrawFrameFunc func(elapsed time.Duration, renderState *RenderState) (stop bool)

// Render starts polling events and rendering to the window.
// The given frameFunc is called every frame, until it returns 1 (="stop"), or the window is closed / the engine is destroyed.
// Returns 1 if the window was closed and the engine must be destroyed. In this case no subsequent calls to "Render()" are permitted.
//
// Note that user interactions are polled once every frame. User input callbacks cannot race with the render function.
// Note that there should always be one active render function, otherwise the window will not respond anymore due to the missing event polling.
func (n *Engine) Render(frameFunc DrawFrameFunc) bool {
	n.rendering.Lock()
	defer n.rendering.Unlock()
	n.window.Show()

	shouldClose := false
	for {
		shouldClose = engine.window.ShouldClose()
		if shouldClose {
			break
		}
		if stop := engine.renderFrame(frameFunc); stop {
			break
		}
	}
	assert.NoGLError("Render loop end")
	return shouldClose
}

func (n *Engine) renderFrame(frameFunc DrawFrameFunc) bool {
	frame, elapsed, framerate := n.fps.NextFrame()
	n.handleResize()

	if time.Since(n.windowTitleUpdate) >= 100*time.Millisecond {
		n.window.SetTitle(fmt.Sprintf("%s [%.2ffps]", n.windowTitle, framerate))
		n.windowTitleUpdate = time.Now()
	}

	renderState := newRenderState(n.Camera, &n.Shaders, &n.samplerManager)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	stop := frameFunc(elapsed, renderState)
	assert.True(renderState.TransformStack.Size() == 1, "Transform stack: not empty after rendering")

	renderStats := RenderStats{
		Frame:           frame,
		Framerate:       framerate,
		TotalDrawCalls:  renderState.totalDrawCalls,
		TotalPrimitives: renderState.totalPrimitives,
	}

	// swapbuffers waits until the next vsync (if swapinterval is 1).
	// This means that the render-thread will be blocked while waiting and no other gl-commands can be executed.
	// To circumvent this, we wait until the frame is nearly over before issuing the call
	delay := n.vSyncDelay - time.Since(n.fps.lastFrame)
	delay = time.Duration(float32(delay) * 0.5)
	time.Sleep(roundMillis(delay))

	n.window.SwapBuffers()
	assert.NoGLError("Render frame %d", frame)
	n.renderStats.Store(renderStats)
	renderThread.Sync()
	glfw.PollEvents()
	return stop
}

func (n *Engine) Destroy() {
	logrus.Debug("Waiting for current frame to finish")
	n.window.SetShouldClose(true)
	n.rendering.Lock() // Ensures that the render loop stopped and can't be started anymore

	assert.NoGLError("Engine shutting down")
	logrus.Debug("Shutting down engine")

	n.InteractionSystem.RemoveAll()
	n.Shaders.UnloadAll()
	n.Textures.UnloadAll()
	assert.NoGLError("Engine shut down")

	n.window.Destroy()
	gl.CheckError() // for some reason, window.Destroy() succeeds, but triggers a gl error; ignore that error

	engine = nil
	engineLock.Unlock()
	logrus.Info("Engine shut down")
}

func (n *Engine) resizeCallback(_ *glfw.Window, _ int, _ int) {
	n.windowResized = true
}

func (n *Engine) maximizeCallback(_ *glfw.Window, _ bool) {
	n.windowResized = true
}

func (n *Engine) handleResize() {
	// TODO: This function might try to change the window size to keep the aspect ratio.
	// 		 If the window got maximized however, this is not possible and the call will do nothing.
	//		 Create black-borders in this scenario.

	if !n.windowResized {
		return
	}
	n.windowResized = false
	width, height := n.window.GetSize()

	switch n.resizePolicy {
	case ResizeAdjustViewport:
		gl.Viewport(0, 0, width, height)

	case ResizeKeepViewport:
		// do nothing

	case ResizeKeepAspectRatio:
		gl.Viewport(0, 0, width, height)

		if n.InteractionSystem.WindowSize()[0] == width { // the height was modified --> adjust width
			newWidth := int(float32(height) * n.desiredAspectRatio)

			if newWidth != width {
				newHeight := int(float32(newWidth) / n.desiredAspectRatio) // ensures that there won't be two size-changes due to rounding
				logrus.Debugf("Changing window size to %d x %d", newWidth, newHeight)
				n.window.SetSize(newWidth, newHeight)
			}
		} else {
			newHeight := int(float32(width) / n.desiredAspectRatio)
			if newHeight != height {
				logrus.Debugf("Changing window size to %d x %d", width, newHeight)
				n.window.SetSize(width, newHeight)
			}
		}
	}

	n.InteractionSystem.updateWindowSize(vmath.Vec2i{width, height})
	logrus.Infof("Window size:    %v\n", n.InteractionSystem.WindowSize())
}

func roundMillis(d time.Duration) time.Duration {
	return d / time.Millisecond * time.Millisecond
}

// Window returns the underlying glfw window.
// Should usually not be required/accessed by the user.
func (n *Engine) Window() *glfw.Window {
	return n.window
}

// RenderThread returns the underlying render thread.
// Should usually not be required/accessed by the user.
func (n *Engine) RenderThread() *render.RenderThread {
	return renderThread
}

// SetClearColor changes the clear color (background color)
func (n *Engine) SetClearColor(color color.Color) {
	gl.ClearColor(color.R, color.G, color.B, color.A)
}

type RenderStats struct {
	Frame           uint64
	Framerate       float32 // frames per second
	TotalDrawCalls  int
	TotalPrimitives int
}

func (r *RenderStats) String() string {
	return fmt.Sprintf(""+
		"Frame         %d\n"+
		"Framerate     %.2f fps\n"+
		"Draw calls    %d\n"+
		"Primitives    %d",
		r.Frame, r.Framerate,
		r.TotalDrawCalls, r.TotalPrimitives)
}

// RenderStats returns statistics about the last rendered frame
func (n *Engine) RenderStats() RenderStats {
	return n.renderStats.Load().(RenderStats)
}

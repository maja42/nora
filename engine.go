package nora

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	glfw2 "github.com/go-gl/glfw/v3.3/glfw"
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

	running     <-chan struct{}
	vSyncDelay  time.Duration // duration of a single frame (depends on monitor frequency)
	fps         FPSCounter    // measures frame number and fps
	renderStats atomic.Value  // atomically stores render statistics

	glSync                        // synchronization of OpenGL resources like buffer targets
	samplerManager samplerManager // manages samplers (=texture targets)

	// The following members members must not be overwritten directly:
	Camera   Camera
	Shaders  ShaderStore
	Textures TextureStore

	DrawFrame         DrawFrameFunc     // frame drawing function
	InteractionSystem InteractionSystem // user interaction (mouse, keyboard, ...)
}

type DrawFrameFunc func(elapsed time.Duration, renderState *RenderState)

var initLock sync.Mutex
var renderThread *render.RenderThread

// Init initializes OpenGL, GLFW and the underlying render thread.
func Init() error {
	initLock.Lock()
	renderThread = render.New()

	if err := glfw.Init(renderThread, gl.ContextWatcher); err != nil {
		renderThread.Terminate()
		initLock.Unlock()
		return fmt.Errorf("init glfw: %w", err)
	}

	gl.Init(renderThread)
	return nil
}

// Destroy destroys all remaining windows, frees any allocated resources and de-initializes the OpenGL and GLFW libraries.
// Stops the render thread afterwards.
func Destroy() {
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

// Run opens a new window and starts the render loop.
// Must not be called before the library is initialized.
func Run(windowSize vmath.Vec2i, windowTitle string, monitor *glfw.Monitor, share *glfw.Window, resizePolicy ResizePolicy) (*Engine, error) {
	engineLock.Lock()

	resizeable := gl.TRUE
	if resizePolicy == ResizeForbid {
		resizeable = gl.FALSE
	}
	glfw.WindowHint(glfw.Resizable, resizeable)

	window, err := glfw.CreateWindow(windowSize[0], windowSize[1], windowTitle, monitor, share)
	if err != nil {
		engineLock.Unlock()
		return nil, err
	}
	window.MakeContextCurrent()

	monitor = glfw.GetPrimaryMonitor()
	vidmode := monitor.GetVideoMode()

	logrus.Infof("OpenGL version: %s\n", gl.GetString(gl.VERSION))
	logrus.Infof("GLSL version:   %s\n", gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	logrus.Infof("Vendor:         %s\n", gl.GetString(gl.VENDOR))
	logrus.Infof("Renderer:       %s\n", gl.GetString(gl.RENDERER))
	logrus.Infof("Monitor:        %d x %d @ %dHz (%s)", vidmode.Width, vidmode.Height, vidmode.RefreshRate, monitor.GetName())
	logrus.Infof("Window size:    %v\n", windowSize)

	cursorX, cursorY := window.GetCursorPos()

	running := make(chan struct{})
	engine = &Engine{
		window:             window,
		windowTitle:        windowTitle,
		resizePolicy:       resizePolicy,
		desiredAspectRatio: float32(windowSize[0]) / float32(windowSize[1]),

		running:    running,
		vSyncDelay: time.Second / time.Duration(vidmode.RefreshRate),
		fps:        NewFPSCounter(),

		Camera:   NewOrthoCamera(),
		Shaders:  newShaderStore(),
		Textures: newTextureStore(),

		DrawFrame:         func(time.Duration, *RenderState) {},
		InteractionSystem: newInteractionSystem(windowSize, vmath.Vec2i{int(cursorX), int(cursorY)}),
	}

	engine.configureOpenGL()

	engine.samplerManager = newSamplerManager(&engine.Textures)
	engine.renderStats.Store(RenderStats{})

	engine.Camera.(*OrthoCamera).SetAspectRatio(engine.desiredAspectRatio, engine.desiredAspectRatio > 1)

	window.SetSizeCallback(engine.resizeCallback)
	window.SetMaximizeCallback(engine.maximizeCallback)
	window.SetCursorPosCallback(engine.InteractionSystem.cursorPosCallback)
	window.SetMouseButtonCallback(engine.InteractionSystem.mouseButtonCallback)
	window.SetScrollCallback(engine.InteractionSystem.scrollCallback)
	window.SetKeyCallback(engine.InteractionSystem.keyCallback)

	var framebufferSize vmath.Vec2i
	framebufferSize[0], framebufferSize[1] = window.GetFramebufferSize()

	go func() {
		defer close(running)
		for !engine.window.ShouldClose() {
			engine.renderFrame()
		}
		assert.NoGLError("Engine execution")
		engine.cleanup()
		assert.NoGLError("Engine cleanup")
	}()
	return engine, nil
}

func (n *Engine) configureOpenGL() {
	logrus.Info("Configuring OpenGL...")

	// OpenGL context configuration
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearDepthf(1)

	gl.Enable(gl.CULL_FACE)

	// enable transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (n *Engine) cleanup() {
	n.InteractionSystem.RemoveAll()
	n.Shaders.UnloadAll()
	n.Textures.UnloadAll()

	engine = nil
	engineLock.Unlock()
}

func (n *Engine) Stop() {
	n.window.SetShouldClose(true)
}

func (n *Engine) Wait() {
	<-n.running
}

func (n *Engine) resizeCallback(_ *glfw.Window, _ int, _ int) {
	n.windowResized = true
}

func (n *Engine) maximizeCallback(_ *glfw2.Window, _ bool) {
	n.windowResized = true
}

func (n *Engine) renderFrame() {
	frame, elapsed, framerate := n.fps.NextFrame()
	n.handleResize()

	if time.Since(n.windowTitleUpdate) >= 100*time.Millisecond {
		n.window.SetTitle(fmt.Sprintf("%s [%.2ffps]", n.windowTitle, framerate))
		n.windowTitleUpdate = time.Now()
	}

	renderState := newRenderState(n.Camera, &n.Shaders, &n.samplerManager)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	n.DrawFrame(elapsed, renderState)
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

// Window returns the underlying render thread.
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

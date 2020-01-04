package nora

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"sync"
	"time"

	"github.com/maja42/gl"
	"github.com/maja42/gl/render"
	"github.com/maja42/glfw"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/color"
	"github.com/maja42/nora/math"
	"github.com/sirupsen/logrus"
)

type Nora struct {
	window             *glfw.Window
	windowSize         math.Vec2i
	windowTitleUpdate  time.Time
	windowTitle        string
	resizePolicy       ResizePolicy
	desiredAspectRatio float32
	windowResized      bool

	running    <-chan struct{}
	fps        FPSCounter
	vSyncDelay time.Duration

	glSync
	renderer       renderer
	samplerManager samplerManager

	Camera       Camera
	Shaders      ShaderStore
	Textures     TextureStore
	Scene        Scene
	Jobs         JobSystem
	Interactives InteractionSystem

	//cursorPos mgl32.Vec2
}

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
	ResizeAdjustViewport  = ResizePolicy(iota) // viewport == window size
	ResizeKeepViewport                         // viewport stays the same; can lead to distortions
	ResizeForbid                               // the window can't be resized
	ResizeKeepAspectRatio                      // the window will be resized to keep the original aspect ratio
)

// Currently, only single-window applications are supported.
//     Multiple windows have multiple contexts. This either requires automatic context-switching, or a context-object
//     instead of accessing gl-functions directly. --> Find out if multiple contexts/windows can be used simultaneously,
//     when having multiple renderThreads executed from different OS-threads.
var noraLock sync.Mutex
var nora *Nora // For global access

// Run opens a new window and starts the render loop.
func Run(windowSize math.Vec2i, windowTitle string, monitor *glfw.Monitor, share *glfw.Window, resizePolicy ResizePolicy) (*Nora, error) {
	noraLock.Lock()

	resizeable := gl.TRUE
	if resizePolicy == ResizeForbid {
		resizeable = gl.FALSE
	}
	glfw.WindowHint(glfw.Resizable, resizeable)

	window, err := glfw.CreateWindow(windowSize[0], windowSize[1], windowTitle, monitor, share)
	if err != nil {
		noraLock.Unlock()
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
	nora = &Nora{
		window:             window,
		windowSize:         windowSize,
		windowTitle:        windowTitle,
		resizePolicy:       resizePolicy,
		desiredAspectRatio: float32(windowSize[0]) / float32(windowSize[1]),

		running:    running,
		fps:        NewFPSCounter(),
		vSyncDelay: time.Second / time.Duration(vidmode.RefreshRate),

		renderer: newRenderer(),

		Camera:       NewOrthoCamera(),
		Shaders:      newShaderStore(),
		Textures:     newTextureStore(),
		Jobs:         newJobSystem(),
		Interactives: newInteractionSystem(mgl32.Vec2{float32(cursorX), float32(cursorY)}),
	}
	nora.Scene = newScene(&nora.Jobs)
	nora.samplerManager = newSamplerManager(&nora.Textures)

	nora.Camera.(*OrthoCamera).SetAspectRatio(nora.desiredAspectRatio, nora.desiredAspectRatio > 1)

	window.SetFramebufferSizeCallback(nora.resizeCallback)
	window.SetCursorPosCallback(nora.Interactives.cursorPosCallback)
	window.SetMouseButtonCallback(nora.Interactives.mouseButtonCallback)
	window.SetKeyCallback(nora.Interactives.keyCallback)

	var framebufferSize math.Vec2i
	framebufferSize[0], framebufferSize[1] = window.GetFramebufferSize()

	go func() {
		defer close(running)
		for !nora.window.ShouldClose() {
			nora.renderFrame()
		}
		nora.cleanup()
	}()
	return nora, nil
}

func (n *Nora) cleanup() {
	n.Jobs.RemoveAll()
	n.Interactives.RemoveAll()
	n.Shaders.UnloadAll()
	n.Textures.UnloadAll()
	n.Scene.DetachAndDestroyAll()

	nora = nil
	noraLock.Unlock()
}

func (n *Nora) Stop() {
	n.window.SetShouldClose(true)
}

func (n *Nora) Wait() {
	<-n.running
}

func (n *Nora) resizeCallback(_ *glfw.Window, _, _ int) {
	n.windowResized = true
}

func (n *Nora) renderFrame() {
	frame, elapsed, framerate := n.fps.NextFrame()
	n.handleResize()

	if time.Since(n.windowTitleUpdate) >= 100*time.Millisecond {
		n.window.SetTitle(fmt.Sprintf("%s [%.2ffps]", n.windowTitle, framerate))
		n.windowTitleUpdate = time.Now()
	}

	n.Jobs.run(elapsed)

	n.renderer.renderAll(n.Camera, &n.Shaders, &n.Scene, &n.samplerManager)

	// swapbuffers waits until the next vsync (if swapinterval is 1).
	// This means that the render-thread will be blocked while waiting and no other gl-commands can be executed.
	// To circumvent this, we wait until the frame is nearly over before issuing the call
	delay := n.vSyncDelay - time.Since(n.fps.lastFrame)
	delay = time.Duration(float32(delay) * 0.5)
	time.Sleep(roundMillis(delay))

	n.window.SwapBuffers()
	assert.NoGLError("Render frame %d", frame)
	renderThread.Sync()
	glfw.PollEvents()
}

func (n *Nora) handleResize() {
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

		if n.windowSize[0] == width { // the height was modified --> adjust width
			newWidth := int(float32(height) * n.desiredAspectRatio)

			if newWidth != width {
				newHeight := int(float32(newWidth) / n.desiredAspectRatio) // ensures that there won't be two size-changes due to rounding
				logrus.Debugf("Changing window size to %d x %d", newWidth, newHeight)
				n.window.SetSize(newWidth, newHeight)
				return
			}
		} else {
			newHeight := int(float32(width) / n.desiredAspectRatio)
			if newHeight != height {
				logrus.Debugf("Changing window size to %d x %d", width, newHeight)
				n.window.SetSize(width, newHeight)
				return
			}
		}
	}

	n.windowSize[0], n.windowSize[1] = width, height
	logrus.Infof("Window size:    %v\n", n.windowSize)
}

func roundMillis(d time.Duration) time.Duration {
	return d / time.Millisecond * time.Millisecond
}

// Window returns the underlying glfw window.
// Should usually not be required/accessed by the user.
func (n *Nora) Window() *glfw.Window {
	return n.window
}

// Window returns the underlying render thread.
// Should usually not be required/accessed by the user.
func (n *Nora) RenderThread() *render.RenderThread {
	return renderThread
}

func (n *Nora) SetClearColor(color color.Color) {
	gl.ClearColor(color.R, color.G, color.B, color.A)
}

package rendering

import (
	"sync"

	"github.com/maja42/logicat/rendering/color"

	"golang.org/x/mobile/gl"
)

// Engine is responsible for producing one frame after another.
type Engine struct {
	//glCtx gl.Context
	ctx *Context

	clearColor color.Color

	fps FPS
	//scene    *SceneGraph
	renderer *Renderer
}

// TODO: I think I don't like passing around the context everywhere.
// The context is nice (mocking, multi-ctx-support), but passing it around
// allows him to creep in everywhere and increase memory overhead.
// --> Implement it as a global variable + context-switcher?

// TODO: rename function and return scene, ctx, engine!
func NewEngine(glCtx gl.Context) *Engine {
	logger.Info("Setting up rendering engine...")

	ctx := &Context{
		Context: glCtx,
		Mutex:   sync.Mutex{},
	}
	ctx.scene = *NewSceneGraph(ctx)

	renderer := NewRenderer(ctx)
	renderer.SetClearColor(color.Black)

	engine := &Engine{
		ctx:        ctx,
		clearColor: color.Black,

		fps:      *NewFPS(),
		renderer: renderer,
	}
	return engine
}

func (e *Engine) Destroy() {
	e.ctx.scene.Destroy()
	//e.renderer.Destroy()
}

func (e *Engine) Context() *Context {
	return e.ctx
}

func (e *Engine) Scene() *SceneGraph {
	return &e.ctx.scene
}

func (e *Engine) SetClearColor(c color.Color) {
	e.renderer.SetClearColor(c)
}

func (e *Engine) Clear() {
	e.renderer.Clear()
}

func (e *Engine) RenderSingleFrame(cam Camera) {
	frame, elapsed, _ := e.fps.NextFrame()
	//if (this.resizeDirty) {
	//	this.handleRenderPanelChange();
	//	this.resizeDirty = false;
	//}

	e.ctx.scene.RunUpdateJobs(elapsed)

	e.renderer.Clear()
	iAssertNoGLError(e.ctx, "Clear frame %d", frame)

	e.renderer.RenderAll(cam)
	iAssertNoGLError(e.ctx, "Render frame %d", frame)
}

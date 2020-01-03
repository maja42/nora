package rendering

import (
	"github.com/go-gl/mathgl/mgl32/matstack"
	"github.com/maja42/logicat/rendering/color"
	"golang.org/x/mobile/gl"
)

// RenderTarget represents windows or textures.
type RenderTarget interface {
	// Clear()
	// TODO: more functionality at  https://www.sfml-dev.org/documentation/2.5.1/classsf_1_1RenderTarget.php
	Draw(*Mesh, *RenderState)
}

type RenderState struct {
	material *Material      // currently applied material
	sProg    *ShaderProgram // currently used shader program
	sProgID  sProgID

	camera Camera

	TransformStack matstack.MatStack

	totalDrawCalls  int
	totalPrimitives int
}

func NewRenderState(cam Camera) *RenderState {
	return &RenderState{
		camera:         cam,
		TransformStack: *matstack.NewMatStack(),
	}
}

// Renderer is responsible for drawing the world to a specific target
type Renderer struct {
	ctx        *Context
	clearColor color.Color

	texTargets *textureTargets // manages texture targets
}

func NewRenderer(ctx *Context) *Renderer {
	logger.Info("Setting up renderer...")

	renderer := &Renderer{
		ctx:        ctx,
		clearColor: color.Black,
		texTargets: NewTextureTargets(ctx, ctx.scene.resolveTexture),
	}

	// OpenGL context configuration
	ctx.Enable(gl.DEPTH_TEST)
	ctx.DepthFunc(gl.LESS)
	ctx.ClearDepthf(1)

	// enable transparency
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	renderer.SetClearColor(color.Black)
	return renderer
}

func (r *Renderer) SetClearColor(c color.Color) {
	r.clearColor = c
	r.ctx.ClearColor(c.R, c.G, c.B, c.A)
}

func (r *Renderer) Clear() {
	r.ctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) RenderAll(cam Camera) {
	renderState := NewRenderState(cam)

	r.ctx.Lock()
	defer r.ctx.Unlock()

	models := r.ctx.scene.borrowModels()
	defer r.ctx.scene.returnModels()

	for _, m := range models {
		renderState.TransformStack.LoadIdent()
		m.Draw(r, renderState)
		assert(len(renderState.TransformStack) == 1, "Transform stack: not empty after rendering")
	}

}

// implements RenderTarget.
// called by models.
// TODO: renderer + renderTarget somehow needs to be separated/defined more cleanly.
// Renderer should NOT implement renderTarget, instead there should be a window or texture.
// Renderer just draws everything correctly to a renderTarget...
func (r *Renderer) Draw(mesh *Mesh, renderState *RenderState) {
	ctx := r.ctx
	material := mesh.material

	sProgKey := material.sProgKey
	sProg, sProgID := ctx.scene.getShaderProgram(sProgKey)
	if sProg == nil {
		assertFail("shader %q is not loaded", sProgKey)
		return
	}

	if renderState.sProgID != sProgID {
		sProg.Use()

		vpMatrix := sProg.vpMatrixLocation
		if vpMatrix.Value >= 0 { // The shader supports view-projection transforms
			trans, _ := renderState.camera.Matrix()
			//fmt.Printf("%v\n", trans)
			ctx.UniformMatrix4fv(vpMatrix, trans[:])
		}

		renderState.sProg = sProg
		renderState.sProgID = sProgID
	}
	if renderState.material != material {
		material.apply(ctx, sProg, r.texTargets)
		renderState.material = material
	}

	modelTransform := sProg.modelTransformLocation
	if modelTransform.Value >= 0 { // The shader supports model transforms
		trans := renderState.TransformStack.Peek()
		ctx.UniformMatrix4fv(modelTransform, trans[:])
	}

	mesh.draw(renderState)
}

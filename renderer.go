package nora

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
	"github.com/sirupsen/logrus"
)

//// RenderTarget represents windows or textures.
//type RenderTarget interface {
//	// ClearTransform()
//	// TODO: more functionality at  https://www.sfml-dev.org/documentation/2.5.1/classsf_1_1RenderTarget.php
//	Draw(*Mesh, *RenderState)
//}

// renderer is responsible for drawing the world to a specific render target
type renderer struct {
	//texTargets *samplerManager // manages texture targets
}

func newRenderer() renderer {
	logrus.Info("Setting up renderer...")

	renderer := renderer{
		//texTargets: NewTextureTargets(ctx, ctx.scene.resolveTexture),
	}

	// OpenGL context configuration
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearDepthf(1)

	gl.Enable(gl.CULL_FACE)

	// enable transparency
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return renderer
}

func (r *renderer) renderAll(cam Camera, shaders *ShaderStore, scene *Scene, samplerManager *samplerManager) (int, int) {
	renderState := newRenderState(cam, shaders, samplerManager)

	models, ret := scene.borrowModels()
	defer ret()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	for _, m := range models {
		renderState.TransformStack.LoadIdent()
		m.Draw(renderState)
		assert.True(len(renderState.TransformStack) == 1, "Transform stack: not empty after rendering")
	}

	return renderState.totalDrawCalls, renderState.totalPrimitives
}

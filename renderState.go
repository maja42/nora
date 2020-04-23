package nora

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
)

type RenderState struct {
	camera         Camera
	shaders        *ShaderStore
	samplerManager *samplerManager

	// state
	material *Material // currently applied material
	sProgID  sProgID   // currently used shader program

	TransformStack vmath.MatStack4f

	// statistics
	totalDrawCalls  int
	totalPrimitives int
}

func newRenderState(cam Camera, shaders *ShaderStore, samplerManager *samplerManager) *RenderState {
	return &RenderState{
		camera:         cam,
		shaders:        shaders,
		samplerManager: samplerManager,
		TransformStack: *vmath.NewMatStack4f(),
	}
}

func (r *RenderState) applyMaterial(material *Material) *shaderProgram {
	sProgKey := material.sProgKey
	sProg := r.applyShader(sProgKey)
	if sProg == nil {
		return nil
	}

	if r.material != material {
		material.apply(sProg, r.samplerManager)
		r.material = material
	}

	modelTransform := sProg.modelTransformLocation
	if modelTransform.Value >= 0 { // The shader supports model transforms
		trans := r.TransformStack.Top()
		gl.UniformMatrix4fv(modelTransform, trans[:])
	}
	return sProg
}

func (r *RenderState) applyShader(sProgKey ShaderProgKey) *shaderProgram {
	sProg, sProgID := r.shaders.resolve(sProgKey)
	if sProg == nil {
		assert.Fail("shader %q is not loaded", sProgKey)
		return nil
	}

	if r.sProgID != sProgID {
		sProg.Use()

		vpMatrix := sProg.vpMatrixLocation
		if vpMatrix.Value >= 0 { // The shader supports view-projection transforms
			trans, _ := r.camera.Matrix()
			gl.UniformMatrix4fv(vpMatrix, trans[:])
		}

		r.sProgID = sProgID
	}
	return sProg
}

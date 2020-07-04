package main

import (
	"fmt"

	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/geometry/geo2d"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/rtree"
	"github.com/maja42/vmath"
)

func MousePos() vmath.Vec2f {
	pos := engine.InteractionSystem.MousePosClipSpace().AddScalar(1).DivScalar(2)
	return pos.Mul(resolution)
}

func UpdateVisualGeometry() {
	mat := nora.NewMaterial(shader.RGB_2D)
	geo := &nora.Geometry{}

	meshCount := 0

	addMesh := func(geo *nora.Geometry) {
		if geo.Empty() {
			return
		}
		var mesh *nora.Mesh
		if len(meshes) > meshCount {
			mesh = meshes[meshCount]
		} else {
			mesh = nora.NewMesh(mat)
			meshes = append(meshes, mesh)
		}
		mesh.SetGeometry(geo)
		meshCount++
	}

	addRect := func(bounds vmath.Rectf, color color.Color) {
		newGeo := geo2d.Border(bounds, 1, color)
		if geo.CanAppendVertexCount(newGeo.VertexCount()) {
			geo.AppendGeometry(newGeo)
			return
		}
		addMesh(geo)
		geo = newGeo
	}

	maxHeight := float32(tree.Height())
	tree.IterateInternalNodes(func(bounds vmath.Rectf, height int, leaf bool) bool {
		t := float32(height) / maxHeight
		col := color.InterpolateHSLA(color.Red, color.Green, t)
		addRect(bounds, col)
		return false
	})

	itemCount := 0
	tree.IterateItems(func(item rtree.Item) bool {
		it := item.(*Item)
		bounds := it.bounds
		addRect(bounds, it.color)
		itemCount++
		return false
	})
	addMesh(geo)

	for i := meshCount; i < len(meshes); i++ {
		meshes[i].Destroy()
	}
	meshes = meshes[:meshCount]
	engine.SetWindowTitle(fmt.Sprintf("%d elements", itemCount))
}

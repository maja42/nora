package main

import (
	"math/rand"
	"time"

	"github.com/maja42/glfw"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/geometry/geo2d"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/rtree"
	"github.com/maja42/vmath"
	"github.com/sirupsen/logrus"
)

var resolution = vmath.Vec2f{1280, 720}

var bulkSize = 5000

//var bulkSize = 1000000

type Item struct {
	bounds vmath.Rectf
	color  color.Color
}

func (i *Item) Bounds() vmath.Rectf {
	return i.bounds
}

func main() {
	rand.Seed(time.Now().UnixNano())
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
	err := run()
	if err != nil {
		logrus.Fatalln(err)
	}
}

var engine *nora.Engine
var tree *rtree.RTree
var meshes []*nora.Mesh

func run() error {
	if err := nora.Init(); err != nil {
		return err
	}
	defer nora.Destroy()
	var err error

	engine, err = nora.CreateWindow(nora.Settings{
		WindowTitle:  "R-Tree Demo",
		WindowSize:   resolution.Vec2i(),
		ResizePolicy: nora.ResizeAdjustViewport,
		Samples:      4,
	})
	if err != nil {
		return err
	}
	defer engine.Destroy()

	if err := engine.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}
	engine.SetClearColor(color.Gray(0.1))

	cam := engine.Camera.(*nora.OrthoCamera)
	cam.SetOrthoWidth(resolution[0])
	cam.SetPosition(resolution.DivScalar(2))

	tree = rtree.New()
	for i := 0; i < 5; i++ {
		tree.Insert(RandomItem())
	}

	UpdateVisualGeometry()

	isSelecting := false
	selectionMesh := nora.NewMesh(nora.NewMaterial(shader.RGB_2D))
	selectionArea := vmath.Rectf{}
	var selected []rtree.Item

	stop := false
	engine.InteractionSystem.OnKeyEvent(func(k glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		if action != glfw.Press && action != glfw.Repeat {
			return
		}
		switch k {
		case glfw.KeyEscape:
			stop = true
			return
		case glfw.KeyBackspace:
			tree.Clear()
		case glfw.KeySpace:
			tree.Insert(RandomItem())
		case glfw.KeyEnter:
			items := RandomItems(bulkSize)
			start := time.Now()
			for _, it := range items {
				tree.Insert(it)
			}
			logrus.Infof("insertion of %d items took %v", bulkSize, time.Since(start))
		case glfw.KeyB:
			items := RandomItems(bulkSize)
			start := time.Now()
			tree.BulkLoad(items)
			logrus.Infof("bulk insert of %d items took %v", bulkSize, time.Since(start))
		case glfw.KeyDelete:
			start := time.Now()
			for _, s := range selected {
				tree.Remove(s, nil)
			}
			logrus.Infof("deletion of %d items took %v", len(selected), time.Since(start))
			selected = nil
			selectionMesh.SetGeometry(&nora.Geometry{})
		}
		UpdateVisualGeometry()
	})

	engine.InteractionSystem.OnMouseButton(glfw.MouseButtonLeft, glfw.Press, func(_ glfw.ModifierKey) {
		pos := MousePos()
		item := Item{bounds: vmath.RectfFromPosSize(
			vmath.Vec2f{pos[0] - 5, pos[1] - 5},
			vmath.Vec2f{10, 10},
		), color: color.Cyan}

		tree.Insert(&item)
		UpdateVisualGeometry()
	})

	engine.InteractionSystem.OnMouseButtonEvent(func(button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		if button != glfw.MouseButtonRight {
			return
		}
		pos := MousePos()
		if action == glfw.Press {
			selectionArea = vmath.Rectf{pos, pos}
			isSelecting = true
			selectionMesh.SetGeometry(&nora.Geometry{})
		}
		if action == glfw.Release {
			if mod != glfw.ModControl {
				for _, s := range selected {
					s.(*Item).color = color.Cyan
				}
				selected = nil
			}
			mustCover := mod == glfw.ModShift
			start := time.Now()
			results := tree.Search(selectionArea, mustCover)
			logrus.Infof("search found %d items and took %v", len(results), time.Since(start))
			selected = append(selected, results...)
			for _, s := range selected {
				s.(*Item).color = color.Red
			}
			UpdateVisualGeometry()
			isSelecting = false
		}
	})

	engine.InteractionSystem.OnMouseMoveEvent(func(_, _ vmath.Vec2i) {
		if !isSelecting {
			return
		}
		selectionArea.Max = MousePos()
		geo := geo2d.Border(selectionArea, 2, color.Blue)
		selectionMesh.SetGeometry(geo)
	})

	engine.Render(func(elapsed time.Duration, renderState *nora.RenderState) bool {
		for _, m := range meshes {
			m.Draw(renderState)
		}
		if isSelecting {
			selectionMesh.Draw(renderState)
		}
		return stop
	})
	return nil
}

func RandomItems(count int) []rtree.Item {
	items := make([]rtree.Item, count)
	for i := 0; i < len(items); i++ {
		items[i] = RandomItem()
	}
	return items
}

func RandomItem() *Item {
	minSize := vmath.Vec2f{2, 2}
	maxSize := resolution.MulScalar(0.05)

	size := vmath.Vec2f{rand.Float32(), rand.Float32()}.
		Mul(maxSize.Sub(minSize)).
		Add(minSize)

	maxPos := resolution.Sub(size)

	item := Item{
		bounds: vmath.RectfFromPosSize(vmath.Vec2f{
			rand.Float32() * maxPos[0],
			rand.Float32() * maxPos[1],
		}, size),
		color: color.Cyan,
	}
	return &item
}

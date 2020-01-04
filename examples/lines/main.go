package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/glfw"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
	"github.com/maja42/nora/color"
	"github.com/maja42/nora/math"
	"github.com/sirupsen/logrus"
)

func main() {
	err := run()
	if err != nil {
		logrus.Fatalln(err)
	}
}

func run() error {
	if err := nora.Init(); err != nil {
		return err
	}
	defer nora.Destroy()

	n, err := nora.Run(math.Vec2i{1920, 1080}, "Demo", nil, nil, nora.ResizeKeepAspectRatio)
	if err != nil {
		return err
	}
	defer n.Wait()

	if err := n.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	n.SetClearColor(color.Gray(0.1))

	line := shapes.NewLine2D(0.06, shapes.BevelJoint, true)
	line.SetColor(color.Gray(0.6))
	n.Scene.Attach(line)

	line.AddPoints([]mgl32.Vec2{
		{0, 0},
		{1, 0},
		{1, 0.9},
		{0, 0.9},
		{0.5, 0.45},
	}...)

	line = shapes.NewLine2D(0.06, shapes.MitterJoint, true)
	line.SetColor(color.Gray(0.6))
	n.Scene.Attach(line)

	line.AddPoints([]mgl32.Vec2{
		{-0.1, 0},
		{-1.1, 0},
		{-1.1, 0.9},
		{-0.1, 0.9},
		{0.4, 0.45},
	}...)

	line = shapes.NewLine2D(0.006, shapes.MitterJoint, true)
	line.SetColor(color.Gray(0.6))
	n.Scene.Attach(line)

	line.AddPoints([]mgl32.Vec2{
		{-0.122, 0.05},
		{-1.05, 0.05},
		{-1.05, 0.85},
		{-0.122, 0.85},
		{0.32, 0.45},
	}...)

	n.Interactives.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		n.Stop()
	})

	return nil
}

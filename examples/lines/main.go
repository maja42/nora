package main

import (
	"time"

	"github.com/maja42/gl"
	"github.com/maja42/glfw"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})
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

	engine, err := nora.Run(vmath.Vec2i{1920, 1080}, "Demo", nil, nil, nora.ResizeKeepAspectRatio)
	if err != nil {
		return err
	}
	defer engine.Wait()

	if err := engine.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	engine.SetClearColor(color.Gray(0.1))

	line1 := shapes.NewLine2D(0.06, shapes.BevelJoint, true)
	line1.SetColor(color.Gray(0.6))
	line1.AddPoints([]vmath.Vec2f{
		{0, 0},
		{1, 0},
		{1, 0.9},
		{0, 0.9},
		{0.5, 0.45},
	}...)

	line2 := shapes.NewLine2D(0.06, shapes.MitterJoint, true)
	line2.SetColor(color.Gray(0.6))
	line2.AddPoints([]vmath.Vec2f{
		{-0.1, 0},
		{-1.1, 0},
		{-1.1, 0.9},
		{-0.1, 0.9},
		{0.4, 0.45},
	}...)

	line3 := shapes.NewLine2D(0.006, shapes.MitterJoint, true)
	line3.SetColor(color.Gray(0.6))
	line3.AddPoints([]vmath.Vec2f{
		{-0.122, 0.05},
		{-1.05, 0.05},
		{-1.05, 0.85},
		{-0.122, 0.85},
		{0.32, 0.45},
	}...)

	engine.DrawFrame = func(elapsed time.Duration, renderState *nora.RenderState) {
		line1.Draw(renderState)
		line2.Draw(renderState)
		line3.Draw(renderState)
	}
	var mode gl.Enum = gl.FILL
	engine.InteractionSystem.OnKey(glfw.KeySpace, glfw.Press, func(key glfw.ModifierKey) {
		go func() { // calling gl functions within a key-callback is currently not possible
			if mode == gl.FILL {
				mode = gl.LINE
			} else {
				mode = gl.FILL
			}
			gl.PolygonMode(mode)
		}()
	})
	engine.InteractionSystem.OnKeyEvent(func(k glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		if k != glfw.KeySpace {
			engine.Stop()
		}
	})
	return nil
}

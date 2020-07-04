package main

import (
	gomath "math"
	"math/rand"
	"time"

	"github.com/maja42/glfw"
	"github.com/maja42/vmath/math32"

	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
	"github.com/maja42/nora/color"
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

	engine, err := nora.CreateWindow(nora.Settings{
		WindowTitle:  "Triangle Demo",
		ResizePolicy: nora.ResizeKeepAspectRatio,
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
	cam.SetAspectRatio(float32(16)/9, true)

	tris := make([]*shapes.Triangle2D, 2000)

	for i := 0; i < len(tris); i++ {
		tri := shapes.NewTriangle2D()
		tri.SetUniformScale(0.01 + 0.05*rand.Float32())
		tri.SetPositionXY(-1+2*rand.Float32(), -1+2*rand.Float32())
		tri.SetColor(color.Color{
			R: rand.Float32(), G: rand.Float32(), B: rand.Float32(), A: 1,
		})
		tri.SetRotationZ(rand.Float32() * gomath.Pi * 2)
		tris[i] = tri
	}
	stop := false
	engine.InteractionSystem.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		stop = true
	})

	engine.Render(func(elapsed time.Duration, renderState *nora.RenderState) bool {
		if elapsed > 30*time.Millisecond { // clamp
			elapsed = 30 * time.Millisecond
		}

		mv := float32(elapsed/time.Millisecond) / 1000
		sz := float32(elapsed/time.Millisecond) / 1000
		rt := float32(elapsed/time.Millisecond) / 100
		for _, tri := range tris {
			tri.MoveXY((rand.Float32()-0.5)*mv, (rand.Float32()-0.5)*mv)
			tri.RotateZ((rand.Float32() - 0.5) * rt)
			xs := tri.GetScale().X() * 5
			tri.UniformScale(1 + ((rand.Float32() - 0.5) * sz / xs))
			scale := tri.GetScale()
			scale[0] = math32.Max(scale[0], 0.03)
			scale[1] = math32.Max(scale[1], 0.03)
			tri.SetScale(scale)

			tri.Draw(renderState)
		}

		return stop
	})
	return nil
}

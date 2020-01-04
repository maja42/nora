package main

import (
	"github.com/maja42/glfw"
	gomath "math"
	"math/rand"
	"time"

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

	cam := n.Camera.(*nora.OrthoCamera)
	cam.SetAspectRatio(float32(16)/9, true)

	tris := make([]*shapes.Triangle2D, 0)
	n.Jobs.Once(func(elapsed time.Duration) {
		for i := 0; i < 2000; i++ {
			tri := shapes.NewTriangle2D()
			tri.SetUniformScale(0.01 + 0.05*rand.Float32())
			tri.SetPositionXY(-1+2*rand.Float32(), -1+2*rand.Float32())
			tri.SetColor(color.Color{
				R: rand.Float32(), G: rand.Float32(), B: rand.Float32(), A: 1,
			})
			tri.SetRotationZ(rand.Float32() * gomath.Pi * 2)
			n.Scene.Attach(tri)
			tris = append(tris, tri)
		}

		n.Jobs.Add(func(elapsed time.Duration) {
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
			}
		})
	})

	n.Interactives.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		n.Stop()
	})

	return nil
}

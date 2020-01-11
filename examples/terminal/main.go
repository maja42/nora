package main

import (
	"math/rand"
	"time"

	"github.com/maja42/glfw"

	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
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

	engine, err := nora.Run(math.Vec2i{1920, 1080}, "Demo", nil, nil, nora.ResizeKeepAspectRatio)
	if err != nil {
		return err
	}
	defer engine.Wait()

	if err := engine.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	f, err := nora.LoadFont("builtin/fonts/ibm plex mono", "ibm_plex_mono_regular_32.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	cam := engine.Camera.(*nora.OrthoCamera)

	size := math.Vec2i{174, 56}
	term := shapes.NewTerminal(f, size, 0.8)
	term.SetPositionXY(cam.Left(), cam.Top())
	term.SetUniformScale(0.0006)
	engine.Scene.Attach(term)

	runes := f.Runes()
	engine.Jobs.AddFixed(100*time.Microsecond, func(elapsed time.Duration) {
		pos := math.Vec2i{
			rand.Intn(size[0]),
			rand.Intn(size[1]),
		}
		r := runes[rand.Intn(len(runes))]
		term.SetRune(pos, r)
	})

	engine.InteractionSystem.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		engine.Stop()
	})

	return nil
}

package main

import (
	"math/rand"
	"time"

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

	n, err := nora.Run(math.Vec2i{1920, 1080}, "Demo", nil, nil, nora.ResizeKeepAspectRatio)
	if err != nil {
		return err
	}
	defer n.Wait()

	if err := n.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	f, err := nora.LoadFont("builtin/fonts/ibm plex mono", "ibm_plex_mono_regular_32.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	cam := n.Camera.(*nora.OrthoCamera)

	size := math.Vec2i{174, 56}
	term := shapes.NewTerminal(f, size, 0.8)
	term.SetPositionXY(cam.Left(), cam.Top())
	term.SetUniformScale(0.0006)
	n.Scene.Attach(term)

	runes := f.Runes()
	n.Jobs.AddFixed(100*time.Microsecond, func(elapsed time.Duration) {
		pos := math.Vec2i{
			rand.Intn(size[0]),
			rand.Intn(size[1]),
		}
		r := runes[rand.Intn(len(runes))]
		term.SetRune(pos, r)
	})
	return nil
}

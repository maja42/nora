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

	roboto, err := nora.LoadFont("builtin/fonts/roboto/roboto_regular_65.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	monospace, err := nora.LoadFont("builtin/fonts/ibm plex mono/ibm_plex_mono_regular_32.xml")
	if err != nil {
		logrus.Errorf("Failed to load font: %s", err)
	}

	txt1 := shapes.NewText(roboto, "Nora rendering engine")
	txt1.SetUniformScale(0.1)
	txt1.MoveXY(-1.01, 0.775)

	txt2 := shapes.NewText(monospace, "To-Do:\n"+
		"\t ✓ Support textures\n"+
		"\t ✓ Support fonts\n"+
		"\t ✓ Support special characters: ⇆ ‡ ⅑ ↶ ₹\n"+
		"\t ✓ Press space to switch polygon mode\n"+
		"\t ❌ Support truetype fonts with kerning and hinting\n"+
		"\t - Make something nice with it\n"+
		"")
	txt2.SetUniformScale(0.08)
	txt2.MoveXY(-1.01, 0.65)

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

	engine.DrawFrame = func(elapsed time.Duration, renderState *nora.RenderState) {
		txt1.Draw(renderState)
		txt2.Draw(renderState)
	}
	engine.InteractionSystem.OnKeyEvent(func(k glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		if k != glfw.KeySpace {
			engine.Stop()
		}
	})
	return nil
}

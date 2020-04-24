package main

import (
	"time"

	"github.com/maja42/gl"
	"github.com/maja42/glfw"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/builtin/shapes"
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
		WindowTitle:  "Sprite Demo",
		ResizePolicy: nora.ResizeKeepAspectRatio,
	})
	if err != nil {
		return err
	}
	defer engine.Destroy()

	if err := engine.Shaders.LoadAll(shader.Builtins("builtin/shader")); err != nil {
		logrus.Errorf("Failed to load builtin shaders: %s", err)
	}

	engine.Textures.Load("sheep", &nora.TextureDefinition{
		Path: "examples/sprite/sheep.png",
		Properties: nora.TextureProperties{
			MinFilter: gl.LINEAR,
			MagFilter: gl.LINEAR,
			WrapS:     gl.REPEAT,
			WrapT:     gl.REPEAT,
		},
	})
	sprite := shapes.NewSprite()
	sprite.SetTexture("sheep")
	sprite.MoveXY(-0.5, -0.5)

	stop := false
	engine.InteractionSystem.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		stop = true
	})
	engine.Render(func(elapsed time.Duration, renderState *nora.RenderState) bool {
		sprite.Draw(renderState)
		return stop
	})
	return nil
}

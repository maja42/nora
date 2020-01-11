package main

import (
	"github.com/maja42/gl"
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
	engine.Scene.Attach(sprite)

	engine.InteractionSystem.OnKeyEvent(func(_ glfw.Key, _ int, _ glfw.Action, _ glfw.ModifierKey) {
		engine.Stop()
	})

	return nil
}

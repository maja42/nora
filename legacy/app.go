package main

import (
	"context"
	"strconv"
	"time"

	"github.com/maja42/logicat/rendering/color"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/maja42/logicat/rendering/resources/shader"

	"github.com/maja42/logicat/rendering/shapes"

	"github.com/maja42/logicat/rendering"
	"github.com/sirupsen/logrus"
	"golang.org/x/mobile/gl"
)

var logger = SetupLogger(logrus.DebugLevel)

type Application struct {
	ctx           *rendering.Context
	width, height int

	camera *rendering.OrthoCamera
	scene  *rendering.SceneGraph
	engine *rendering.Engine
}

/*
	TODO's:
		- Cleanup code, todos and comments
		- See what's up with the "configTexTarget" variable
		- Text: Error when rendering more than 6 characters. Something's odd with the VBO/IBO rendering backend...
            Comment out the indices for the first few triangles --> the last few triangles will not be rendered?!
		- Implement texture-unloading:
			for every new frame (or material?), invalidate the texture targets as "not needed anymore"?
			load textures afterwards?
		- the context should not be passed around anymore. Use context-switching instead!
		- care about exported/unexported stuff (move to internal package)
		- get rid of that stupid, broken gl/mobile thing. Just resize the window and see what happens...
*/

func (a *Application) Start(glCtx gl.Context) {
	defer func() {
		if err := rendering.GetGLError(glCtx); err != nil {
			logger.Errorf("OpenGL error at application exit: %s", err)
		}
	}()

	logger.Infof("OpenGL version: %s\n", glCtx.GetString(gl.VERSION))
	logger.Infof("GLSL version:   %s\n", glCtx.GetString(gl.SHADING_LANGUAGE_VERSION))
	logger.Infof("Vendor:         %s\n", glCtx.GetString(gl.VENDOR))
	logger.Infof("Renderer:       %s\n", glCtx.GetString(gl.RENDERER))

	//a.scene = rendering.NewSceneGraph(glCtx)
	a.engine = rendering.NewEngine(glCtx)
	a.engine.SetClearColor(color.Gray(0.1))

	a.scene = a.engine.Scene()
	a.ctx = a.engine.Context()
	a.camera = rendering.NewOrthoCamera()

	if err := a.scene.LoadShaderPrograms(shader.Builtins("rendering/resources/shader")); err != nil {
		logger.Errorf("Failed to load builtin shaders: %s", err)
	}
	if _, err := a.scene.LoadTexture("sheep", &rendering.TextureDefinition{
		Path: "sheep.png",
		Properties: rendering.TextureProperties{
			MinFilter: gl.LINEAR,
			MagFilter: gl.LINEAR,
			WrapS:     gl.REPEAT,
			WrapT:     gl.REPEAT,
		},
	}); err != nil {
		logger.Errorf("Failed to load texture: %s", err)
	}

	if _, err := a.scene.LoadTexture("comp1", &rendering.TextureDefinition{
		Path: "comp1.png",
		Properties: rendering.TextureProperties{
			MinFilter: gl.LINEAR,
			MagFilter: gl.LINEAR,
			WrapS:     gl.REPEAT,
			WrapT:     gl.REPEAT,
		},
	}); err != nil {
		logger.Errorf("Failed to load texture: %s", err)
	}
	if _, err := a.scene.LoadTexture("comp2", &rendering.TextureDefinition{
		Path: "comp2.png",
		Properties: rendering.TextureProperties{
			MinFilter: gl.LINEAR,
			MagFilter: gl.LINEAR,
			WrapS:     gl.REPEAT,
			WrapT:     gl.REPEAT,
		},
	}); err != nil {
		logger.Errorf("Failed to load texture: %s", err)
	}

	lucida65, err := rendering.LoadFont(a.scene, "rendering/resources/fonts", "lucida_console_regular_65.xml")
	if err != nil {
		logger.Errorf("Failed to load font: %s", err)
	}

	txt := shapes.NewText(a.ctx, lucida65, "◘1+,4°")
	txt.SetUniformScale(0.001)
	txt.MoveXY(0.4, -0.6)
	a.scene.Attach(txt)

	arial, err := rendering.LoadFont(a.scene, "rendering/resources/fonts", "arial_regular_65.xml")
	if err != nil {
		logger.Errorf("Failed to load font: %s", err)
	}
	txt = shapes.NewText(a.ctx, arial, "VAi\nTXQ")
	txt.SetUniformScale(0.001)
	txt.MoveXY(0.4, -0.75)
	a.scene.Attach(txt)

	//defer f.Destroy(a.scene)

	go func() {
		ctx := context.Background()
		if err := a.scene.StartHotReloading(ctx); err != nil {
			logger.Errorf("Failed to start shader hot-reloading: %s", err)
		}
	}()

	colShape := shapes.NewStaticShape(a.ctx)
	a.scene.Attach(colShape)

	texShape1 := shapes.NewTexturedShape(a.ctx)
	texShape1.SetTexture("sheep")
	a.scene.Attach(texShape1)

	texShape2 := shapes.NewTexturedShape(a.ctx)
	texShape2.SetPositionXY(-0.5, -0.5)
	texShape2.SetTexture("comp1")
	a.scene.Attach(texShape2)

	texShape3 := shapes.NewTexturedShape(a.ctx)
	texShape3.SetPositionXY(-1.5, -0.5)
	texShape3.SetTexture("comp2")
	a.scene.Attach(texShape3)

	testCube := shapes.NewCube(a.ctx, 0.2)
	testCube.MoveXY(-0.5, 0.4)
	a.scene.Attach(testCube)

	texCube := shapes.NewTexturedCube(a.ctx, 0.2)
	texCube.MoveXY(-1.0, 0.4)
	texCube.SetTexture("sheep")
	a.scene.Attach(texCube)

	sum := 0.0
	a.scene.AddUpdateJob(func(d time.Duration) {
		testCube.RotateZ(mgl32.DegToRad(1))
		testCube.RotateX(mgl32.DegToRad(1))
		testCube.RotateY(mgl32.DegToRad(1))

		texCube.RotateZ(mgl32.DegToRad(-1))
		texCube.RotateX(mgl32.DegToRad(-1))
		texCube.RotateY(mgl32.DegToRad(-1))

		sum += d.Seconds()
		//testCube.SetPositionXY(float32(math.Sin(sum)/2), 0)
	})

	lucida10, err := rendering.LoadFont(a.scene, "rendering/resources/fonts", "lucida_console_regular_10.xml")
	if err != nil {
		logger.Errorf("Failed to load font: %s", err)
	}

	for i := 0; i < 50; i++ {
		txt = shapes.NewText(a.ctx, lucida10, strconv.Itoa(i))
		txt.SetUniformScale(0.0018)
		txt.MoveXY(-1, 1-2*float32(i+1)/50)
		a.scene.Attach(txt)
		if i%5 == 0 {
			txt.SetColor(color.Red)
		}
	}

	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			a.scene.Detach(colShape)
			time.Sleep(300 * time.Millisecond)
			a.scene.Attach(colShape)
		}
	}()

}

func (a *Application) Stop() {
	//a.glCtx.DeleteProgram(program)
	//a.glCtx.DeleteBuffer(buf)
}

func (a *Application) Resize(width, height int) {
	a.width = width
	a.height = height
}

func (a *Application) Paint() { //sz size.Event) {
	a.camera.SetTrueOrthoWidth(5)
	a.engine.RenderSingleFrame(a.camera)

	//fps.Draw(sz)
}

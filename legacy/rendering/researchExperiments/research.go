package main

/*
  OpenGL support:
	"github.com/go-gl/gl/"
		Contains libraries for all OpenGL versions and extensions.
		Auto-generated bindings for the C-API and therefore suboptimal.
		Lacks support for other targets (web, mobile).

		Working example code for rendering a cube: https://raw.githubusercontent.com/go-gl/example/master/gl41core-cube/cube.go

	"golang.org/x/mobile/gl"
		Higher-Level API than github.com/go-gl/gl/ (but still very close to C --> that's what I want)
		All API-calls are pushed into a channel, which are then executed by the render thread.
		Supports multithreading / go-routines.
		Currently targeted for mobile devices, but is planned to also support desktop and web (with very slow progress).
		Only supports OpenGL ES 2.0 and OpenGL ES 3.0 - which is required if mobile/web should be supported.
		Official but experimental go repository.

		Previously, this library used "github.com/go-gl/gl/" under the hood.
		At that time, "https://github.com/goxjs/gl" was forked.
		Afterwards, they implemented the OpenGL ES backend themselves and added the render-thread approach.

		Desktop is supported via ANGLE (because OpenGL ES support is done by gpu vendors and not very stable/trustworthy).
			Details: https://github.com/golang/go/issues/9306
		ANGLE originates from Google and is used by Chrome/Chromium.
		Therefore, ANGLE DLLs are required, which are loaded at "golang.org/x/mobile/gl/dll_windows.go"
			This file searches in various locations for the DLLs (including potential chrome installation folders) and,
			if the DLLs are not found, downloads them into AppData/Local/GoGL.
			This code looks kinda strange and didn't work immediately on my machine. However, I managed to get the DLLs anyways.

	"https://github.com/goxjs/gl"
		A fork of x/mobile/gl with additional support for Windows and Web.
		Supports Desktop, Mobile and Web. Not actively maintained.
		It seems that this library was forked before "golang.org/x/mobile/gl" got the render-thread/worker using channels.
		Obsolete, since "golang.org/x/mobile/gl" seems to support Windows already - but not Web.
		Web is only supported via gopherJS, not web assembly - but there is an (open and incomplete) pull request for that.


  Windowing system:
	"github.com/go-gl/glfw/"
		For Desktop.
		Contains libraries for all OpenGL versions and extensions.
		Auto-generated bindings for the C-API.
	"golang.org/x/mobile/app"
		For mobile devices. Also supports desktop for testing purposes (very poor and buggy support).
	"github.com/goxjs/glfw"
		GLFW-like interface with support for desktop (glfw) and web.
		Uses "github.com/go-gl/glfw/" behind the scenes and should be easily extensible if needed.
	"golang.org/x/exp/shiny/screen"
		Cross-plattform library for acquiring windows.
		Uses "golang.org/x/mobile/gl" for the OpenGL driver and is also referenced by "golang.org/x/mobile/gl" for creating a context.
		Used by "github.com/golang/mobile/app". Example usage: https://github.com/golang/mobile/blob/master/app/shiny.go
		Not needed if "golang.org/x/mobile/app" is used (cross-platform & higher level).
	"https://github.com/hajimehoshi/ebiten": Use syscalls
		It's always possible to load the DLLs manually and perform syscalls: "https://github.com/hajimehoshi/ebiten"
		ebiten also has a custom version similar to go-gl (also generated with glow), but without cgo: https://github.com/hajimehoshi/ebiten/tree/master/internal/graphicsdriver/opengl/gl
		I don't want that.


  Math:
	azul3d:
		https://github.com/azul3d/engine/tree/master/lmath
		Looks nice and complete.
		Not a stand-alone library.
	github.com/go-gl/mathgl
		"godoc.org/github.com/go-gl/mathgl/mgl32"
		Also has "godoc.org/github.com/go-gl/mathgl/mgl32/matstack".
		Part of the code is auto-generated.
		==> this one!


  Fonts:
	"github.com/nullboundary/glfont"
		Uses go-gl and needs to be ported
	"github.com/4ydx/gltext"
		(research pending)
		...uses freetype:
	"github.com/golang/freetype"
		very low-level, and therefore not usable
		(research pending)
	Converting fonts into textures using FontBuilder, and handling rendering manually
		--> easiest solution, compatible with everything
		Bitmap-based (not freetype/truetype) --> won't look as crisp as it could (no hinting/kerning/...)

  Textures:
	x/mobile: https://github.com/golang/mobile/blob/master/exp/sprite/portable/portable.go
		User can provide an image.Image type (loaded via std-lib from png, jpg, gif, ...)
		--> image is then drawn (="uploaded") into another image.Image with the image.RGBA implementation.
		--> this image.RGBA is then pushed to the GPU (?)
	  ... super confusing. Somehow textures and scene-graphs/hierarchies are mixed? Textures are standalone-nodes with
		a specific shader? Sprites are supported...
	=> Supporting / being compatible with the image.Image interface is probably a good idea (image-decoder for files).

	go-gl: https://raw.githubusercontent.com/go-gl/example/master/gl41core-cube/cube.go
		The example also copies a loaded image into an RGBA image.
	=> I don't want "RGBA only". What about height maps and other stuff? They need different formats.
		We have RGBA, NRGBA (pre-multiplied alpha), Grey, Grey16, Alpha and Alpha16
	=> OpenGL ES 2.0 / WebGL 1.0 supports the following formats:
		OpenGL format		OpenGL type				Channels  bytes á texel  go-type
		gl.RGB				UNSIGNED_BYTE 				3			3
		gl.RGB				UNSIGNED_SHORT_5_6_5		3			2
		gl.RGBA				UNSIGNED_BYTE				4			4 		 image.RGBA
		gl.RGBA				UNSIGNED_SHORT_4_4_4_4		4			2
		gl.RGBA				UNSIGNED_SHORT_5_5_5_1		4			2
		gl.LUMINANCE		UNSIGNED_BYTE				1			1 		 image.Gray
		gl.LUMINANCE_ALPHA  UNSIGNED_BYTE				2			2
		gl.ALPHA 			UNSIGNED_BYTE				1			1  		 image.Alpha
	Other formats/types can be implemented manually by implementing image.Image and a color model.
	That's not too much work. https://golang.org/pkg/image/
	PROBLEM:
		... there is no nice way to get the underlying image-data/buffer from an "image.Image".
		The interface doesn't provide it! All implementations use an exported "Pix []uint8".
	==> Possible solution:
		(1.) NewTexture(size, format, type, []uint8) for raw textures with any format ==> we can support anything
		(2.) NewRGBATexture(image.RGBA), NewLuminanceTexture(image.Gray), NewAlphaTexture(image.Alpha), ...
		Problem: what if the user only has an "image.Image"?
			==> we could provide:
				NewRGBATexture(image.Image), NewLuminanceTexture(image.Image), NewAlphaTexture(image.Image),...
			...and perform a copy-operation into a solid image.XXX type.
			...this allows us to support any image type (even paletted types or custom implementations).
	PROBLEM: With this flexibility, how to return pixel data?
			convert each any everything into image.RGBA?
			==> Don't return it. Don't store it in RAM.
			If the user wants to have pixel-data, he needs to store image.Image himself.
			(And lose the possibility of hot-reloading?)
	PROBLEM: go-std-libs only provide png, jpg, gif, .... - but no DXT-formats.
			Also, there is no RGB format.

	Azul3D:
			... seems to support what I need. how???

	LOD


	==> WHAT DOES AZUL3D DO????
	... I want to keep it simple. Otherwise, I could have used Azul3D in the first place.





  Other:
	"https://github.com/google/gxui"
		A cross-platform GUI library (unmaintained).
		They face similar problems/decisions: https://github.com/google/gxui/issues/49
		For desktop, they started with "github.com/go-gl/glfw/", but switched to goxjs/glfw due its web-support.
		They started with "github.com/go-gl/gl/" (around since the beginning), but considered switching to "golang.org/x/mobile/gl".

	Example for "goxjs/gl" "goxjs/glfw":
		https://github.com/goxjs/example/blob/master/motionblur/shaders.go
		Motion-blur example that works flawlessly (tested on windows).

	azul3d:
		Rendering engine. Uses auto-generated OpenGL 2.0 / OpenGL ES 2.0 wrapper under the hood (GLOW).
		For desktop only (no web/mobile).
		Contains useful ideas for engine interfaces: https://github.com/azul3d/engine/blob/master/gfx/

    github.com/shurcooL/eX0/tree/master/eX0-go
		A game that uses goxjs/gl.
		Runs on Desktop and Web (gopherJS), but I couldn't get the web-part running.
	"https://github.com/go-qml/qml"
		QML also offers an OpenGL context.
		Unmaintained for several years.


  Combinations:
	"github.com/goxjs/glfw" + "https://github.com/goxjs/gl"
		OK
		Requires manual work if I want a context-object and (later) a dedicated render-thread like "golang.org/x/mobile/gl" does.
		I created an issue/question: https://github.com/goxjs/gl/issues/28
	"github.com/goxjs/glfw" + "golang.org/x/mobile/gl"
		Probably possible (But not too trivial)
		No Web support.
		Using this combination doesn't make too much sense -> instead of screen, I can use:
	"golang.org/x/exp/shiny/screen" + "golang.org/x/mobile/gl"
		OK, but panics when closing the window?
		No Web support.
		Using this combination doesn't make too much sense -> instead of screen, I can use:
	"golang.org/x/mobile/app" + "golang.org/x/mobile/gl"
		Works. But also no Web support.
	"github.com/go-gl/glfw/" + "github.com/go-gl/gl/"
		OK. Low-level and feature complete. I like it.
		Desktop only - no WebGL support.


  What I want:
	- Desktop & Web
		==> OpenGL ES 2.0 / 3.0 / WebGL
	- Context = interface, ideally with render-thread
	- Direct access to glCtx

  I will be using "golang.org/x/mobile/app" + "golang.org/x/mobile/gl".
  I can add Web-Support myself if needed.
		+ Actively maintained
		+ OpenGL interfaces and wrapper
		+ Render-Thread/Worker
		+ Easy to get it running on desktop
		- No WebGL support (I need to add it myself - possible, but a lot of (interesting) work)
		- Requires ANGLE-DLLs. This is not a problem, only the way they are retrieved/downloaded by the library is. I will ignore that for now.
*/

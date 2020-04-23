package nora

import (
	"github.com/maja42/glfw"
	"github.com/maja42/vmath"
)

// Settings control the application window.
type Settings struct {
	WindowTitle  string       // Initial window title
	WindowSize   vmath.Vec2i  // Initial window size
	ResizePolicy ResizePolicy // Behaviour if the window is resized

	Monitor *glfw.Monitor // Monitor on which the window should appear; nil: no preference

	Samples int // MSAA samples; 0: Disable multisampling
}

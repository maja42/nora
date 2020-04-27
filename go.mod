module github.com/maja42/nora

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.7
	github.com/maja42/gl v0.0.0-20200425200650-ab435bab8352
	github.com/maja42/glfw v0.0.0-20200425201231-b4f1c2b6f895
	github.com/maja42/vmath v0.0.0-20200427154541-8d10eba29da1
	github.com/sirupsen/logrus v1.5.0
	go.uber.org/atomic v1.5.1
)

// replace github.com/maja42/gl => ../gl
// replace github.com/maja42/glfw => ../glfw
// replace github.com/maja42/vmath => ../vmath

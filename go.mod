module github.com/maja42/nora

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20191125211704-12ad95a8df72
	github.com/maja42/gl v0.0.0-20200104193129-d6452f6faa31
	github.com/maja42/glfw v0.0.0-20200103101146-e269b0c3fdcb
	github.com/maja42/vmath v0.0.0-20200417115057-e683f27cb622
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.5.1 // indirect
	go.uber.org/atomic v1.5.1
)

replace github.com/maja42/vmath => ../vmath

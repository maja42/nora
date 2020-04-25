
# nora

Nora is a simple rendering engine for 2D and 3D applications, using OpenGL and GLFW.
It is targeted for desktop PCs and browsers (WebGL).

Nora is not intended to be a fully-fledged game engine.

It is designed to be a minimalistic framework/library for quickly getting something to the screen,
while maintaining good performance and low-level access to OpenGL functionality.

Nore is still under development and tries to mainly cover my own use-cases (at least for now).  

**Licence:** [GNU GPLv3](https://choosealicense.com/licenses/gpl-3.0/) (for now), if  not stated otherwise.

## Concurrency

Nora uses [github.com/maja42/gl](https://www.github.com/maja42/gl) to direct all OpenGL calls into a dedicated render thread.
This means that nora supports concurrency and can be used from multiple go-routines simultaneously.

Note, however, that OpenGL calls are serialized and forwarded to the render thread to be executed asynchronously.
The render-thread is only synchronized with the render-loop at specific points (end of frame, or when waiting for data from OpenGL calls).

Other engine objects (eg. meshes) are not concurrency-safe due to performance reasons. If they are accessed concurrently, 
synchronization primitives need to be used accordingly.

**Exception 0xc0000005:** If the application crashes or exits with this error code, it is very likely that engine objects 
(and their underlying OpenGL objects) were accessed concurrently. \
It often comes with the error message *"signal arrived during external code execution"*, which indicates that the underlying cgo code
caused a segmentation fault due to invalid ordering of OpenGL calls. It can be fixed by proper synchronization.

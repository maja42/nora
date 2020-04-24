package nora

import (
	"sync"

	"github.com/maja42/glfw"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
	"go.uber.org/atomic"
)

// CallbackID uniquely represents an interactive component.
type CallbackID struct {
	uint64
}

type OnMouseMoveEventFunc func(windowPos, movement vmath.Vec2i)
type OnMouseButtonEventFunc func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
type OnScrollEventFunc func(offset vmath.Vec2i)
type OnKeyEventFunc func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)

// InteractionSystem handles user interactions.
// 	- tracks the current mouse position
//	- tracks pressed buttons
//	- asynchronously informs components about an interaction
// User interactions are polled once every frame. Callbacks can therefore not race with the render function.
type InteractionSystem struct {
	m     sync.Mutex
	idSeq atomic.Uint64
	// listeners:
	mouseMoveEventFuncs   map[CallbackID]OnMouseMoveEventFunc
	mouseButtonEventFuncs map[CallbackID]OnMouseButtonEventFunc
	scrollEventFuncs      map[CallbackID]OnScrollEventFunc
	keyEventFuncs         map[CallbackID]OnKeyEventFunc
	// state:
	windowSize          vmath.Vec2i
	cursorPos           vmath.Vec2i // window coordinates
	pressedMouseButtons map[glfw.MouseButton]struct{}
	pressedKeys         map[glfw.Key]struct{}
}

// newInteractionSystem returns a new, empty interaction manager.
func newInteractionSystem(windowSize vmath.Vec2i, cursorPos vmath.Vec2i) InteractionSystem {
	return InteractionSystem{
		mouseMoveEventFuncs:   make(map[CallbackID]OnMouseMoveEventFunc),
		mouseButtonEventFuncs: make(map[CallbackID]OnMouseButtonEventFunc),
		scrollEventFuncs:      make(map[CallbackID]OnScrollEventFunc),
		keyEventFuncs:         make(map[CallbackID]OnKeyEventFunc),

		windowSize:          windowSize,
		cursorPos:           cursorPos,
		pressedMouseButtons: make(map[glfw.MouseButton]struct{}, 3),
		pressedKeys:         make(map[glfw.Key]struct{}, 5),
	}
}

// OnMouseMoveEvent adds a callback function to be executed on  mouse movements.
func (i *InteractionSystem) OnMouseMoveEvent(fn OnMouseMoveEventFunc) CallbackID {
	id := CallbackID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.mouseMoveEventFuncs[id] = fn
	return id
}

// RemoveMouseMoveEventFunc removes a previously added callback function for mouse movements.
func (i *InteractionSystem) RemoveMouseMoveEventFunc(cbID CallbackID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.mouseMoveEventFuncs, cbID)
}

// OnMouseButtonEvent adds a callback function to be executed on mouse button events.
func (i *InteractionSystem) OnMouseButtonEvent(fn OnMouseButtonEventFunc) CallbackID {
	id := CallbackID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.mouseButtonEventFuncs[id] = fn
	return id
}

// OnMouseButton adds a callback function to be executed if a specific mouse button performs a given action.
func (i *InteractionSystem) OnMouseButton(button glfw.MouseButton, action glfw.Action, fn func(glfw.ModifierKey)) CallbackID {
	return i.OnMouseButtonEvent(func(b glfw.MouseButton, a glfw.Action, mods glfw.ModifierKey) {
		if b != button || a != action {
			return
		}
		fn(mods)
	})
}

// RemoveMouseButtonEventFunc removes a previously added callback function for mouse button events.
func (i *InteractionSystem) RemoveMouseButtonEventFunc(cbID CallbackID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.mouseButtonEventFuncs, cbID)
}

// OnScroll adds a callback function to be executed on mouse scrolling events.
func (i *InteractionSystem) OnScroll(fn OnScrollEventFunc) CallbackID {
	id := CallbackID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.scrollEventFuncs[id] = fn
	return id
}

// RemoveScrollEventFunc removes a previously added callback function for mouse scrolling events.
func (i *InteractionSystem) RemoveScrollEventFunc(cbID CallbackID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.scrollEventFuncs, cbID)
}

// OnKeyEvent adds a callback function to be executed on key events.
func (i *InteractionSystem) OnKeyEvent(fn OnKeyEventFunc) CallbackID {
	id := CallbackID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.keyEventFuncs[id] = fn
	return id
}

// OnKey adds a callback function to be executed if a specific key performs a given action.
func (s *InteractionSystem) OnKey(key glfw.Key, action glfw.Action, fn func(glfw.ModifierKey)) CallbackID {
	return s.OnKeyEvent(func(k glfw.Key, _ int, a glfw.Action, mods glfw.ModifierKey) {
		if k != key || a != action {
			return
		}
		fn(mods)
	})
}

// RemoveKeyEventFunc removes a previously added callback function for key events.
func (i *InteractionSystem) RemoveKeyEventFunc(cbID CallbackID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.keyEventFuncs, cbID)
}

// RemoveAll removes all callback functions
func (i *InteractionSystem) RemoveAll() {
	i.m.Lock()
	defer i.m.Unlock()

	i.keyEventFuncs = make(map[CallbackID]OnKeyEventFunc)
	i.mouseMoveEventFuncs = make(map[CallbackID]OnMouseMoveEventFunc)
	i.mouseButtonEventFuncs = make(map[CallbackID]OnMouseButtonEventFunc)
}

// WindowSize returns the current size of the opened window
func (i *InteractionSystem) WindowSize() vmath.Vec2i {
	return i.windowSize
}

// WindowSpaceToClipSpace converts 2D window space [0, windowSize] into 2D clip space [-1,+1] coordinates.
func (i *InteractionSystem) WindowSpaceToClipSpace(windowSpace vmath.Vec2f) vmath.Vec2f {
	return vmath.Vec2f{
		2*windowSpace[0]/float32(i.windowSize[0]) - 1,
		-2*windowSpace[1]/float32(i.windowSize[1]) + 1,
	}
}

// ClipSpaceToWindowSpace converts 2D clip space [-1,+1] into 2D window space [0, windowSize] coordinates.
func (i *InteractionSystem) ClipSpaceToWindowSpace(clipSpace vmath.Vec2f) vmath.Vec2f {
	return vmath.Vec2f{
		(clipSpace[0] + 1) * float32(i.windowSize[0]) / 2,
		(clipSpace[0] - 1) * float32(i.windowSize[0]) / -2,
	}
}

// WindowSpaceDistToClipSpaceDist converts a 2D window space distance into a clip space distance.
// The calculation is independent of the clip space's origin (center of screen).
func (i *InteractionSystem) WindowSpaceDistToClipSpaceDist(windowSpaceDist vmath.Vec2f) vmath.Vec2f {
	return vmath.Vec2f{
		windowSpaceDist[0] * 2 / float32(i.windowSize[0]),
		windowSpaceDist[1] * -2 / float32(i.windowSize[1]),
	}
}

// ClipSpaceDistToWindowSpaceDist converts a 2D clip space distance into a window space distance.
// The calculation is independent of the clip space's origin (center of screen).
func (i *InteractionSystem) ClipSpaceDistToWindowSpaceDist(clipSpaceDist vmath.Vec2f) vmath.Vec2f {
	return vmath.Vec2f{
		clipSpaceDist[0] / 2 * float32(i.windowSize[0]),
		clipSpaceDist[1] / -2 * float32(i.windowSize[1]),
	}
}

// MousePosWindowSpace returns the current cursor position in window coordinates.
// (0,0) = top left corner
func (i *InteractionSystem) MousePosWindowSpace() vmath.Vec2i {
	return i.cursorPos
}

// MousePosClipSpace returns the current cursor position in clip space.
// (0,0) = top left corner; (1,1) = bottom right corner
func (i *InteractionSystem) MousePosClipSpace() vmath.Vec2f {
	return i.WindowSpaceToClipSpace(i.cursorPos.Vec2f())
}

// Returns true if the given mouse button is currently pressed.
// Call with glfw.MouseButtonLeft, glfw.MouseButtonRight, glfw.MouseButtonMiddle, ...
func (i *InteractionSystem) IsMouseButtonPressed(button glfw.MouseButton) bool {
	_, ok := i.pressedMouseButtons[button]
	return ok
}

// Returns true if the given key is currently pressed.
func (i *InteractionSystem) IsKeyPressed(key glfw.Key) bool {
	_, ok := i.pressedKeys[key]
	return ok
}

func (i *InteractionSystem) updateWindowSize(size vmath.Vec2i) {
	i.windowSize = size
}

// Executes the 'onMouseMove' callback of every interactive component
func (i *InteractionSystem) cursorPosCallback(_ *glfw.Window, x, y float64) {
	// No locking needed (iteration is safe)
	// If components are added/removed from within callbacks, it's unspecified if they receive the event that triggered the removal.
	// TODO: run in parallel

	newPos := vmath.Vec2i{int(x), int(y)}
	movement := newPos.Sub(i.cursorPos)
	i.cursorPos = newPos

	for _, fn := range i.mouseMoveEventFuncs {
		fn(i.cursorPos, movement)
	}
}

// Executes the 'onMouseButton' callback of every interactive component
func (i *InteractionSystem) mouseButtonCallback(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		assert.False(i.IsMouseButtonPressed(button), "Mouse button state inconsistent with events")
		i.pressedMouseButtons[button] = struct{}{}
	case glfw.Release:
		assert.True(i.IsMouseButtonPressed(button), "Mouse button state inconsistent with events")
		delete(i.pressedMouseButtons, button)
	case glfw.Repeat:
		assert.True(i.IsMouseButtonPressed(button), "Mouse button state inconsistent with events")
	}

	// No locking needed (iteration is safe)
	// If components are added/removed from within callbacks, it's unspecified if they receive the event that triggered the removal.
	// TODO: run in parallel

	for _, fn := range i.mouseButtonEventFuncs {
		fn(button, action, mods)
	}
}

// Executes the 'onMouseButton' callback of every interactive component
func (i *InteractionSystem) scrollCallback(_ *glfw.Window, xOff float64, yOff float64) {
	// No locking needed (iteration is safe)
	// If components are added/removed from within callbacks, it's unspecified if they receive the event that triggered the removal.
	// TODO: run in parallel

	offset := vmath.Vec2i{int(xOff), int(yOff)}
	for _, fn := range i.scrollEventFuncs {
		fn(offset)
	}
}

// Executes the 'onKey' callback of every interactive component
func (i *InteractionSystem) keyCallback(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		assert.False(i.IsKeyPressed(key), "Key state inconsistent with events")
		i.pressedKeys[key] = struct{}{}
	case glfw.Release:
		assert.True(i.IsKeyPressed(key), "Key state inconsistent with events")
		delete(i.pressedKeys, key)
	case glfw.Repeat:
		assert.True(i.IsKeyPressed(key), "Key state inconsistent with events")
	}

	// No locking needed (iteration is safe)
	// If components are added/removed from within callbacks, it's unspecified if they receive the event that triggered the removal.
	// TODO: run in parallel

	for _, fn := range i.keyEventFuncs {
		fn(key, scancode, action, mods)
	}
}

package nora

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/glfw"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/math"
	"go.uber.org/atomic"
	"sync"
)

// InterID uniquely represents an interactive component.
type InterID struct {
	uint64
}

type OnMouseMoveEventFunc func(windowPos math.Vec2i, worldspace mgl32.Vec2)
type OnMouseButtonEventFunc func(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
type OnKeyEventFunc func(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)

// InteractionSystem handles user interactions.
// 	- tracks the current mouse position
//	- tracks pressed buttons
//	- asynchronously informs components about an interaction
type InteractionSystem struct {
	m     sync.Mutex
	idSeq atomic.Uint64
	// listeners:
	mouseMoveEventFuncs   map[InterID]OnMouseMoveEventFunc
	mouseButtonEventFuncs map[InterID]OnMouseButtonEventFunc
	keyEventFuncs         map[InterID]OnKeyEventFunc
	// state:
	cursorPos           math.Vec2i // window coordinates
	pressedMouseButtons map[glfw.MouseButton]struct{}
	pressedKeys         map[glfw.Key]struct{}
}

// newInteractionSystem returns a new, empty interaction manager.
func newInteractionSystem(currentCursorPos math.Vec2i) InteractionSystem {
	return InteractionSystem{
		mouseMoveEventFuncs:   make(map[InterID]OnMouseMoveEventFunc),
		mouseButtonEventFuncs: make(map[InterID]OnMouseButtonEventFunc),
		keyEventFuncs:         make(map[InterID]OnKeyEventFunc),

		cursorPos:           currentCursorPos,
		pressedMouseButtons: make(map[glfw.MouseButton]struct{}, 3),
		pressedKeys:         make(map[glfw.Key]struct{}, 5),
	}
}

// OnMouseMoveEvent adds a callback function to be executed on  mouse movements.
func (i *InteractionSystem) OnMouseMoveEvent(fn OnMouseMoveEventFunc) InterID {
	id := InterID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.mouseMoveEventFuncs[id] = fn
	return id
}

// RemoveMouseMoveEventFunc removes a previously added callback function for mouse movements.
func (i *InteractionSystem) RemoveMouseMoveEventFunc(interID InterID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.mouseMoveEventFuncs, interID)
}

// OnMouseButtonEvent adds a callback function to be executed on mouse button events.
func (i *InteractionSystem) OnMouseButtonEvent(fn OnMouseButtonEventFunc) InterID {
	id := InterID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.mouseButtonEventFuncs[id] = fn
	return id
}

// OnMouseButton adds a callback function to be executed if a specific mouse button performs a given action.
func (s *InteractionSystem) OnMouseButton(button glfw.MouseButton, action glfw.Action, fn func(glfw.ModifierKey)) InterID {
	return s.OnMouseButtonEvent(func(b glfw.MouseButton, a glfw.Action, mods glfw.ModifierKey) {
		if b != button || a != action {
			return
		}
		fn(mods)
	})
}

// RemoveMouseButtonEventFunc removes a previously added callback function for mouse button events.
func (i *InteractionSystem) RemoveMouseButtonEventFunc(interID InterID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.mouseButtonEventFuncs, interID)
}

// OnKeyEvent adds a callback function to be executed on key events.
func (i *InteractionSystem) OnKeyEvent(fn OnKeyEventFunc) InterID {
	id := InterID{i.idSeq.Inc()}
	i.m.Lock()
	defer i.m.Unlock()
	i.keyEventFuncs[id] = fn
	return id
}

// OnKey adds a callback function to be executed if a specific key performs a given action.
func (s *InteractionSystem) OnKey(key glfw.Key, action glfw.Action, fn func(glfw.ModifierKey)) InterID {
	return s.OnKeyEvent(func(k glfw.Key, _ int, a glfw.Action, mods glfw.ModifierKey) {
		if k != key || a != action {
			return
		}
		fn(mods)
	})
}

// RemoveKeyEventFunc removes a previously added callback function for key events.
func (i *InteractionSystem) RemoveKeyEventFunc(interID InterID) {
	i.m.Lock()
	defer i.m.Unlock()
	delete(i.keyEventFuncs, interID)
}

// RemoveAll removes all callback functions
func (i *InteractionSystem) RemoveAll() {
	i.m.Lock()
	defer i.m.Unlock()

	i.keyEventFuncs = make(map[InterID]OnKeyEventFunc)
	i.mouseMoveEventFuncs = make(map[InterID]OnMouseMoveEventFunc)
	i.mouseButtonEventFuncs = make(map[InterID]OnMouseButtonEventFunc)
}

// MousePos returns the current cursor position in window coordinates.
// (0,0) = top left corner
func (i *InteractionSystem) MousePos() math.Vec2i {
	return i.cursorPos
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

// Executes the 'onMouseMove' callback of every interactive component
func (i *InteractionSystem) cursorPosCallback(_ *glfw.Window, x, y float64) {
	// No locking needed (iteration is safe)
	// If components are added/removed from within callbacks, it's unspecified if they receive the event that triggered the removal.
	// TODO: run in parallel

	i.cursorPos = math.Vec2i{int(x), int(y)}
	for _, fn := range i.mouseMoveEventFuncs {
		fn(i.cursorPos, mgl32.Vec2{}) // TODO: worldPos?
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

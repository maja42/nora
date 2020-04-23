package nora

import (
	"time"
)

// Drawable object.
type Drawable interface {
	Draw(*RenderState)
}

// Updateable object.
type Updateable interface {
	Update(elapsed time.Duration)
}

// Destroyable object.
// Objects implementing this interface must be destroyed after their use to free GPU resources.
type Destroyable interface {
	Destroy()
}

// Model represents a drawable object in the world.
// 	- Responsible for applying transformations before drawing
// 	- Responsible for invoking draw() on all underlying meshes.
// Models can optionally implement Updateable.
type Model interface {
	Drawable
	Destroyable
}

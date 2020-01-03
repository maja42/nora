package main

//
//  Things to change in x/mobile/gl when switching to/forking goxjs:
//
//		- add mapping from constants to strings (debug-build only)
//		- window-stuff should be left to the user. Allow changing window title, size, cursor, ...
//		- interactions (mouse/kb/touch) should not be part of the gl-library
//				(I know that windowing/interactions are not part of mobile/gl, but the expectation to use shiny (which does all this) is pretty high.
//
//		- types: uniform/attrib types should be consistent to GetActiveAttrib / GetActiveUniform indices.
//		- buffer(Sub)Data sucks
//
//		- maybe don't restrict to OpenGL ES? But supporting OpenGL 4 and such is a lot of work, and useless when targeting the web.
//				At least don't make it impossible.
//		- configurable logger?
//		- attrib + uniform: Better not structs?
//
//		- error reporting via context. Should not be part of the gl-lib, but should be doable by users via embedding
//
//		- dedicated render-worker that is feeded via a channel:
//				YES! this is great, because gl-commands can be issued from any go-routine/thread. This makes things a lot easier.
//				BUT: it doesn't go far enough. Some OpenGL commands need to be performed in sequence, without being interrupted by another operation.
//					Eg.: binding + writing to buffers; rendering; ...
//					Even if the commands themselves can be run from any go-routine, they must not mix.
//					There needs to be some kind of Lock() and Unlock() for complex operations and it would be nice to just
//						submit a batch-call of commands that are executed one after another without requiring a lock.
// 						For most operations, a batch-call is enough.
//						For rendering, a lock is needed.
//						maybe allow submitting an arbitrary function? (I don't see any use-case right now, and it could proof difficult to implement)
//
//

package hotreload

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// Watcher monitors a set of files on the filesystem
type Watcher struct {
	m       sync.RWMutex
	targets map[string][]interface{} // multiple "watches" can be interested in the same path

	//watcherM sync.Mutex
	watcher *fsnotify.Watcher
}

// NewWatcher creates a new filesystem watcher
// It monitors changes of multiple files and executes a callback method if they are modified.
// Watched files can be added and removed at runtime
func NewWatcher() *Watcher {
	return &Watcher{
		targets: make(map[string][]interface{}),
	}
}

// Add a new path to watch
func (w *Watcher) Add(path string, key interface{}) {
	path = filepath.Clean(path)

	w.m.Lock()
	defer w.m.Unlock()
	w.targets[path] = append(w.targets[path], key)

	if w.watcher != nil && len(w.targets[path]) == 1 {
		w.addToWatch(path)
	}
}

// Remove a watched path
func (w *Watcher) Remove(path string, key interface{}) error {
	path = filepath.Clean(path)

	w.m.Lock()
	defer w.m.Unlock()

	usages, ok := w.targets[path]
	if !ok {
		return fmt.Errorf("not watched: %q", path)
	}
	for idx, k := range usages {
		if k == key {
			usages = append(usages[:idx], usages[idx+1:]...)
			w.targets[path] = usages

			if w.watcher != nil && len(usages) == 0 {
				w.removeFromWatch(path)
			}
			return nil
		}
	}
	return fmt.Errorf("not watched with key %q: %q", key, path)
}

// Watch filesystem changes until the context is closed.
// Each path can be associated with multiple keys. If the file changes, the onChange method will be called for each registered key.
func (w *Watcher) Watch(ctx context.Context, onChanged func(key interface{})) error {
	w.m.Lock()
	if w.watcher != nil {
		w.m.Unlock()
		return fmt.Errorf("already watching")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		w.m.Unlock()
		return fmt.Errorf("initialize filesystem watcher: %s", err)
	}
	defer watcher.Close()

	w.watcher = watcher

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					logrus.Debugf("File changed: %s", event)
					w.onFileChanged(event.Name, onChanged)
				}
			case err := <-watcher.Errors:
				logrus.Errorf("Filesystem watcher error: %s", err)
			}
		}
	}()

	for path := range w.targets {
		w.addToWatch(path)
	}
	w.m.Unlock()

	<-done
	return nil
}

func (w *Watcher) addToWatch(path string) {
	if err := w.watcher.Add(path); err != nil {
		logrus.Warnf("Failed to watch %q (ignoring)", path)
	}
}

func (w *Watcher) removeFromWatch(path string) {
	if err := w.watcher.Remove(path); err != nil {
		logrus.Warnf("Failed to remove watch %q (ignoring)", path)
	}
}

func (w *Watcher) onFileChanged(path string, onChanged func(key interface{})) {
	w.m.RLock()
	defer w.m.RUnlock()

	pathTargets, ok := w.targets[path]
	if !ok {
		logrus.Errorf("Unknown filesystem change: no mapping for %q", path)
		return
	}

	for _, t := range pathTargets {
		onChanged(t)
	}
}

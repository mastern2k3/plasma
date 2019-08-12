package plasma

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"

	"github.com/pkg/errors"

	u "github.com/mastern2k3/plasma/util"
)

func isDir(path string) (bool, error) {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// file does not exist, gracefully ignore
			u.Logger.WithError(err).Warnf("file change detected but file nonexistant")
			return false, nil
		} else {
			return false, err
		}
	} else {
		return stat.IsDir(), nil
	}
}

func StartWatching(ctx context.Context, path string, output chan string) error {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	debounced := debounce.New(time.Second)

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			if err = watcher.Add(path); err != nil {
				return err
			}
		}

		return nil

	}); err != nil {
		return err
	}

	if err = watcher.Add(path); err != nil {
		return err
	}

	changedFileSet := make(map[string]bool)

	for {

		select {
		case event, ok := <-watcher.Events:

			if !ok {
				return errors.New("watcher events channel closed")
			}

			u.Logger.Printf("event: %v", event)

			if event.Op&fsnotify.Write == fsnotify.Write {

				u.Logger.Printf("modified file: %s", event.Name)

				changedFileSet[event.Name] = true

				debounced(func() {
					u.Logger.Printf("modified file: %s, debounce", event.Name)
					// if err = c.LoadActorClassFile(ctx, cl); err != nil {
					// 	return nil
					// }
					output <- event.Name
				})

			} else if event.Op&fsnotify.Create == fsnotify.Create {

				u.Logger.Printf("added element: %s", event.Name)

				is, err := isDir(filepath.Join(path, event.Name))
				if err != nil {
					return err
				}

				if is {
					if err = watcher.Add(event.Name); err != nil {
						return err
					}
				} else {

					changedFileSet[event.Name] = true

					debounced(func() {
						u.Logger.Printf("added file: %s, debounce", event.Name)

						// if err = c.LoadActorClassFile(ctx, cl); err != nil {
						// 	return err
						// }
						output <- event.Name
					})
				}

			} else if event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {

				if err = watcher.Remove(event.Name); err != nil {
					return err
				}
			}

		case err, ok := <-watcher.Errors:

			if !ok {
				return errors.New("watcher error events channel closed")
			}

			u.Logger.WithError(err).Error("error from watcher error channel")

			return err

		case <-ctx.Done():
			return nil
		}
	}
}

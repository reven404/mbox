package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type FileEvent struct {
	Path      string
	Operation string
	Timestamp time.Time
}

type FileWatcher struct {
	watcher   *fsnotify.Watcher
	watchDir  string
	eventChan chan FileEvent
	logger    *logrus.Logger
}

func NewFileWatcher(watchDir string, logger *logrus.Logger) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		watcher:   watcher,
		watchDir:  watchDir,
		eventChan: make(chan FileEvent, 100),
		logger:    logger,
	}

	if err := fw.addWatchRecursively(watchDir); err != nil {
		return nil, err
	}

	go fw.watchLoop()

	return fw, nil
}

func (fw *FileWatcher) addWatchRecursively(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") && path != root {
				return filepath.SkipDir
			}
			fw.logger.Debugf("Adding watch for directory: %s", path)
			return fw.watcher.Add(path)
		}
		return nil
	})
}

func (fw *FileWatcher) watchLoop() {
	defer close(fw.eventChan)

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if fw.shouldIgnoreEvent(event) {
				continue
			}

			fw.logger.Debugf("File event: %s %s", event.Op, event.Name)

			fileEvent := FileEvent{
				Path:      event.Name,
				Operation: fw.getOperationType(event.Op),
				Timestamp: time.Now(),
			}

			select {
			case fw.eventChan <- fileEvent:
			default:
				fw.logger.Warn("Event channel is full, dropping event")
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					if err := fw.watcher.Add(event.Name); err != nil {
						fw.logger.Errorf("Failed to add watch for new directory %s: %v", event.Name, err)
					}
				}
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Errorf("File watcher error: %v", err)
		}
	}
}

func (fw *FileWatcher) shouldIgnoreEvent(event fsnotify.Event) bool {
	baseName := filepath.Base(event.Name)
	
	if strings.HasPrefix(baseName, ".") {
		return true
	}
	
	if strings.HasSuffix(baseName, "~") {
		return true
	}
	
	if strings.HasSuffix(baseName, ".tmp") {
		return true
	}

	return false
}

func (fw *FileWatcher) getOperationType(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create == fsnotify.Create:
		return "CREATE"
	case op&fsnotify.Write == fsnotify.Write:
		return "WRITE"
	case op&fsnotify.Remove == fsnotify.Remove:
		return "REMOVE"
	case op&fsnotify.Rename == fsnotify.Rename:
		return "RENAME"
	case op&fsnotify.Chmod == fsnotify.Chmod:
		return "CHMOD"
	default:
		return "UNKNOWN"
	}
}

func (fw *FileWatcher) Events() <-chan FileEvent {
	return fw.eventChan
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}
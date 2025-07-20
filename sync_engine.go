package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type SyncEngine struct {
	config      *Config
	watcher     *FileWatcher
	tusdClient  *TusdClient
	logger      *logrus.Logger
	syncQueue   chan string
	processing  map[string]bool
	mutex       sync.RWMutex
	stopChan    chan struct{}
	wg          sync.WaitGroup
}

func NewSyncEngine(config *Config, logger *logrus.Logger) (*SyncEngine, error) {
	watcher, err := NewFileWatcher(config.WatchDir, logger)
	if err != nil {
		return nil, err
	}

	tusdClient := NewTusdClient(
		config.TusdURL,
		config.ChunkSize,
		config.MaxRetries,
		time.Duration(config.RetryDelay)*time.Second,
		logger,
	)

	return &SyncEngine{
		config:     config,
		watcher:    watcher,
		tusdClient: tusdClient,
		logger:     logger,
		syncQueue:  make(chan string, 1000),
		processing: make(map[string]bool),
		stopChan:   make(chan struct{}),
	}, nil
}

func (se *SyncEngine) Start() error {
	se.logger.Info("Starting sync engine...")

	if err := os.MkdirAll(se.config.WatchDir, 0755); err != nil {
		return err
	}

	se.wg.Add(3)
	go se.eventProcessor()
	go se.syncProcessor()
	go se.initialSync()

	se.logger.Infof("Sync engine started, watching directory: %s", se.config.WatchDir)
	return nil
}

func (se *SyncEngine) Stop() {
	se.logger.Info("Stopping sync engine...")
	
	close(se.stopChan)
	se.wg.Wait()
	
	if se.watcher != nil {
		se.watcher.Close()
	}
	
	close(se.syncQueue)
	se.logger.Info("Sync engine stopped")
}

func (se *SyncEngine) eventProcessor() {
	defer se.wg.Done()
	
	for {
		select {
		case event, ok := <-se.watcher.Events():
			if !ok {
				return
			}
			se.handleFileEvent(event)
			
		case <-se.stopChan:
			return
		}
	}
}

func (se *SyncEngine) handleFileEvent(event FileEvent) {
	if se.shouldIgnoreFile(event.Path) {
		return
	}

	se.logger.Debugf("Processing file event: %s %s", event.Operation, event.Path)

	switch event.Operation {
	case "CREATE", "WRITE":
		if se.isRegularFile(event.Path) {
			se.queueForSync(event.Path)
		}
	case "REMOVE":
		se.logger.Infof("File removed: %s", event.Path)
	case "RENAME":
		if se.isRegularFile(event.Path) {
			se.queueForSync(event.Path)
		}
	}
}

func (se *SyncEngine) queueForSync(filePath string) {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	if se.processing[filePath] {
		se.logger.Debugf("File %s is already being processed, skipping", filePath)
		return
	}

	select {
	case se.syncQueue <- filePath:
		se.logger.Debugf("Queued file for sync: %s", filePath)
	default:
		se.logger.Warnf("Sync queue is full, dropping file: %s", filePath)
	}
}

func (se *SyncEngine) syncProcessor() {
	defer se.wg.Done()

	for {
		select {
		case filePath, ok := <-se.syncQueue:
			if !ok {
				return
			}
			se.processFile(filePath)
			
		case <-se.stopChan:
			return
		}
	}
}

func (se *SyncEngine) processFile(filePath string) {
	se.mutex.Lock()
	se.processing[filePath] = true
	se.mutex.Unlock()

	defer func() {
		se.mutex.Lock()
		delete(se.processing, filePath)
		se.mutex.Unlock()
	}()

	if !se.isRegularFile(filePath) {
		se.logger.Debugf("Skipping non-regular file: %s", filePath)
		return
	}

	time.Sleep(100 * time.Millisecond)

	if !se.fileExists(filePath) {
		se.logger.Debugf("File no longer exists, skipping: %s", filePath)
		return
	}

	se.logger.Infof("Syncing file: %s", filePath)
	
	if err := se.tusdClient.UploadFile(filePath); err != nil {
		se.logger.Errorf("Failed to sync file %s: %v", filePath, err)
		
		time.Sleep(time.Duration(se.config.RetryDelay) * time.Second)
		se.queueForSync(filePath)
		return
	}

	se.logger.Infof("Successfully synced file: %s", filePath)
}

func (se *SyncEngine) initialSync() {
	defer se.wg.Done()

	se.logger.Info("Starting initial sync...")
	
	err := filepath.Walk(se.config.WatchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if se.shouldIgnoreFile(path) {
			return nil
		}

		se.queueForSync(path)
		return nil
	})

	if err != nil {
		se.logger.Errorf("Error during initial sync: %v", err)
		return
	}

	se.logger.Info("Initial sync completed")
}

func (se *SyncEngine) shouldIgnoreFile(filePath string) bool {
	baseName := filepath.Base(filePath)
	
	if strings.HasPrefix(baseName, ".") {
		return true
	}
	
	if strings.HasSuffix(baseName, "~") {
		return true
	}
	
	if strings.HasSuffix(baseName, ".tmp") {
		return true
	}
	
	if strings.Contains(baseName, ".DS_Store") {
		return true
	}

	return false
}

func (se *SyncEngine) isRegularFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func (se *SyncEngine) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
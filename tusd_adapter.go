package main

import (
	"time"
	"mdriver/providers"
)

// TusdClientAdapter adapts the main TusdClient to providers.TusdClient interface
type TusdClientAdapter struct {
	client *TusdClient
}

// NewTusdClientAdapter creates an adapter
func NewTusdClientAdapter(client *TusdClient) providers.TusdClient {
	return &TusdClientAdapter{client: client}
}

// UploadFile implements providers.TusdClient
func (a *TusdClientAdapter) UploadFile(filePath string) error {
	return a.client.UploadFile(filePath)
}

// ListFiles implements providers.TusdClient  
func (a *TusdClientAdapter) ListFiles() ([]providers.TusdFileInfo, error) {
	// Since the original TusdClient doesn't have ListFiles,
	// simulate some files for demonstration
	
	demoFiles := []providers.TusdFileInfo{
		{
			ID:   "upload_123",
			Name: "remote-file-1.txt",
			Size: 1024,
			ModTime: time.Now().Add(-1 * time.Hour),
			Metadata: map[string]string{
				"filename": "remote-file-1.txt",
				"status":   "completed",
			},
		},
		{
			ID:   "upload_456", 
			Name: "remote-file-2.txt",
			Size: 2048,
			ModTime: time.Now().Add(-30 * time.Minute),
			Metadata: map[string]string{
				"filename": "remote-file-2.txt", 
				"status":   "completed",
			},
		},
		{
			ID:   "upload_789",
			Name: "remote-file-3.txt", 
			Size: 4096,
			ModTime: time.Now().Add(-15 * time.Minute),
			Metadata: map[string]string{
				"filename": "remote-file-3.txt",
				"status":   "uploading",
			},
		},
	}
	
	return demoFiles, nil
}
package main

import (
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
	// return empty list for now
	return []providers.TusdFileInfo{}, nil
}
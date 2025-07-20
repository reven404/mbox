package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type TusdClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	chunkSize  int64
	maxRetries int
	retryDelay time.Duration
}

type UploadInfo struct {
	ID       string
	Size     int64
	Offset   int64
	Metadata map[string]string
}

func NewTusdClient(baseURL string, chunkSize int64, maxRetries int, retryDelay time.Duration, logger *logrus.Logger) *TusdClient {
	return &TusdClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:     logger,
		chunkSize:  chunkSize,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

func (c *TusdClient) CreateUpload(filePath string) (*UploadInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	filename := filepath.Base(filePath)
	
	req, err := http.NewRequest("POST", c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", strconv.FormatInt(fileInfo.Size(), 10))
	req.Header.Set("Upload-Metadata", fmt.Sprintf("filename %s", encodeBase64(filename)))
	req.Header.Set("Content-Length", "0")

	var resp *http.Response
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		
		if attempt < c.maxRetries-1 {
			c.logger.Warnf("Create upload attempt %d failed: %v, retrying...", attempt+1, err)
			time.Sleep(c.retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create upload after %d attempts: %w", c.maxRetries, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("no location header in response")
	}

	uploadID := filepath.Base(location)
	
	return &UploadInfo{
		ID:       uploadID,
		Size:     fileInfo.Size(),
		Offset:   0,
		Metadata: map[string]string{"filename": filename},
	}, nil
}

func (c *TusdClient) UploadFile(filePath string) error {
	uploadInfo, err := c.CreateUpload(filePath)
	if err != nil {
		return fmt.Errorf("failed to create upload: %w", err)
	}

	c.logger.Infof("Created upload %s for file %s", uploadInfo.ID, filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	for uploadInfo.Offset < uploadInfo.Size {
		chunkSize := c.chunkSize
		remaining := uploadInfo.Size - uploadInfo.Offset
		if remaining < chunkSize {
			chunkSize = remaining
		}

		chunk := make([]byte, chunkSize)
		n, err := file.ReadAt(chunk, uploadInfo.Offset)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk: %w", err)
		}
		chunk = chunk[:n]

		if err := c.uploadChunk(uploadInfo, chunk); err != nil {
			return fmt.Errorf("failed to upload chunk: %w", err)
		}

		uploadInfo.Offset += int64(n)
		c.logger.Debugf("Uploaded chunk, progress: %d/%d bytes", uploadInfo.Offset, uploadInfo.Size)
	}

	c.logger.Infof("Successfully uploaded file %s", filePath)
	return nil
}

func (c *TusdClient) uploadChunk(uploadInfo *UploadInfo, chunk []byte) error {
	url := fmt.Sprintf("%s/%s", c.baseURL, uploadInfo.ID)
	
	var err error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		req, reqErr := http.NewRequest("PATCH", url, bytes.NewReader(chunk))
		if reqErr != nil {
			return fmt.Errorf("failed to create request: %w", reqErr)
		}

		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Upload-Offset", strconv.FormatInt(uploadInfo.Offset, 10))
		req.Header.Set("Content-Type", "application/offset+octet-stream")
		req.Header.Set("Content-Length", strconv.Itoa(len(chunk)))

		resp, err := c.httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			
			if resp.StatusCode == http.StatusNoContent {
				return nil
			}
			
			body, _ := io.ReadAll(resp.Body)
			err = fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
		}

		if attempt < c.maxRetries-1 {
			c.logger.Warnf("Upload chunk attempt %d failed: %v, retrying...", attempt+1, err)
			time.Sleep(c.retryDelay)
		}
	}

	return fmt.Errorf("failed to upload chunk after %d attempts: %w", c.maxRetries, err)
}

func (c *TusdClient) GetUploadInfo(uploadID string) (*UploadInfo, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, uploadID)
	
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	size, _ := strconv.ParseInt(resp.Header.Get("Upload-Length"), 10, 64)
	offset, _ := strconv.ParseInt(resp.Header.Get("Upload-Offset"), 10, 64)

	return &UploadInfo{
		ID:     uploadID,
		Size:   size,
		Offset: offset,
	}, nil
}

func encodeBase64(s string) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	
	padding := ""
	switch len(s) % 3 {
	case 1:
		padding = "=="
	case 2:
		padding = "="
	}
	
	var result strings.Builder
	for i := 0; i < len(s); i += 3 {
		chunk := []byte{0, 0, 0}
		for j := 0; j < 3 && i+j < len(s); j++ {
			chunk[j] = s[i+j]
		}
		
		encoded := (int(chunk[0]) << 16) | (int(chunk[1]) << 8) | int(chunk[2])
		
		for j := 3; j >= 0; j-- {
			if i*4/3+3-j < len(s)*4/3+len(padding) {
				result.WriteByte(base64Chars[(encoded>>(j*6))&0x3F])
			}
		}
	}
	
	result.WriteString(padding)
	return result.String()
}
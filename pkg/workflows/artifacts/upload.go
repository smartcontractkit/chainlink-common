package artifacts

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type Field struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Input for uploading artifacts to storage service using presigned URLs
type UploadInput struct {
	PresignedURL    string        `json:"presignedUrl"`
	PresignedFields []Field       `json:"presignedFields"`
	Filepath        string        `json:"-"`
	Timeout         time.Duration `json:"-"`
}

type ArtifactType string

const (
	ArtifactTypeBinary ArtifactType = "BINARY"
	ArtifactTypeConfig ArtifactType = "CONFIG"
)

// Read in an artifact file from a given filepath and calculate the content hash
type ArtifactUpload struct {
	Content     []byte
	ContentType ArtifactType
	ContentHash string
}

// Constructor for ArtifactUpload
func NewArtifactUpload(filepath string) (*ArtifactUpload, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	contentType := ArtifactTypeBinary
	if strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml") {
		contentType = ArtifactTypeConfig
	}
	return &ArtifactUpload{
		Content:     content,
		ContentType: contentType,
		ContentHash: CalculateContentHash(content),
	}, nil
}

// Calculate the content hash of the artifact to generate the presigned URL
// for the artifact in the storage service
func CalculateContentHash(content []byte) string {
	hash := md5.Sum(content)                                  //nolint:gosec
	contentHash := base64.StdEncoding.EncodeToString(hash[:]) // Convert to base64 string
	return contentHash
}

// Upload artifacts to storage service using presigned URLs
func (a *Artifacts) upload(uploadInput *UploadInput) error {
	artifactUpload, err := NewArtifactUpload(uploadInput.Filepath)
	if err != nil {
		return err
	}

	a.log.Debug("Uploading artifact",
		"filepath", uploadInput.Filepath,
		"content type", artifactUpload.ContentType,
		"content hash", artifactUpload.ContentHash)

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add the presigned form fields to the request (do not add extra fields).
	for _, field := range uploadInput.PresignedFields {
		if err := w.WriteField(field.Key, field.Value); err != nil {
			a.log.Error("Failed to write presigned field", "error", err, "field", field.Key)
			return err
		}
	}

	// Add the Content-Type header to the request.
	err = w.WriteField("Content-Type", string(artifactUpload.ContentType))
	if err != nil {
		return err
	}
	// Add the Content-MD5 header to the request.
	err = w.WriteField("Content-MD5", artifactUpload.ContentHash)
	if err != nil {
		return err
	}

	// Add the file to the request as the last field.
	fileWriter, err := w.CreateFormFile("file", "artifact")
	if err != nil {
		a.log.Error("Failed to create form file field", "error", err)
		return err
	}
	if _, err := fileWriter.Write(artifactUpload.Content); err != nil {
		a.log.Error("Failed to write file content to form", "error", err)
		return err
	}
	if err := w.Close(); err != nil {
		a.log.Error("Failed to close multipart writer", "error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), uploadInput.Timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", uploadInput.PresignedURL, &b)
	if err != nil {
		a.log.Error("Failed to create HTTP request", "error", err)
		return err
	}
	httpReq.Header.Set("Content-Type", w.FormDataContentType())

	httpClient := &http.Client{Timeout: uploadInput.Timeout + 2*time.Second}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		a.log.Error("HTTP request to origin failed", "error", err)
		return err
	}
	defer func() {
		if cerr := httpResp.Body.Close(); cerr != nil {
			a.log.Warn("Failed to close origin response body", "error", cerr)
		}
	}()

	// Accept 204 No Content or 201 Created as success.
	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(httpResp.Body)
		a.log.Error("Artifact upload failed", "status", httpResp.StatusCode, "body", string(body))
		return fmt.Errorf("expected status 204 or 201, got %d: %s", httpResp.StatusCode, string(body))
	}

	a.log.Info("Artifact uploaded successfully", "status", httpResp.StatusCode)
	return nil
}

// DurableUpload uploads an artifact with up to 3 attempts and exponential backoff.
func (a *Artifacts) DurableUpload(uploadInput *UploadInput) error {
	var lastErr error
	const maxUploadAttempts = 3
	for attempt := 0; attempt < maxUploadAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt)) * time.Second
			a.log.Debug("Retrying upload after backoff", "attempt", attempt+1, "backoff", backoff)
			time.Sleep(backoff)
		}
		lastErr = a.upload(uploadInput)
		if lastErr == nil {
			return nil
		}
		a.log.Warn("Upload attempt failed", "attempt", attempt+1, "error", lastErr)
	}
	return fmt.Errorf("upload failed after %d attempts: %w", maxUploadAttempts, lastErr)
}

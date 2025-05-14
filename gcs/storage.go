package gcs

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Service holds the GCS client and bucket name
type GCService struct {
	Client     *storage.Client
	BucketName string
}

// NewGCSService creates and initializes a new GCS service
func NewGCSService(ctx context.Context, credentialFile, projectID, bucketName string) (*GCService, error) {
	var client *storage.Client
	var err error

	if credentialFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialFile))
		log.Printf("Initializing GCS client with credentials file: %s", credentialFile)
	} else {
		// If no credentialFile is provided, NewClient will use Application Default Credentials.
		client, err = storage.NewClient(ctx)
		log.Println("Initializing GCS client with Application Default Credentials.")
	}

	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}

	log.Printf("Google Cloud Storage client initialized successfully for bucket: %s", bucketName)
	return &GCService{
		Client:     client,
		BucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to GCS
func (s *GCService) UploadFile(ctx context.Context, fileReader io.Reader, objectName, contentType string) (*storage.ObjectAttrs, error) {
	if s.Client == nil {
		return nil, fmt.Errorf("GCS client not initialized in service")
	}
	if s.BucketName == "" {
		return nil, fmt.Errorf("GCS bucket name not configured in service")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*120) // 120-second timeout for upload
	defer cancel()

	bucket := s.Client.Bucket(s.BucketName)
	obj := bucket.Object(objectName)
	writer := obj.NewWriter(ctx)

	if contentType != "" {
		writer.ContentType = contentType
	}
	// Example: writer.CacheControl = "public, max-age=31536000"

	if _, err := io.Copy(writer, fileReader); err != nil {
		_ = writer.Close() // Attempt to close writer even on error
		return nil, fmt.Errorf("io.Copy to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("GCS Writer.Close: %w", err)
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("GCS obj.Attrs: %w", err)
	}

	log.Printf("File %s uploaded to gs://%s/%s\n", objectName, attrs.Bucket, attrs.Name)
	return attrs, nil
}

// DeleteFile deletes an object from GCS
func (s *GCService) DeleteFile(ctx context.Context, objectName string) error {
	if s.Client == nil || s.BucketName == "" {
		return fmt.Errorf("GCS client or bucket not configured in service")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	obj := s.Client.Bucket(s.BucketName).Object(objectName)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			log.Printf("Attempted to delete non-existent GCS object: %s", objectName)
			return nil // Not an error in this context for cleanup
		}
		return fmt.Errorf("GCS Object(%q).Delete: %w", objectName, err)
	}
	log.Printf("Object %s deleted from GCS bucket %s.", objectName, s.BucketName)
	return nil
}

// Close closes the GCS client
func (s *GCService) Close() error {
	if s.Client != nil {
		log.Println("Closing GCS client.")
		return s.Client.Close()
	}
	return nil
}

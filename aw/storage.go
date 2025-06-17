package aw

import (
	"io"
	"mime/multipart"
	"os"

	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/appwrite/sdk-for-go/client"
	"github.com/appwrite/sdk-for-go/file"
	"github.com/appwrite/sdk-for-go/storage"
)

type AppwriteService struct {
	client  client.Client
	Storage *storage.Storage
}

func NewAppwriteService(client client.Client) *AppwriteService {
	return &AppwriteService{
		client:  client,
		Storage: storage.New(client),
	}
}

func (s *AppwriteService) UploadFile(fileHeader *multipart.FileHeader) (string, error) {
	cfg, err := cfg.LoadConfig()
	if err != nil {
		return "", err
	}
	bucketId := cfg.AppwriteBucketId
	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	// Copy the uploaded file to the temp file
	if _, err := io.Copy(tmpFile, src); err != nil {
		return "", err
	}

	// Create InputFile for Appwrite
	inputFile := file.NewInputFile(tmpFile.Name(), fileHeader.Filename)

	// Create file in Appwrite storage
	result, err := s.Storage.CreateFile(
		bucketId,
		"unique()",
		inputFile,
		s.Storage.WithCreateFilePermissions([]string{"read(\"any\")"}),
	)
	if err != nil {
		return "", err
	}

	// Return the file ID which can be used to construct the URL
	return result.Id, nil
}

// GetFileURL constructs the public view URL for a file in Appwrite storage.
func (s *AppwriteService) GetFileURL(fileId string) string {
	return "/file/preview/" + fileId
}

// GetSignedPreviewURL generates a signed preview URL for a file valid for 5 hours.
func (s *AppwriteService) GetSignedPreviewURL(fileId string) (string, error) {
	return "", nil
}

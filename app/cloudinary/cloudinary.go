package cloudinary

import (
	"bytes"
	"context"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &CloudinaryService{cld: cld}, nil
}

// Helper function to convert bool to *bool
func boolPtr(b bool) *bool {
	return &b
}

func (s *CloudinaryService) UploadImage(file *bytes.Buffer, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadResult, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       filename,
		ResourceType:   "auto",
		Folder:         "images",
		UseFilename:    boolPtr(true),
		UniqueFilename: boolPtr(true),
	})

	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

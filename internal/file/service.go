package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

type Service interface {
	UploadImage(ctx context.Context, fileBytes []byte, contentType string) (entity.File, error)
}

func NewService(localStoragePath string, logger log.Logger) Service {
	return service{localStoragePath, logger}
}

type service struct {
	localStoragePath string
	logger           log.Logger
}

func (s service) getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	default:
		return ""
	}
}

func (s service) saveImageToLocalStorage(_ context.Context, fileName string, fileSubject string, fileBytes []byte) (string, error) {
	absoluteDir := filepath.Join(s.localStoragePath, fileSubject)
	if _, err := os.Stat(absoluteDir); os.IsNotExist(err) {
		err := os.MkdirAll(absoluteDir, 0755)
		if err != nil {
			s.logger.Errorf("File save directory not created %v", err)
			return "", errors.InternalServerError("File save directory not created")
		}
	}

	absolutePath := filepath.Join(absoluteDir, fileName)
	file, err := os.Create(absolutePath)

	if err != nil {
		s.logger.Errorf("Error creating file in the local storage %v", err)
		return "", errors.InternalServerError("Error creating file in the local storage")
	}
	defer file.Close()

	n, err := file.Write(fileBytes)
	if err != nil {
		s.logger.Errorf("Error writing file in the local storage %v", err)
		return "", errors.InternalServerError("Error writing file in the local storage")
	}
	s.logger.Infof("✅ Total bytes written %d", n)
	return fmt.Sprintf("http://localhost:%d/files/image/%s", 8080, fileName), nil
}

// UploadImage implements Service.
func (s service) UploadImage(ctx context.Context, fileBytes []byte, contenType string) (entity.File, error) {
	// userID := auth.CurrentUser(ctx).GetID()
	fileID := uuid.New().String()
	fileExt := s.getExtensionFromContentType(contenType)
	fileSubject := "album"
	fileName := fmt.Sprintf("%s%s", fileID, fileExt)
	useLocalStorage := true

	var fileURL string
	var err error
	if useLocalStorage {
		fileURL, err = s.saveImageToLocalStorage(ctx, fileName, fileSubject, fileBytes)
		if err != nil {
			return entity.File{}, err
		}
	} else {
		// TODO(resultanyildizi): save image to cloud storage
	}

	// If successful:
	// - Yes: Write file metadata (userID, contentType, fileID, createdAt, deletedAt, album) to database
	//        If successful:
	//		  - Yes: Return file metadata to user
	//        - No: Return error
	// - No: Return error

	return entity.File{ID: fileID, URL: fileURL}, nil
}

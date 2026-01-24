package file

import (
	"context"

	"github.com/google/uuid"
	"github.com/qiangxue/go-rest-api/internal/auth"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

type FileStorage interface {
	WriteFile(_ context.Context, file entity.File, bytes []byte) (string, error)
	GetFileURL(_ context.Context, file entity.File) (string, error)
}

type Service interface {
	UploadImage(ctx context.Context, fileBytes []byte, fileSize int64, contentType string) (entity.File, error)
}

func NewService(repository Repository, fileStorage FileStorage, logger log.Logger) Service {
	return service{repository, fileStorage, logger}
}

type service struct {
	repository  Repository
	fileStorage FileStorage
	logger      log.Logger
}

// UploadImage implements Service.
func (s service) UploadImage(ctx context.Context, fileBytes []byte, fileSize int64, contenType string) (entity.File, error) {
	// userID := auth.CurrentUser(ctx).GetID()
	fileID := uuid.New().String()
	fileSubject := "album"
	userID := auth.CurrentUser(ctx).ID

	file := entity.File{
		ID:          fileID,
		ContentType: contenType,
		Subject:     fileSubject,
		UserID:      userID,
		Size:        int64(len(fileBytes)),
	}

	var fileURL string
	var err error
	fileURL, err = s.fileStorage.WriteFile(ctx, file, fileBytes)
	if err != nil {
		return entity.File{}, err
	}

	err = s.repository.CreateFile(ctx, file)
	if err != nil {
		s.logger.Errorf("Could not add file to database %v", err)
		return entity.File{}, errors.InternalServerError("Could not add file to database")
	}

	file.URL = fileURL
	return file, nil
}

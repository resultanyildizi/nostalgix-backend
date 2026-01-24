package file

import (
	"context"
	"errors"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/dbcontext"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

type Repository interface {
	CreateFile(ctx context.Context, file entity.File) error
}

func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

func (r repository) CreateFile(ctx context.Context, file entity.File) error {
	timeNow := time.Now()

	result, err := r.db.With(ctx).Insert("file", dbx.Params{
		"id":           file.ID,
		"user_id":      file.UserID,
		"size":         file.Size,
		"subject":      file.Subject,
		"content_type": file.ContentType,
		"created_at":   timeNow,
		"updated_at":   timeNow,
		"deleted_at":   nil,
	}).Execute()

	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if affectedRows <= 0 {
		return errors.New("No rows added")
	}

	return nil
}

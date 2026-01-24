package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

func NewLocalStorage(localStoragePath string, logger log.Logger) FileStorage {
	return localStorage{localStoragePath, logger}
}

type localStorage struct {
	localStoragePath string
	logger           log.Logger
}

// WriteFile implements FileStorage.
func (l localStorage) WriteFile(ctx context.Context, file entity.File, bytes []byte) (string, error) {
	absoluteDir := filepath.Join(l.localStoragePath, file.Subject)
	if _, err := os.Stat(absoluteDir); os.IsNotExist(err) {
		err := os.MkdirAll(absoluteDir, 0755)
		if err != nil {
			l.logger.Errorf("File save directory not created %v", err)
			return "", errors.InternalServerError("File save directory not created")
		}
	}

	absolutePath := filepath.Join(absoluteDir, file.GetName())
	osfile, err := os.Create(absolutePath)

	if err != nil {
		l.logger.Errorf("Error creating file in the local storage %v", err)
		return "", errors.InternalServerError("Error creating file in the local storage")
	}
	defer osfile.Close()

	n, err := osfile.Write(bytes)
	if err != nil {
		l.logger.Errorf("Error writing file in the local storage %v", err)
		return "", errors.InternalServerError("Error writing file in the local storage")
	}
	l.logger.Infof("✅ Total bytes written %d", n)

	return l.GetFileURL(ctx, file)

}

// GetFileURL implements FileStorage.
func (l localStorage) GetFileURL(_ context.Context, file entity.File) (string, error) {
	return fmt.Sprintf("http://localhost:%d/v1/files/image/%s/%s", 8080, file.Subject, file.GetName()), nil
}

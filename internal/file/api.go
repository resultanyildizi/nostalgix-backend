package file

import (
	"fmt"
	"io"
	"net/http"

	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

func RegisterHandlers(r *routing.RouteGroup, service Service, authHandler routing.Handler, logger log.Logger) {
	res := resource{service, logger}

	r.Use(authHandler)
	r.Post("/files/image", res.uploadImage)
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r resource) uploadImage(c *routing.Context) error {
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MiB =>  10 * 2 ^ 10 = 10 * 1024

	if err != nil {
		r.logger.Errorf("Error parsing the form data %v", err)
		return errors.BadRequest(fmt.Sprintf("Error parsing the form data %v", err))
	}

	file, header, err := c.Request.FormFile("image")

	if err != nil {
		r.logger.Errorf("Error reading form file %v", err)
		return errors.BadRequest(fmt.Sprintf("Error reading form file %v", err))
	}

	defer file.Close()

	if header.Size > (10 << 20) {
		return errors.BadRequest("Image file is too big. Maximum 5 MiB allowed.")
	}

	contentType := header.Header.Get("content-type")

	switch contentType {
	case "image/png":
	case "image/jpeg":
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			r.logger.Errorf("Error while reading file content %v", err)
			return errors.InternalServerError("Error while reading file content")
		}

		ctx := c.Request.Context()
		file, err := r.service.UploadImage(ctx, fileBytes, contentType)
		if err != nil {
			return err
		}
		return c.WriteWithStatus(file, http.StatusOK)
	default:
		r.logger.Errorf("Invalid image type, %s is not supported", contentType)
		return errors.BadRequest("Invalid image type. Not supported.")
	}

	return nil
}

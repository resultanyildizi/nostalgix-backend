package file

import (
	stdbytes "bytes"
	"context"
	stderrors "errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/errors"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

func NewCloudStorage(awsClient *s3.Client, bucketName, publicDomain string, logger log.Logger) FileStorage {
	return CloudStorage{awsClient, bucketName, publicDomain, logger}
}

type CloudStorage struct {
	awsClient    *s3.Client
	bucketName   string
	publicDomain string
	logger       log.Logger
}

// WriteFile implements FileStorage.
func (c CloudStorage) WriteFile(ctx context.Context, file entity.File, bytes []byte) (string, error) {
	absolutePath := filepath.Join(file.Subject, file.GetName())

	_, err := c.awsClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(absolutePath),
		Body:        stdbytes.NewReader(bytes),
		ContentType: aws.String(file.ContentType),
	})

	if err != nil {
		var apiErr smithy.APIError
		if stderrors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			c.logger.Errorf("Error while uploading object to %s. The object is too large.\n"+
				"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
				"or the multipart upload API (5TB max).", c.bucketName)
			return "", errors.InternalServerError("Error while uploading object. File is too large.")
		} else {
			c.logger.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				file.GetName(), c.bucketName, absolutePath, err)
			return "", errors.InternalServerError("Error while uploading the object.")
		}
	} else {
		err = s3.NewObjectExistsWaiter(c.awsClient).Wait(
			ctx,
			&s3.HeadObjectInput{Bucket: aws.String(c.bucketName), Key: aws.String(absolutePath)},
			time.Minute,
		)
		if err != nil {
			c.logger.Errorf("Failed attempt to wait for object %s to exist.\n", absolutePath)
			return "", errors.InternalServerError("Error while uploading the object.")
		}
	}

	return c.GetFileURL(ctx, file)
}

// GetFileURL implements FileStorage.
func (c CloudStorage) GetFileURL(_ context.Context, file entity.File) (string, error) {
	// return filepath.Join(c.publicDomain, file.Subject, file.GetName()), nil
	return fmt.Sprintf("%s/%s/%s", c.publicDomain, file.Subject, file.GetName()), nil
}

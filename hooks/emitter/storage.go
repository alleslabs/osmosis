package emitter

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

func UploadFile(uploader *manager.Uploader, bucket string, objectPath string, message []byte) error {
	reader := bytes.NewReader(message)
	_, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectPath),
		ContentType: aws.String("application/json"),
		Body:        reader,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	return nil
}

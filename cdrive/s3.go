package drive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

type S3Drive struct {
	*FSDrive
	*s3.S3
	ctx    context.Context
	config *S3Config
}

func initS3Drive(config S3Config) *S3Drive {
	return &S3Drive{S3: config.newClient(), ctx: context.Background(), config: &config}
}

func (c *S3Config) newClient() *s3.S3 {
	awsConfig := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Region:           aws.String(c.Region),
		S3ForcePathStyle: aws.Bool(true),
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		panic(err)
	}

	return s3.New(awsSession)
}

func (s *S3Drive) name() {

}

func (s *S3Drive) ListBlobs(prefix string) ([]string, error) {
	return s.ListBlobsUntil(prefix, nil)
}

func (s *S3Drive) ListBlobsUntil(prefix string, until *time.Time) ([]string, error) {
	response, err := s.ListObjectsWithContext(s.ctx, &s3.ListObjectsInput{
		Bucket:  aws.String(s.config.BucketName),
		MaxKeys: aws.Int64(1000),
		Prefix:  aws.String(prefix),
	})
	if err != nil {
		return []string{}, fmt.Errorf("failed to list objects from S3, %s", err.Error())
	}

	files := make([]string, 0)
	for _, item := range response.Contents {
		if until != nil && item.LastModified.After(*until) {
			continue
		}

		key := *item.Key

		// if it ends with '/' it's not an actual blob
		if strings.HasSuffix(key, "/") {
			continue
		}

		fullKey := prefix + key[len(prefix):]
		files = append(files, strings.TrimLeft(fullKey, "/"))
	}

	sort.Strings(files)
	return files, nil
}

func (s *S3Drive) GetBlobByKey(key string) (*GetBlob, error) {
	fullPath := s.s3KeyFullPath(key)
	resp, err := s.GetObjectWithContext(s.ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to get object '%s' from S3, %s", fullPath, err.Error())
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	file, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object '%s' from S3, %s", fullPath, err.Error())
	}

	return &GetBlob{
		Size:         *resp.ContentLength,
		Content:      file,
		ContentType:  *resp.ContentType,
		LastModified: *resp.LastModified,
	}, nil
}

func (s *S3Drive) SaveBlob(b *NewBlob) error {
	if err := ValidateKey(b.Key); err != nil {
		return fmt.Errorf("failed to validate object key '%s', %s", b.Key, err.Error())
	}

	fullPath := s.s3KeyFullPath(b.Key)
	reader := bytes.NewReader(b.Content)
	_, err := s.PutObjectWithContext(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(fullPath),
		ContentType: aws.String(b.ContentType),
		Body:        reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object '%s' to S3, %s", fullPath, err.Error())
	}
	return nil
}

func (s *S3Drive) SaveBlobWithResult(b *NewBlob) (*SaveBlobResult, error) {
	if err := ValidateKey(b.Key); err != nil {
		return nil, fmt.Errorf("failed to validate object key '%s', %s", b.Key, err.Error())
	}

	fullPath := s.s3KeyFullPath(b.Key)
	reader := bytes.NewReader(b.Content)
	_, err := s.PutObjectWithContext(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(fullPath),
		ContentType: aws.String(b.ContentType),
		Body:        reader,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload object '%s' to S3, %s", fullPath, err.Error())
	}
	return &SaveBlobResult{
		Path:        fullPath,
		URL:         s.cloudFrontURL(b.Key),
		Size:        int64(len(b.Content)),
		ContentType: b.ContentType,
		Drive:       string(S3DriveOption),
	}, nil
}

func (s *S3Drive) DeleteBlob(key string) error {
	fullPath := s.s3KeyFullPath(key)
	_, err := s.DeleteObjectWithContext(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to delete object '%s' on S3, %s", fullPath, err.Error())
	}
	return nil
}

func (s *S3Drive) s3KeyFullPath(key string) string {
	u, _ := url.Parse(s.config.CloudFrontBaseURL)
	return path.Join(strings.TrimLeft(u.Path, "/"), path.Join(key))
}

func isNotFound(err error) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == s3.ErrCodeNoSuchKey
	}
	return false
}

func (s *S3Drive) cloudFrontURL(key string) string {
	return fmt.Sprintf("%s/%s", s.config.CloudFrontBaseURL, key)
}

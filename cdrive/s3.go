package drive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
		Region:           aws.String(c.region()),
		S3ForcePathStyle: aws.Bool(true),
	}
	if c.Endpoint != "" {
		awsConfig.Endpoint = aws.String(c.Endpoint)
	}

	awsSession, err := session.NewSession(awsConfig)
	if err != nil {
		panic(err)
	}

	return s3.New(awsSession)
}

func (c *S3Config) region() string {
	if c.Region != "" {
		return c.Region
	}
	if c.provider() == string(R2DriveOption) {
		return "auto"
	}
	return c.Region
}

func (c *S3Config) provider() string {
	if c.Provider != "" {
		return strings.ToLower(c.Provider)
	}
	if c.Endpoint != "" && strings.Contains(c.Endpoint, "r2.cloudflarestorage.com") {
		return string(R2DriveOption)
	}
	return string(S3DriveOption)
}

func (c *S3Config) cloudOption() Option {
	if c.provider() == string(R2DriveOption) {
		return R2DriveOption
	}
	return S3DriveOption
}

func (s *S3Drive) name() {

}

func (s *S3Drive) ListBlobs(prefix string) ([]string, error) {
	return s.ListBlobsUntil(prefix, nil)
}

func (s *S3Drive) ListBlobsUntil(prefix string, until *time.Time) ([]string, error) {
	fullPrefix := s.objectKeyFullPath(prefix)
	response, err := s.ListObjectsWithContext(s.ctx, &s3.ListObjectsInput{
		Bucket:  aws.String(s.config.BucketName),
		MaxKeys: aws.Int64(1000),
		Prefix:  aws.String(fullPrefix),
	})
	if err != nil {
		return []string{}, fmt.Errorf("failed to list objects from %s, %s", s.config.provider(), err.Error())
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

		files = append(files, strings.TrimLeft(key, "/"))
	}

	sort.Strings(files)
	return files, nil
}

func (s *S3Drive) GetBlobByKey(key string) (*GetBlob, error) {
	fullPath := s.objectKeyFullPath(key)
	resp, err := s.GetObjectWithContext(s.ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to get object '%s' from %s, %s", fullPath, s.config.provider(), err.Error())
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	file, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object '%s' from %s, %s", fullPath, s.config.provider(), err.Error())
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

	fullPath := s.objectKeyFullPath(b.Key)
	reader := bytes.NewReader(b.Content)
	_, err := s.PutObjectWithContext(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(fullPath),
		ContentType: aws.String(b.ContentType),
		Body:        reader,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object '%s' to %s, %s", fullPath, s.config.provider(), err.Error())
	}
	return nil
}

func (s *S3Drive) SaveBlobWithResult(b *NewBlob) (*SaveBlobResult, error) {
	if err := ValidateKey(b.Key); err != nil {
		return nil, fmt.Errorf("failed to validate object key '%s', %s", b.Key, err.Error())
	}

	fullPath := s.objectKeyFullPath(b.Key)
	reader := bytes.NewReader(b.Content)
	_, err := s.PutObjectWithContext(s.ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.BucketName),
		Key:         aws.String(fullPath),
		ContentType: aws.String(b.ContentType),
		Body:        reader,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload object '%s' to %s, %s", fullPath, s.config.provider(), err.Error())
	}
	return &SaveBlobResult{
		Path:        fullPath,
		URL:         s.publicURL(fullPath),
		Size:        int64(len(b.Content)),
		ContentType: b.ContentType,
		Drive:       s.config.provider(),
	}, nil
}

func (s *S3Drive) DeleteBlob(key string) error {
	fullPath := s.objectKeyFullPath(key)
	_, err := s.DeleteObjectWithContext(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.BucketName),
		Key:    aws.String(fullPath),
	})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to delete object '%s' on %s, %s", fullPath, s.config.provider(), err.Error())
	}
	return nil
}

func (s *S3Drive) objectKeyFullPath(key string) string {
	cleanKey := cleanObjectKey(key)
	keyPrefix := s.config.keyPrefix()
	if keyPrefix == "" {
		return cleanKey
	}
	if cleanKey == "" {
		return keyPrefix
	}
	if cleanKey == keyPrefix || strings.HasPrefix(cleanKey, keyPrefix+"/") {
		return cleanKey
	}
	return path.Join(keyPrefix, cleanKey)
}

func (c *S3Config) keyPrefix() string {
	if c.KeyPrefix != "" {
		return cleanObjectKey(c.KeyPrefix)
	}
	u, _ := url.Parse(c.publicBaseURL())
	return cleanObjectKey(u.Path)
}

func cleanObjectKey(key string) string {
	key = strings.TrimSpace(strings.TrimLeft(key, "/"))
	if key == "" {
		return ""
	}
	return strings.TrimLeft(path.Clean(key), "/")
}

func isNotFound(err error) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == s3.ErrCodeNoSuchKey
	}
	return false
}

func (s *S3Drive) publicURL(objectKey string) string {
	baseURL := strings.TrimRight(s.config.publicBaseURL(), "/")
	key := cleanObjectKey(objectKey)
	u, _ := url.Parse(baseURL)
	urlPrefix := cleanObjectKey(u.Path)
	if urlPrefix != "" && (key == urlPrefix || strings.HasPrefix(key, urlPrefix+"/")) {
		key = strings.TrimPrefix(key, urlPrefix)
		key = strings.TrimLeft(key, "/")
	}
	return fmt.Sprintf("%s/%s", baseURL, key)
}

func (c *S3Config) publicBaseURL() string {
	if c.PublicBaseURL != "" {
		return c.PublicBaseURL
	}
	return c.CloudFrontBaseURL
}

package drive

import (
	"context"
	"fmt"
	"time"
)

type S3Config struct {
	Region            string `yaml:"region"`
	AccessKeyID       string `yaml:"accessKeyID"`
	SecretAccessKey   string `yaml:"secretAccessKey"`
	BucketName        string `yaml:"bucketName"`
	Endpoint          string `yaml:"endpoint"`
	PublicBaseURL     string `yaml:"publicBaseURL"`
	KeyPrefix         string `yaml:"keyPrefix"`
	Provider          string `yaml:"provider"`
	CloudFrontBaseURL string `yaml:"cloudFrontBaseURL"`
}

type FSConfig struct {
	UploadPath string `yaml:"uploadPath"`
	Directory  string `yaml:"directory"`
	BaseURL    string `yaml:"baseURL"`
}

type Config struct {
	TempOpt Option
	FS      FSConfig
	S3      S3Config
}

type BaseDrive struct {
	tempOpt Option
	fs      *FSDrive
	s3      *S3Drive
}

func InitWithConfig(config Config) *BaseDrive {
	return &BaseDrive{
		tempOpt: config.TempOpt,
		fs:      initFSDrive(config.FS),
		s3:      initS3Drive(config.S3),
	}
}

func (d *BaseDrive) Object(ctx context.Context, opt Option) *ObjectDrive {
	switch opt {
	case FSDriveOption:
		return newObjectDrive(FSDriveOption, d.fs)
	case S3DriveOption:
		return newObjectDrive(S3DriveOption, d.s3.withContext(ctx))
	case R2DriveOption:
		return newObjectDrive(R2DriveOption, d.s3.withContext(ctx))
	default:
		return newObjectDrive(FSDriveOption, d.fs)
	}
}

func (d *BaseDrive) S3(ctx context.Context) *ObjectDrive {
	return d.Object(ctx, S3DriveOption)
}

func (d *BaseDrive) R2(ctx context.Context) *ObjectDrive {
	return d.Object(ctx, R2DriveOption)
}

func (d *BaseDrive) Cloud(ctx context.Context) *ObjectDrive {
	return d.Object(ctx, d.s3.config.cloudOption())
}

func (d *BaseDrive) FS() *ObjectDrive {
	return d.Object(context.Background(), FSDriveOption)
}

func (d *BaseDrive) Temp(ctx context.Context) *ObjectDrive {
	return d.Object(ctx, d.tempOpt)
}

type objectDriveBackend interface {
	ListBlobs(prefix string) ([]string, error)
	ListBlobsUntil(prefix string, until *time.Time) ([]string, error)
	GetBlobByKey(key string) (*GetBlob, error)
	SaveBlob(b *NewBlob) error
	SaveBlobWithResult(b *NewBlob) (*SaveBlobResult, error)
	DeleteBlob(key string) error
}

type ObjectDrive struct {
	opt     Option
	backend objectDriveBackend
}

func newObjectDrive(opt Option, backend objectDriveBackend) *ObjectDrive {
	return &ObjectDrive{opt: opt, backend: backend}
}

func (d *ObjectDrive) Option() Option {
	return d.opt
}

func (d *ObjectDrive) ListBlobs(prefix string) ([]string, error) {
	if d == nil || d.backend == nil {
		return nil, fmt.Errorf("object drive is not configured")
	}
	return d.backend.ListBlobs(prefix)
}

func (d *ObjectDrive) ListBlobsUntil(prefix string, until *time.Time) ([]string, error) {
	if d == nil || d.backend == nil {
		return nil, fmt.Errorf("object drive is not configured")
	}
	return d.backend.ListBlobsUntil(prefix, until)
}

func (d *ObjectDrive) GetBlobByKey(key string) (*GetBlob, error) {
	if d == nil || d.backend == nil {
		return nil, fmt.Errorf("object drive is not configured")
	}
	return d.backend.GetBlobByKey(key)
}

func (d *ObjectDrive) SaveBlob(b *NewBlob) error {
	if d == nil || d.backend == nil {
		return fmt.Errorf("object drive is not configured")
	}
	return d.backend.SaveBlob(b)
}

func (d *ObjectDrive) SaveBlobWithResult(b *NewBlob) (*SaveBlobResult, error) {
	if d == nil || d.backend == nil {
		return nil, fmt.Errorf("object drive is not configured")
	}
	return d.backend.SaveBlobWithResult(b)
}

func (d *ObjectDrive) DeleteBlob(key string) error {
	if d == nil || d.backend == nil {
		return fmt.Errorf("object drive is not configured")
	}
	return d.backend.DeleteBlob(key)
}

func (s *S3Drive) withContext(ctx context.Context) *S3Drive {
	s3Drive := *s
	s3Drive.ctx = ctx
	return &s3Drive
}

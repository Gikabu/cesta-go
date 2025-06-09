package drive

import "context"

type S3Config struct {
	Region            string `yaml:"region"`
	AccessKeyID       string `yaml:"accessKeyID"`
	SecretAccessKey   string `yaml:"secretAccessKey"`
	BucketName        string `yaml:"bucketName"`
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
	base := &BaseDrive{
		tempOpt: config.TempOpt,
		fs:      initFSDrive(config.FS),
		s3:      initS3Drive(config.S3),
	}
	return base
}

func (d *BaseDrive) S3(ctx context.Context) *S3Drive {
	s3Drive := d.s3
	s3Drive.ctx = ctx
	return s3Drive
}

func (d *BaseDrive) FS() *FSDrive {
	return d.fs
}

func (d *BaseDrive) Temp() TempDrive {
	switch d.tempOpt {
	case S3DriveOption:
		return d.s3
	default:
		return d.fs
	}
}

type TempDrive interface {
	ListBlobs(prefix string) ([]string, error)
	GetBlobByKey(key string) (*GetBlob, error)
	SaveBlob(b *NewBlob) error
	DeleteBlob(key string) error
}

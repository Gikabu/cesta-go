package drive

import (
	"time"
)

type NewBlob struct {
	Key         string
	Content     []byte
	ContentType string
}

type GetBlob struct {
	Size         int64
	Content      []byte
	ContentType  string
	LastModified time.Time
}

type SaveBlobResult struct {
	Path        string
	URL         string
	Size        int64
	ContentType string
	Drive       string
}

type FileMeta struct {
	Path   string `json:"path"`
	Length int64  `json:"length"`
	Drive  string `json:"drive"`
}

func (r *SaveBlobResult) FileMeta() *FileMeta {
	return &FileMeta{
		Path:   r.Path,
		Length: r.Size,
		Drive:  r.Drive,
	}
}

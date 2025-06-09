package drive

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
)

var perm os.FileMode = 0744

type FSDrive struct {
	UploadPath string
	BaseURL    string
}

func initFSDrive(config FSConfig) *FSDrive {
	return &FSDrive{UploadPath: path.Join(config.UploadPath, config.Directory), BaseURL: config.BaseURL}
}

func (f *FSDrive) ListBlobs(prefix string) ([]string, error) {
	basePath := fsBasePath(f.UploadPath, prefix)
	files := make([]string, 0)

	err := filepath.Walk(basePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, prefix+path[len(basePath)+1:])
			}
			return nil
		})
	if err != nil && !os.IsNotExist(err) {
		return []string{}, fmt.Errorf("failed to read dir '%s', %s", basePath, err.Error())
	}
	sort.Strings(files)
	return files, nil
}

func (f *FSDrive) GetBlobByKey(key string) (*GetBlob, error) {
	fullPath := fsKeyFullPath(f.UploadPath, key)
	stats, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to get stats '%s' from FileSystem, %s", key, err.Error())
	}

	file, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob '%s' from FileSystem, %s", key, err.Error())
	}

	return &GetBlob{
		Size:        stats.Size(),
		Content:     file,
		ContentType: http.DetectContentType(file),
	}, nil
}

func (f *FSDrive) SaveBlob(b *NewBlob) error {
	if err := ValidateKey(b.Key); err != nil {
		return fmt.Errorf("failed to validate blob key '%s', %s", b.Key, err.Error())
	}

	fullPath := fsKeyFullPath(f.UploadPath, b.Key)
	err := os.MkdirAll(filepath.Dir(fullPath), perm)

	if err != nil {
		return fmt.Errorf("failed to create folder '%s' on FileSystem, %s", fullPath, err.Error())
	}

	err = os.WriteFile(fullPath, b.Content, perm)
	if err != nil {
		return fmt.Errorf("failed to create file '%s' on FileSystem, %s", fullPath, err.Error())
	}

	return nil
}

func (f *FSDrive) SaveBlobWithResult(b *NewBlob) (*SaveBlobResult, error) {
	if err := ValidateKey(b.Key); err != nil {
		return nil, fmt.Errorf("failed to validate blob key '%s', %s", b.Key, err.Error())
	}

	fullPath := fsKeyFullPath(f.UploadPath, b.Key)
	err := os.MkdirAll(filepath.Dir(fullPath), perm)

	if err != nil {
		return nil, fmt.Errorf("failed to create folder '%s' on FileSystem, %s", fullPath, err.Error())
	}

	err = os.WriteFile(fullPath, b.Content, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to create file '%s' on FileSystem, %s", fullPath, err.Error())
	}

	stats, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats '%s' from FileSystem, %s", b.Key, err.Error())
	}

	return &SaveBlobResult{
		Path:        b.Key,
		URL:         f.fsURL(b.Key),
		Size:        stats.Size(),
		ContentType: b.ContentType,
		Drive:       string(FSDriveOption),
	}, nil
}

func (f *FSDrive) DeleteBlob(key string) error {
	fullPath := fsKeyFullPath(f.UploadPath, key)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file '%s' from FileSystem, %s", key, err.Error())
	}
	return nil
}

func fsKeyFullPath(uploadPath, key string) string {
	return path.Join(uploadPath, key)
}

func fsBasePath(uploadPath, segment string) string {
	return path.Join(uploadPath, segment)
}

func (f *FSDrive) fsURL(key string) string {
	return fmt.Sprintf("%s/%s", f.BaseURL, key)
}

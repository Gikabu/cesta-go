package drive

import (
	"fmt"
	"github.com/gosimple/slug"
	"path/filepath"
	"strings"
)

// SanitizeFileName replaces invalid characters from given filename
func SanitizeFileName(fileName string, removeExt bool) string {
	fileName = strings.TrimSpace(fileName)
	ext := filepath.Ext(fileName)
	if removeExt {
		return strings.TrimSuffix(slug.Make(fileName[0:len(fileName)-len(ext)]), ".")
	}
	if ext != "" {
		return slug.Make(fileName[0:len(fileName)-len(ext)]) + ext
	}
	return slug.Make(fileName)
}

// ValidateKey checks if key is valid format
func ValidateKey(key string) error {
	if len(key) == 0 || len(key) > 512 || strings.Contains(key, " ") {
		return fmt.Errorf("blob key is in invalid format")
	}
	if strings.HasPrefix(key, "/") || strings.HasSuffix(key, "/") {
		return fmt.Errorf("blob key is in invalid format")
	}
	return nil
}

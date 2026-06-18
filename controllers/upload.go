package controllers

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// maxUploadBytes caps the accepted size of an uploaded image (5 MB).
const maxUploadBytes = 5 << 20

// allowedImageExts is the set of permitted image file extensions.
var allowedImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// allowedImageMimes is the set of permitted MIME types as declared by the
// client. This is a defence-in-depth check alongside the extension allow-list.
var allowedImageMimes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// errNoFile is returned when the request carries no file under the given field.
var errNoFile = errors.New("no file uploaded")

// saveUploadedImage validates and stores an uploaded image found under the
// given form field, returning a publicly accessible URL. subdir is an optional
// path segment under uploads/ (e.g. "proofs"). It rejects oversized files and
// any file whose extension or declared MIME type is not an allowed image type.
// errNoFile is returned when the field is absent so callers can treat the
// upload as optional.
func saveUploadedImage(c *gin.Context, field, subdir string) (string, error) {
	file, err := c.FormFile(field)
	if err != nil {
		return "", errNoFile
	}

	if file.Size > maxUploadBytes {
		return "", fmt.Errorf("file too large (max %d MB)", maxUploadBytes>>20)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		return "", errors.New("unsupported file type")
	}
	if ct := contentType(file); ct != "" && !allowedImageMimes[ct] {
		return "", errors.New("unsupported file type")
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	rel := filepath.Join("uploads", subdir, filename)
	if err := c.SaveUploadedFile(file, rel); err != nil {
		return "", errors.New("failed to store file")
	}

	urlPath := "/uploads/"
	if subdir != "" {
		urlPath += subdir + "/"
	}
	return "http://" + c.Request.Host + urlPath + filename, nil
}

// contentType returns the Content-Type the client declared for the upload, if any.
func contentType(file *multipart.FileHeader) string {
	if file.Header == nil {
		return ""
	}
	// Strip any parameters such as "; charset=...".
	return strings.TrimSpace(strings.SplitN(file.Header.Get("Content-Type"), ";", 2)[0])
}

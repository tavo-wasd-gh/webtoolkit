package s3

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Client struct {
	RootDir string
}

func New(root string) *Client {
	return &Client{RootDir: root}
}

func (s *Client) getPath(bucket, key string) string {
	return filepath.Join(s.RootDir, bucket, key)
}

func (s *Client) CreateBucket(bucket string) error {
	path := filepath.Join(s.RootDir, bucket)
	return os.MkdirAll(path, 0755)
}

func (s *Client) PutObject(bucket, key string, reader io.Reader) error {
	fullPath := s.getPath(bucket, key)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, reader)
	return err
}

func (s *Client) GetObject(bucket, key string) (*os.File, error) {
	fullPath := s.getPath(bucket, key)
	return os.Open(fullPath)
}

func (s *Client) DeleteObject(bucket, key string) error {
	fullPath := s.getPath(bucket, key)
	return os.Remove(fullPath)
}

func DetectFileType(file multipart.File) (string, []byte, error) {
	const sniffLen = 512
	buffer := make([]byte, sniffLen)

	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	if seeker, ok := file.(io.Seeker); ok {
		_, _ = seeker.Seek(0, io.SeekStart)
	} else {
		return "", nil, fmt.Errorf("file is not seekable")
	}

	mimeType := http.DetectContentType(buffer[:n])
	return mimeType, buffer[:n], nil
}

func VerifyFileType(file multipart.File, allowedMimeTypes []string) error {
	mimeType, _, err := DetectFileType(file)
	if err != nil {
		return err
	}

	valid := false
	for _, mt := range allowedMimeTypes {
		if strings.EqualFold(mimeType, mt) {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid file type: %s", mimeType)
	}

	return nil
}

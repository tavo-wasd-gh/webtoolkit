package s3

import (
    "io"
    "os"
    "path/filepath"
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

package s3

import (
    "io"
    "os"
    "path/filepath"
)

type LocalS3 struct {
    RootDir string
}

func New(root string) *LocalS3 {
    return &LocalS3{RootDir: root}
}

func (s *LocalS3) getPath(bucket, key string) string {
    return filepath.Join(s.RootDir, bucket, key)
}

func (s *LocalS3) CreateBucket(bucket string) error {
    path := filepath.Join(s.RootDir, bucket)
    return os.MkdirAll(path, 0755)
}

func (s *LocalS3) PutObject(bucket, key string, reader io.Reader) error {
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

func (s *LocalS3) GetObject(bucket, key string) (*os.File, error) {
    fullPath := s.getPath(bucket, key)
    return os.Open(fullPath)
}

func (s *LocalS3) DeleteObject(bucket, key string) error {
    fullPath := s.getPath(bucket, key)
    return os.Remove(fullPath)
}

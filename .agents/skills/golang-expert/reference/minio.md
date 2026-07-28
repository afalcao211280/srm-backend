# MinIO (minio-go/v7)

## Setup

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "time"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct { client *minio.Client; bucket string }

func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Storage, error) {
    cli, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil { return nil, fmt.Errorf("minio new: %w", err) }
    
    ctx := context.Background()
    ok, err := cli.BucketExists(ctx, bucket)
    if err != nil { return nil, fmt.Errorf("bucket exists: %w", err) }
    if !ok {
        if err := cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, fmt.Errorf("create bucket: %w", err)
        }
    }
    return &Storage{client: cli, bucket: bucket}, nil
}
```

## Operacoes

```go
func (s *Storage) Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
    _, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{ContentType: contentType})
    return err
}

func (s *Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
    return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) PresignedGet(ctx context.Context, key string, expiry time.Duration) (string, error) {
    u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
    if err != nil { return "", err }
    return u.String(), nil
}
```

## Padroes
- Bucket criado automaticamente se nao existir
- Presigned URLs para downloads seguros
- ContentType explicito no upload

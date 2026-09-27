package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Destination struct {
	client          *minio.Client
	bucket          string
	prefix          string
	hostname        string
	retentionCopies int
	retentionDays   int
	contentType     string
}

type DestinationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewS3Destination(config DestinationConfig, hostname string) (*S3Destination, error) {
	attrs := config.GetOtherAttributes()

	endpoint := attrs["endpoint_url"]
	if endpoint == "" {
		endpoint = "s3.amazonaws.com"
	}

	bucket := attrs["bucket"]
	if bucket == "" {
		return nil, fmt.Errorf("S3 destination missing required 'bucket' attribute")
	}

	region := attrs["region"]
	if region == "" {
		region = "us-east-1"
	}

	accessKey := attrs["aws_access_key_id"]
	secretKey := attrs["aws_secret_access_key"]

	useSSL := !strings.HasPrefix(endpoint, "http://")

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %v", err)
	}

	retentionCopies := 0
	if rc, err := strconv.Atoi(attrs["retention_copies"]); err == nil {
		retentionCopies = rc
	}

	retentionDays := 0
	if rd, err := strconv.Atoi(attrs["retention_days"]); err == nil {
		retentionDays = rd
	}

	prefix := attrs["prefix"]
	if prefix == "" {
		prefix = "backups"
	}

	if hostname == "" {
		h, _ := os.Hostname()
		hostname = h
	}

	return &S3Destination{
		client:          client,
		bucket:          bucket,
		prefix:          prefix,
		hostname:        hostname,
		retentionCopies: retentionCopies,
		retentionDays:   retentionDays,
		contentType:     "application/octet-stream",
	}, nil
}

func (s *S3Destination) UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	now := time.Now()
	dateDir := now.Format("2006-01-02")

	s3Path := fmt.Sprintf("%s/%s/%s/%s", s.prefix, s.hostname, dateDir, filename)

	buf := new(bytes.Buffer)
	size, err := io.Copy(buf, reader)
	if err != nil {
		return "", fmt.Errorf("failed to read backup data: %v", err)
	}

	uploadInfo, err := s.client.PutObject(ctx, s.bucket, s3Path, buf, size, minio.PutObjectOptions{
		ContentType: s.contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	log.Printf("Successfully uploaded %s to s3://%s/%s (size: %d bytes, etag: %s)",
		filename, s.bucket, s3Path, uploadInfo.Size, uploadInfo.ETag)

	if s.retentionCopies > 0 || s.retentionDays > 0 {
		if err := s.enforceRetention(ctx); err != nil {
			log.Printf("Warning: failed to enforce retention policy: %v", err)
		}
	}

	return s3Path, nil
}

func (s *S3Destination) enforceRetention(ctx context.Context) error {
	backups := []minio.ObjectInfo{}
	opts := minio.ListObjectsOptions{
		Prefix:    s.prefix + "/" + s.hostname + "/",
		Recursive: true,
	}

	for object := range s.client.ListObjects(ctx, s.bucket, opts) {
		if object.Err != nil {
			return object.Err
		}
		backups = append(backups, object)
	}

	if s.retentionCopies > 0 && len(backups) > s.retentionCopies {
		toDelete := len(backups) - s.retentionCopies
		for i := 0; i < toDelete; i++ {
			log.Printf("Deleting old backup: %s", backups[i].Key)
			err := s.client.RemoveObject(ctx, s.bucket, backups[i].Key, minio.RemoveObjectOptions{})
			if err != nil {
				log.Printf("Failed to delete %s: %v", backups[i].Key, err)
			}
		}
	}

	if s.retentionDays > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -s.retentionDays)
		for _, backup := range backups {
			if backup.LastModified.Before(cutoffTime) {
				log.Printf("Deleting expired backup: %s", backup.Key)
				err := s.client.RemoveObject(ctx, s.bucket, backup.Key, minio.RemoveObjectOptions{})
				if err != nil {
					log.Printf("Failed to delete %s: %v", backup.Key, err)
				}
			}
		}
	}

	return nil
}

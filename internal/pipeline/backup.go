package pipeline

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/rossigee/backupx/internal/encryption"
	"github.com/rossigee/backupx/notifications"
)

type BackupPipeline struct {
	encryptor    *encryption.GPGEncryptor
	destinations []DestinationWriter
	passphrase   string
	sourceId     string
	sourceName   string
	compressOnly bool
	startTime    time.Time
}

type DestinationWriter interface {
	UploadReader(ctx context.Context, reader io.Reader, filename string) (string, error)
}

type PipelineConfig struct {
	SourceId     string
	SourceName   string
	Passphrase   string
	CompressOnly bool
	Destinations []DestinationWriter
}

func NewBackupPipeline(config PipelineConfig) *BackupPipeline {
	return &BackupPipeline{
		passphrase:   config.Passphrase,
		sourceId:     config.SourceId,
		sourceName:   config.SourceName,
		compressOnly: config.CompressOnly,
		destinations: config.Destinations,
		encryptor:    encryption.NewGPGEncryptor(config.Passphrase, config.CompressOnly),
		startTime:    time.Now(),
	}
}

func (bp *BackupPipeline) Execute(ctx context.Context, sourceReader io.ReadCloser, filename string) notifications.BackupResult {
	defer func() {
		_ = sourceReader.Close()
	}()

	result := notifications.BackupResult{
		SourceId:   bp.sourceId,
		SourceName: bp.sourceName,
		Success:    false,
	}

	encryptedReader, err := bp.encryptor.Encrypt(sourceReader)
	if err != nil {
		result.Error = fmt.Sprintf("encryption failed: %v", err)
		result.Duration = time.Since(bp.startTime).Milliseconds()
		return result
	}
	defer func() {
		_ = encryptedReader.Close()
	}()

	if bp.compressOnly {
		filename += ".gz"
	} else {
		filename += ".gpg"
	}

	var uploadedPath string
	var uploadErr error

	if len(bp.destinations) > 0 {
		for i, dest := range bp.destinations {
			log.Printf("Uploading to destination %d of %d", i+1, len(bp.destinations))

			path, err := dest.UploadReader(ctx, encryptedReader, filename)
			if err != nil {
				log.Printf("Upload failed: %v", err)
				uploadErr = err
				result.Error = fmt.Sprintf("upload to destination %d failed: %v", i, err)
				break
			}

			uploadedPath = path
			log.Printf("Successfully uploaded to: %s", path)
		}
	} else {
		_, err := io.Copy(io.Discard, encryptedReader)
		if err != nil {
			uploadErr = err
			result.Error = fmt.Sprintf("failed to read backup: %v", err)
		}
	}

	result.Duration = time.Since(bp.startTime).Milliseconds()

	if uploadErr != nil {
		return result
	}

	result.Success = true
	result.BackupPath = uploadedPath
	if uploadedPath == "" {
		result.BackupPath = filename
	}
	result.Message = fmt.Sprintf("Backup completed in %dms", result.Duration)

	return result
}

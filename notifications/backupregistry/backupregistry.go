package backupregistry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rossigee/backupx/notifications"
)

type BackupRegistry struct {
	url        string
	token      string
	metadata   map[string]string
	httpClient *http.Client
}

type NotificationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

type backupRunPayload struct {
	RunID              string            `json:"run_id"`
	JobName            string            `json:"job_name"`
	AgentID            string            `json:"agent_id"`
	StartTime          string            `json:"start_time"`
	EndTime            string            `json:"end_time"`
	Status             string            `json:"status"`
	BytesBackedUp      int64             `json:"bytes_backed_up,omitempty"`
	Encrypted          bool              `json:"encrypted"`
	EncryptionStatus   string            `json:"encryption_status"`
	BackupURL          string            `json:"backup_url,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Error              string            `json:"error,omitempty"`
}

func NewBackupRegistry(config NotificationConfig) (*BackupRegistry, error) {
	attrs := config.GetOtherAttributes()

	url := attrs["url"]
	if url == "" {
		return nil, fmt.Errorf("backup registry notification missing required 'url' attribute")
	}

	token := attrs["token"]

	metadata := make(map[string]string)
	for k, v := range attrs {
		if k != "url" && k != "token" && k != "notify_on_success" && k != "notify_on_failure" {
			metadata[k] = v
		}
	}

	return &BackupRegistry{
		url:      url,
		token:    token,
		metadata: metadata,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (b *BackupRegistry) Notify(ctx context.Context, result notifications.BackupResult) error {
	now := time.Now().UTC()
	startTime := now.Add(-time.Duration(result.Duration) * time.Millisecond)

	encrypted := false
	encryptionStatus := "unencrypted"
	if result.BackupPath != "" && bytes.Contains([]byte(result.BackupPath), []byte(".gpg")) {
		encrypted = true
		encryptionStatus = "encrypted"
	}

	payload := backupRunPayload{
		RunID:            uuid.New().String(),
		JobName:          result.SourceName,
		AgentID:          result.SourceId,
		StartTime:        startTime.Format(time.RFC3339),
		EndTime:          now.Format(time.RFC3339),
		Encrypted:        encrypted,
		EncryptionStatus: encryptionStatus,
		Metadata:         b.metadata,
	}

	if result.Success {
		payload.Status = "success"
		payload.BackupURL = result.BackupPath
	} else {
		payload.Status = "failure"
		payload.Error = result.Error
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.url, bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if b.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.token))
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("backup registry returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

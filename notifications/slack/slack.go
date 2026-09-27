package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rossigee/backupx/notifications"
)

type SlackNotification struct {
	webhookURL         string
	notifyOnSuccess    bool
	notifyOnFailure    bool
	httpClient         *http.Client
}

type NotificationConfig interface {
	GetId() string
	GetType() string
	GetName() string
	GetOtherAttributes() map[string]string
}

func NewSlackNotification(config NotificationConfig) (*SlackNotification, error) {
	attrs := config.GetOtherAttributes()

	webhookURL := attrs["url"]
	if webhookURL == "" {
		return nil, fmt.Errorf("slack notification missing required 'url' attribute")
	}

	notifyOnSuccess := attrs["notify_on_success"] != "0" && attrs["notify_on_success"] != ""
	notifyOnFailure := attrs["notify_on_failure"] != "0" && attrs["notify_on_failure"] != ""

	if !notifyOnSuccess && !notifyOnFailure {
		notifyOnSuccess = true
		notifyOnFailure = true
	}

	return &SlackNotification{
		webhookURL:      webhookURL,
		notifyOnSuccess: notifyOnSuccess,
		notifyOnFailure: notifyOnFailure,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

type slackMessage struct {
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	Color  string  `json:"color"`
	Title  string  `json:"title"`
	Fields []field `json:"fields"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func (s *SlackNotification) Notify(ctx context.Context, result notifications.BackupResult) error {
	shouldNotify := (result.Success && s.notifyOnSuccess) || (!result.Success && s.notifyOnFailure)

	if !shouldNotify {
		return nil
	}

	color := "good"
	statusText := "✅ Backup Successful"
	if !result.Success {
		color = "danger"
		statusText = "❌ Backup Failed"
	}

	fields := []field{
		{
			Title: "Source",
			Value: fmt.Sprintf("%s (%s)", result.SourceId, result.SourceName),
			Short: false,
		},
		{
			Title: "Duration",
			Value: fmt.Sprintf("%d ms", result.Duration),
			Short: true,
		},
	}

	if result.Success {
		fields = append(fields, field{
			Title: "Backup Path",
			Value: result.BackupPath,
			Short: false,
		})
	} else {
		fields = append(fields, field{
			Title: "Error",
			Value: result.Error,
			Short: false,
		})
		if result.Message != "" {
			fields = append(fields, field{
				Title: "Details",
				Value: result.Message,
				Short: false,
			})
		}
	}

	msg := slackMessage{
		Text: statusText,
		Attachments: []attachment{
			{
				Color:  color,
				Title:  "Backup Report",
				Fields: fields,
			},
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

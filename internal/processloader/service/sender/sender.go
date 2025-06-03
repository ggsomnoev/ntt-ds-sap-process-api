package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/logger"
	"github.com/ggsomnoev/ntt-ds-sap-process-api/internal/model"
)

type HTTPProcessSender struct {
	client  *http.Client
	baseURL string
}

func NewHTTPProcessSender(baseURL string) *HTTPProcessSender {
	return &HTTPProcessSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

func (s *HTTPProcessSender) Send(def model.ProcessDefinition) error {
	url := fmt.Sprintf("%s/startProcess", s.baseURL)

	body, err := json.Marshal(def)
	if err != nil {
		return fmt.Errorf("failed to marshal process: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.GetLogger().Errorf("could not close response body: %v", err)
		}
	}()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

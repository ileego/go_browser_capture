package application

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/ileego/go_browser_capture/backend/domain"
	"github.com/ileego/go_browser_capture/backend/scheduler"
	"github.com/ileego/go_browser_capture/backend/websocket"
)

type BatchCaptureRequestItem struct {
	ID  string `json:"id" binding:"required"`
	URL string `json:"url" binding:"required"`
}

type BatchCaptureRequest struct {
	URLs    []BatchCaptureRequestItem `json:"urls" binding:"required"`
	Timeout int                       `json:"timeout,omitempty"`
}

type BatchCaptureResult struct {
	ID         string                 `json:"id"`
	URL        string                 `json:"url"`
	Success    bool                   `json:"success"`
	Data       *websocket.CaptureData `json:"data,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ConfigID   string                 `json:"config_id,omitempty"`
	ConfigName string                 `json:"config_name,omitempty"`
}

type BatchCaptureResponse struct {
	Success      bool                  `json:"success"`
	Results      []*BatchCaptureResult `json:"results"`
	Total        int                   `json:"total"`
	SuccessCount int                   `json:"success_count"`
	FailedCount  int                   `json:"failed_count"`
	Timestamp    string                `json:"timestamp"`
}

type BatchCaptureAppService struct {
	selectorDomainService *domain.SelectorConfigService
	scheduler             *scheduler.TaskScheduler
}

func NewBatchCaptureAppService(selectorDomainService *domain.SelectorConfigService, scheduler *scheduler.TaskScheduler) *BatchCaptureAppService {
	return &BatchCaptureAppService{
		selectorDomainService: selectorDomainService,
		scheduler:             scheduler,
	}
}

func (s *BatchCaptureAppService) CaptureBatch(ctx context.Context, req BatchCaptureRequest) (*BatchCaptureResponse, error) {
	timeout := time.Second * 60
	if req.Timeout > 0 {
		timeout = time.Millisecond * time.Duration(req.Timeout)
	}

	results := make([]*BatchCaptureResult, 0, len(req.URLs))
	var wg sync.WaitGroup
	resultChan := make(chan *BatchCaptureResult, len(req.URLs))

	for _, item := range req.URLs {
		wg.Add(1)
		go func(id, urlStr string) {
			defer wg.Done()
			result := s.captureURL(ctx, id, urlStr, timeout)
			resultChan <- result
		}(item.ID, item.URL)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		results = append(results, result)
	}

	successCount := 0
	failedCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failedCount++
		}
	}

	return &BatchCaptureResponse{
		Success:      successCount > 0,
		Results:      results,
		Total:        len(req.URLs),
		SuccessCount: successCount,
		FailedCount:  failedCount,
		Timestamp:    time.Now().Format(time.RFC3339),
	}, nil
}

func (s *BatchCaptureAppService) captureURL(ctx context.Context, id, urlStr string, timeout time.Duration) *BatchCaptureResult {
	maxRetries := 3
	baseDelay := time.Second * 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result := s.doCaptureURL(ctx, id, urlStr, timeout)
		if result.Success {
			return result
		}
		if !isRetryableError(result.Error) {
			return result
		}
		if attempt < maxRetries {
			delay := baseDelay * time.Duration(attempt)
			time.Sleep(delay)
		}
	}

	return &BatchCaptureResult{
		ID:      id,
		URL:     urlStr,
		Success: false,
		Error:   "request timeout after 3 retries",
	}
}

func isRetryableError(err string) bool {
	if err == "" {
		return false
	}
	retryableErrors := []string{
		"request timeout",
		"plugin not available",
		"task timeout",
		"timeout",
		"no available plugin",
		"task queue is full",
	}
	for _, retryable := range retryableErrors {
		if err == retryable || len(err) >= len(retryable) && err[:len(retryable)] == retryable {
			return true
		}
	}
	return false
}

func (s *BatchCaptureAppService) doCaptureURL(ctx context.Context, id, urlStr string, timeout time.Duration) *BatchCaptureResult {
	domain, err := extractDomain(urlStr)
	if err != nil {
		return &BatchCaptureResult{
			ID:      id,
			URL:     urlStr,
			Success: false,
			Error:   err.Error(),
		}
	}

	configs, err := s.selectorDomainService.GetByDomain(ctx, domain)
	if err != nil {
		return &BatchCaptureResult{
			ID:      id,
			URL:     urlStr,
			Success: false,
			Error:   err.Error(),
		}
	}

	if len(configs) == 0 {
		return &BatchCaptureResult{
			ID:      id,
			URL:     urlStr,
			Success: false,
			Error:   "no selector config found for domain: " + domain,
		}
	}

	config := configs[0]
	taskID := generateRequestID(id)

	ch, err := s.scheduler.SubmitTask(ctx, taskID, urlStr, config.Selector, timeout)
	if err != nil {
		return &BatchCaptureResult{
			ID:         id,
			URL:        urlStr,
			Success:    false,
			Error:      err.Error(),
			ConfigID:   string(config.ID),
			ConfigName: config.Name,
		}
	}

	select {
	case resp := <-ch:
		if resp.Success && resp.Data != nil {
			return &BatchCaptureResult{
				ID:         id,
				URL:        urlStr,
				Success:    true,
				Data:       resp.Data,
				ConfigID:   string(config.ID),
				ConfigName: config.Name,
			}
		}
		return &BatchCaptureResult{
			ID:         id,
			URL:        urlStr,
			Success:    false,
			Error:      resp.Error,
			ConfigID:   string(config.ID),
			ConfigName: config.Name,
		}
	case <-time.After(timeout):
		return &BatchCaptureResult{
			ID:         id,
			URL:        urlStr,
			Success:    false,
			Error:      "request timeout",
			ConfigID:   string(config.ID),
			ConfigName: config.Name,
		}
	}
}

func extractDomain(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return u.Hostname(), nil
}

func generateRequestID(id string) string {
	return fmt.Sprintf("batch_%s_%d", id, time.Now().UnixNano()%1000000)
}

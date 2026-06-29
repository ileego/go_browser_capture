package application

import (
	"context"

	"github.com/ileego/go_browser_capture/backend/domain"
)

type CreateSelectorConfigRequest struct {
	Name     string `json:"name" binding:"required"`
	Domain   string `json:"domain" binding:"required"`
	Selector string `json:"selector" binding:"required"`
	Regex    string `json:"regex"`
}

type UpdateSelectorConfigRequest struct {
	Name     string `json:"name" binding:"required"`
	Domain   string `json:"domain" binding:"required"`
	Selector string `json:"selector" binding:"required"`
	Regex    string `json:"regex"`
}

type SelectorConfigResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Selector  string `json:"selector"`
	Regex     string `json:"regex"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type SelectorConfigAppService struct {
	domainService *domain.SelectorConfigService
}

func NewSelectorConfigAppService(domainService *domain.SelectorConfigService) *SelectorConfigAppService {
	return &SelectorConfigAppService{domainService: domainService}
}

func (s *SelectorConfigAppService) Create(ctx context.Context, req CreateSelectorConfigRequest) (*SelectorConfigResponse, error) {
	config, err := s.domainService.Create(ctx, req.Name, req.Domain, req.Selector, req.Regex)
	if err != nil {
		return nil, err
	}
	return toResponse(config), nil
}

func (s *SelectorConfigAppService) Update(ctx context.Context, id string, req UpdateSelectorConfigRequest) (*SelectorConfigResponse, error) {
	config, err := s.domainService.Update(ctx, domain.SelectorConfigID(id), req.Name, req.Domain, req.Selector, req.Regex)
	if err != nil {
		return nil, err
	}
	return toResponse(config), nil
}

func (s *SelectorConfigAppService) GetByID(ctx context.Context, id string) (*SelectorConfigResponse, error) {
	config, err := s.domainService.GetByID(ctx, domain.SelectorConfigID(id))
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}
	return toResponse(config), nil
}

func (s *SelectorConfigAppService) GetByName(ctx context.Context, name string) (*SelectorConfigResponse, error) {
	config, err := s.domainService.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, nil
	}
	return toResponse(config), nil
}

func (s *SelectorConfigAppService) GetByDomain(ctx context.Context, domain string) ([]*SelectorConfigResponse, error) {
	configs, err := s.domainService.GetByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	return toResponseList(configs), nil
}

func (s *SelectorConfigAppService) GetAll(ctx context.Context) ([]*SelectorConfigResponse, error) {
	configs, err := s.domainService.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return toResponseList(configs), nil
}

func (s *SelectorConfigAppService) GetActive(ctx context.Context) ([]*SelectorConfigResponse, error) {
	configs, err := s.domainService.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	return toResponseList(configs), nil
}

func (s *SelectorConfigAppService) Delete(ctx context.Context, id string) error {
	return s.domainService.Delete(ctx, domain.SelectorConfigID(id))
}

func (s *SelectorConfigAppService) Activate(ctx context.Context, id string) (*SelectorConfigResponse, error) {
	config, err := s.domainService.Activate(ctx, domain.SelectorConfigID(id))
	if err != nil {
		return nil, err
	}
	return toResponse(config), nil
}

func (s *SelectorConfigAppService) Deactivate(ctx context.Context, id string) (*SelectorConfigResponse, error) {
	config, err := s.domainService.Deactivate(ctx, domain.SelectorConfigID(id))
	if err != nil {
		return nil, err
	}
	return toResponse(config), nil
}

func toResponse(config *domain.SelectorConfig) *SelectorConfigResponse {
	return &SelectorConfigResponse{
		ID:        string(config.ID),
		Name:      config.Name,
		Domain:    config.Domain,
		Selector:  config.Selector,
		Regex:     config.Regex,
		IsActive:  config.IsActive,
		CreatedAt: config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toResponseList(configs []*domain.SelectorConfig) []*SelectorConfigResponse {
	responses := make([]*SelectorConfigResponse, 0, len(configs))
	for _, config := range configs {
		responses = append(responses, toResponse(config))
	}
	return responses
}

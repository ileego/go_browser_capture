package domain

import (
	"context"
	"fmt"
)

type SelectorConfigRepository interface {
	Save(ctx context.Context, config *SelectorConfig) error
	FindByID(ctx context.Context, id SelectorConfigID) (*SelectorConfig, error)
	FindByName(ctx context.Context, name string) (*SelectorConfig, error)
	FindByDomain(ctx context.Context, domain string) ([]*SelectorConfig, error)
	FindAll(ctx context.Context) ([]*SelectorConfig, error)
	FindActive(ctx context.Context) ([]*SelectorConfig, error)
	Delete(ctx context.Context, id SelectorConfigID) error
}

type SelectorConfigService struct {
	repo SelectorConfigRepository
}

func NewSelectorConfigService(repo SelectorConfigRepository) *SelectorConfigService {
	return &SelectorConfigService{repo: repo}
}

func (s *SelectorConfigService) Create(ctx context.Context, name, domain, selector, regex string) (*SelectorConfig, error) {
	existing, err := s.repo.FindByName(ctx, name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("名称已存在: %s", name)
	}

	config, err := NewSelectorConfig(name, domain, selector, regex)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *SelectorConfigService) Update(ctx context.Context, id SelectorConfigID, name, domain, selector, regex string) (*SelectorConfig, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在: %s", id)
	}

	if name != config.Name {
		existing, err := s.repo.FindByName(ctx, name)
		if err == nil && existing != nil && existing.ID != id {
			return nil, fmt.Errorf("名称已存在: %s", name)
		}
	}

	if err := config.Update(name, domain, selector, regex); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *SelectorConfigService) GetByID(ctx context.Context, id SelectorConfigID) (*SelectorConfig, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *SelectorConfigService) GetByName(ctx context.Context, name string) (*SelectorConfig, error) {
	return s.repo.FindByName(ctx, name)
}

func (s *SelectorConfigService) GetByDomain(ctx context.Context, domain string) ([]*SelectorConfig, error) {
	return s.repo.FindByDomain(ctx, domain)
}

func (s *SelectorConfigService) GetAll(ctx context.Context) ([]*SelectorConfig, error) {
	return s.repo.FindAll(ctx)
}

func (s *SelectorConfigService) GetActive(ctx context.Context) ([]*SelectorConfig, error) {
	return s.repo.FindActive(ctx)
}

func (s *SelectorConfigService) Delete(ctx context.Context, id SelectorConfigID) error {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if config == nil {
		return fmt.Errorf("配置不存在: %s", id)
	}

	return s.repo.Delete(ctx, id)
}

func (s *SelectorConfigService) Activate(ctx context.Context, id SelectorConfigID) (*SelectorConfig, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在: %s", id)
	}

	config.Activate()
	if err := s.repo.Save(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *SelectorConfigService) Deactivate(ctx context.Context, id SelectorConfigID) (*SelectorConfig, error) {
	config, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, fmt.Errorf("配置不存在: %s", id)
	}

	config.Deactivate()
	if err := s.repo.Save(ctx, config); err != nil {
		return nil, err
	}

	return config, nil
}


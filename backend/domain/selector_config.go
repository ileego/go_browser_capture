package domain

import (
	"fmt"
	"regexp"
	"time"
)

type SelectorConfigID string

type SelectorConfig struct {
	ID         SelectorConfigID
	Name       string
	Domain     string
	Selector   string
	Regex      string
	IsActive   bool
	CreatedAt  time.Time
}

func NewSelectorConfig(name, domain, selector, regex string) (*SelectorConfig, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateDomain(domain); err != nil {
		return nil, err
	}
	if err := validateSelector(selector); err != nil {
		return nil, err
	}
	if regex != "" {
		if err := validateRegex(regex); err != nil {
			return nil, err
		}
	}

	return &SelectorConfig{
		Name:     name,
		Domain:   domain,
		Selector: selector,
		Regex:    regex,
		IsActive: true,
	}, nil
}

func (c *SelectorConfig) Update(name, domain, selector, regex string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := validateDomain(domain); err != nil {
		return err
	}
	if err := validateSelector(selector); err != nil {
		return err
	}
	if regex != "" {
		if err := validateRegex(regex); err != nil {
			return err
		}
	}

	c.Name = name
	c.Domain = domain
	c.Selector = selector
	c.Regex = regex
	return nil
}

func (c *SelectorConfig) Activate() {
	c.IsActive = true
}

func (c *SelectorConfig) Deactivate() {
	c.IsActive = false
}

func (c *SelectorConfig) ApplyRegex(content string) (string, bool) {
	if c.Regex == "" {
		return content, true
	}

	re, err := regexp.Compile(c.Regex)
	if err != nil {
		return "", false
	}

	match := re.FindString(content)
	if match == "" {
		return "", false
	}
	return match, true
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if len(name) > 100 {
		return fmt.Errorf("名称长度不能超过100个字符")
	}
	return nil
}

func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("域名不能为空")
	}
	if len(domain) > 255 {
		return fmt.Errorf("域名长度不能超过255个字符")
	}
	return nil
}

func validateSelector(selector string) error {
	if selector == "" {
		return fmt.Errorf("选择器不能为空")
	}
	if len(selector) > 500 {
		return fmt.Errorf("选择器长度不能超过500个字符")
	}
	return nil
}

func validateRegex(regex string) error {
	if _, err := regexp.Compile(regex); err != nil {
		return fmt.Errorf("无效的正则表达式: %w", err)
	}
	return nil
}

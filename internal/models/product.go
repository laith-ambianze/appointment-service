package models

import (
	"encoding/json"
	"fmt"
)

// ProductStatus represents the status of a product
type ProductStatus string

const (
	ProductStatusActive    ProductStatus = "active"
	ProductStatusInactive  ProductStatus = "inactive"
	ProductStatusSuspended ProductStatus = "suspended"
)

// Product represents a registered product in the system
type Product struct {
	BaseModel
	Name          string                 `json:"name"`
	Description   string                 `json:"description,omitempty"`
	APIKey        string                 `json:"api_key"`
	APISecretHash string                 `json:"-"` // Never expose in JSON
	Status        ProductStatus          `json:"status"`
	WebhookURL    string                 `json:"webhook_url,omitempty"`
	WebhookSecret string                 `json:"-"` // Never expose in JSON
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the product fields
func (p *Product) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("product name is required")
	}

	if len(p.Name) > 255 {
		return fmt.Errorf("product name must be less than 255 characters")
	}

	if p.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	if p.APISecretHash == "" {
		return fmt.Errorf("API secret hash is required")
	}

	if p.Status == "" {
		p.Status = ProductStatusActive
	}

	// Validate status
	if p.Status != ProductStatusActive &&
		p.Status != ProductStatusInactive &&
		p.Status != ProductStatusSuspended {
		return fmt.Errorf("invalid product status: %s", p.Status)
	}

	return nil
}

// IsActive returns true if the product is active
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive && !p.IsDeleted()
}

// MetadataJSON returns metadata as JSON string
func (p *Product) MetadataJSON() (string, error) {
	if p.Metadata == nil {
		return "{}", nil
	}

	bytes, err := json.Marshal(p.Metadata)
	if err != nil {
		return "", fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return string(bytes), nil
}

// CreateProductInput represents input for creating a product
type CreateProductInput struct {
	Name        string                 `json:"name" binding:"required,max=255"`
	Description string                 `json:"description" binding:"max=1000"`
	WebhookURL  string                 `json:"webhook_url" binding:"omitempty,url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UpdateProductInput represents input for updating a product
type UpdateProductInput struct {
	Name          *string                `json:"name" binding:"omitempty,max=255"`
	Description   *string                `json:"description" binding:"omitempty,max=1000"`
	Status        *ProductStatus         `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	WebhookURL    *string                `json:"webhook_url" binding:"omitempty,url"`
	WebhookSecret *string                `json:"webhook_secret"`
	Metadata      map[string]interface{} `json:"metadata"`
}

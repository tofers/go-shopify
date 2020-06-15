package goshopify

import (
	"fmt"
	"time"
)

const priceRuleBasePath = "price_rules"

// DiscountCodeService is an interface for interfacing with the discount endpoints
// of the Shopify API.
// See: https://shopify.dev/docs/admin-api/rest/reference/discounts/pricerule
type PriceRuleService interface {
	Create(PriceRule) (*PriceRule, error)
	Update(PriceRule) (*PriceRule, error)
	Count(interface{}) (int, error)
	List(int64) ([]PriceRule, error)
	Get(int64) (*PriceRule, error)
	Delete(int64) error
}

// PriceRuleServiceOp handles communication with the discount code
// related methods of the Shopify API.
type PriceRuleServiceOp struct {
	client *Client
}

// PriceRule represents a Shopify Discount Code
type PriceRule struct {
	ID                int64  `json:"id,omitempty"`
	Title             string `json:"title,omitempty"`
	TargetType        string `json:"target_type,omitempty"`
	TargetSelection   string `json:"target_selection,omitempty"`
	AllocationMethod  string `json:"allocation_method,omitempty"`
	ValueType         string `json:"value_type,omitempty"`
	Value             string `json:"value,omitempty"`
	CustomerSelection string `json:"customer_selection,omitempty"`

	UsageLimit      int  `json:"usage_limit,omitempty"`
	AllocationLimit int  `json:"allocation_limit,omitempty"`
	OncePerCustomer bool `json:"once_per_customer,omitempty"`

	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// PriceRuleResource represents the result from the price_rules/X.json endpoint
type PriceRuleResource struct {
	PriceRule *PriceRule `json:"price_rules"`
}

// PriceRulesResource is the result from the price_rules.json endpoint
type PriceRulesResource struct {
	PriceRule []PriceRule `json:"price_rules"`
}

// Create a price rule
func (s *PriceRuleServiceOp) Create(price PriceRule) (*PriceRule, error) {
	path := fmt.Sprintf("%s.json", priceRuleBasePath)
	wrappedData := PriceRuleResource{PriceRule: &price}
	resource := new(PriceRuleResource)
	err := s.client.Post(path, wrappedData, resource)
	return resource.PriceRule, err
}

// Update an existing  price rule
func (s *PriceRuleServiceOp) Update(price PriceRule) (*PriceRule, error) {
	path := fmt.Sprintf("%s/%d.json", priceRuleBasePath, price.ID)
	wrappedData := PriceRuleResource{PriceRule: &price}
	resource := new(PriceRuleResource)
	err := s.client.Put(path, wrappedData, resource)
	return resource.PriceRule, err
}

// List of discount codes
func (s *PriceRuleServiceOp) List(options interface{}) ([]PriceRule, error) {
	path := fmt.Sprintf("%s.json", priceRuleBasePath)
	resource := new(PriceRulesResource)
	err := s.client.Get(path, resource, options)
	return resource.PriceRule, err
}

// Get a single discount code
func (s *PriceRuleServiceOp) Get(price PriceRule) (*PriceRule, error) {
	path := fmt.Sprintf("%s/%v.json", priceRuleBasePath, price.ID)
	resource := new(PriceRuleResource)
	err := s.client.Get(path, resource, nil)
	return resource.PriceRule, err
}

// Delete an existing price_rule
func (s *PriceRuleServiceOp) Delete(priceRuleID int64) error {
	return s.client.Delete(fmt.Sprintf("%s/%d.json", priceRuleBasePath, priceRuleID))
}

// Count products
func (s *PriceRuleServiceOp) Count(options interface{}) (int, error) {
	path := fmt.Sprintf("%s/count.json", priceRuleBasePath)
	return s.client.Count(path, options)
}
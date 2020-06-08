package goshopify

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const customersBasePath = "customers"
const customersResourceName = "customers"

// linkRegex is used to extract pagination links from customer search results.
var linkPaginationRegex = regexp.MustCompile(`^ *<([^>]+)>; rel="(previous|next)" *$`)

// CustomerService is an interface for interfacing with the customers endpoints
// of the Shopify API.
// See: https://help.shopify.com/api/reference/customer
type CustomerService interface {
	List(interface{}) ([]Customer, error)
	ListWithPagination(interface{}) ([]Customer, *Pagination, error)
	Count(interface{}) (int, error)
	Get(int64, interface{}) (*Customer, error)
	Search(interface{}) ([]Customer, error)
	Create(Customer) (*Customer, error)
	Update(Customer) (*Customer, error)
	Delete(int64) error
	ListOrders(int64, interface{}) ([]Order, error)
	ListTags(interface{}) ([]string, error)

	// MetafieldsService used for Customer resource to communicate with Metafields resource
	MetafieldsService
}

// CustomerServiceOp handles communication with the product related methods of
// the Shopify API.
type CustomerServiceOp struct {
	client *Client
}

// Customer represents a Shopify customer
type Customer struct {
	ID                  int64              `json:"id,omitempty"`
	Email               string             `json:"email,omitempty"`
	FirstName           string             `json:"first_name,omitempty"`
	LastName            string             `json:"last_name,omitempty"`
	State               string             `json:"state,omitempty"`
	Note                string             `json:"note,omitempty"`
	VerifiedEmail       bool               `json:"verified_email,omitempty"`
	MultipassIdentifier string             `json:"multipass_identifier,omitempty"`
	OrdersCount         int                `json:"orders_count,omitempty"`
	TaxExempt           bool               `json:"tax_exempt,omitempty"`
	TotalSpent          *decimal.Decimal   `json:"total_spent,omitempty"`
	Phone               string             `json:"phone,omitempty"`
	Tags                string             `json:"tags,omitempty"`
	LastOrderId         int64              `json:"last_order_id,omitempty"`
	LastOrderName       string             `json:"last_order_name,omitempty"`
	AcceptsMarketing    bool               `json:"accepts_marketing,omitempty"`
	DefaultAddress      *CustomerAddress   `json:"default_address,omitempty"`
	Addresses           []*CustomerAddress `json:"addresses,omitempty"`
	CreatedAt           *time.Time         `json:"created_at,omitempty"`
	UpdatedAt           *time.Time         `json:"updated_at,omitempty"`
	Metafields          []Metafield        `json:"metafields,omitempty"`
}

// Represents the result from the customers/X.json endpoint
type CustomerResource struct {
	Customer *Customer `json:"customer"`
}

// Represents the result from the customers.json endpoint
type CustomersResource struct {
	Customers []Customer `json:"customers"`
}

// Represents the result from the customers/tags.json endpoint
type CustomerTagsResource struct {
	Tags []string `json:"tags"`
}

// Represents the options available when searching for a customer
type CustomerSearchOptions struct {
	Page   int    `url:"page,omitempty"`
	Limit  int    `url:"limit,omitempty"`
	Fields string `url:"fields,omitempty"`
	Order  string `url:"order,omitempty"`
	Query  string `url:"query,omitempty"`
}

// List customers
func (s *CustomerServiceOp) List(options interface{}) ([]Customer, error) {
	path := fmt.Sprintf("%s.json", customersBasePath)
	resource := new(CustomersResource)
	err := s.client.Get(path, resource, options)
	return resource.Customers, err
}

// ListWithPagination lists customers and return pagination to retrieve next/previous results.
func (s *CustomerServiceOp) ListWithPagination(options interface{}) ([]Customer, *Pagination, error) {
	path := fmt.Sprintf("%s.json", customersBasePath)
	resource := new(CustomersResource)
	headers := http.Header{}

	headers, err := s.client.createAndDoGetHeaders("GET", path, nil, options, resource)
	if err != nil {
		return nil, nil, err
	}

	// Extract pagination info from header
	linkHeader := headers.Get("Link")

	pagination, err := extractPagination(linkHeader)
	if err != nil {
		return nil, nil, err
	}

	return resource.Customers, pagination, nil
}

// extractPagination extracts pagination info from linkHeader.
// Details on the format are here:
// https://help.shopify.com/en/api/guides/paginated-rest-results
func extractPagination(linkHeader string) (*Pagination, error) {
	pagination := new(Pagination)

	if linkHeader == "" {
		return pagination, nil
	}

	for _, link := range strings.Split(linkHeader, ",") {
		match := linkPaginationRegex.FindStringSubmatch(link)
		// Make sure the link is not empty or invalid
		if len(match) != 3 {
			// We expect 3 values:
			// match[0] = full match
			// match[1] is the URL and match[2] is either 'previous' or 'next'
			err := ResponseDecodingError{
				Message: "could not extract pagination link header",
			}
			return nil, err
		}

		rel, err := url.Parse(match[1])
		if err != nil {
			err = ResponseDecodingError{
				Message: "pagination does not contain a valid URL",
			}
			return nil, err
		}

		params, err := url.ParseQuery(rel.RawQuery)
		if err != nil {
			return nil, err
		}

		paginationListOptions := ListOptions{}

		paginationListOptions.PageInfo = params.Get("page_info")
		if paginationListOptions.PageInfo == "" {
			err = ResponseDecodingError{
				Message: "page_info is missing",
			}
			return nil, err
		}

		limit := params.Get("limit")
		if limit != "" {
			paginationListOptions.Limit, err = strconv.Atoi(params.Get("limit"))
			if err != nil {
				return nil, err
			}
		}

		// 'rel' is either next or previous
		if match[2] == "next" {
			pagination.NextPageOptions = &paginationListOptions
		} else {
			pagination.PreviousPageOptions = &paginationListOptions
		}
	}

	return pagination, nil
}

// Count customers
func (s *CustomerServiceOp) Count(options interface{}) (int, error) {
	path := fmt.Sprintf("%s/count.json", customersBasePath)
	return s.client.Count(path, options)
}

// Get customer
func (s *CustomerServiceOp) Get(customerID int64, options interface{}) (*Customer, error) {
	path := fmt.Sprintf("%s/%v.json", customersBasePath, customerID)
	resource := new(CustomerResource)
	err := s.client.Get(path, resource, options)
	return resource.Customer, err
}

// Create a new customer
func (s *CustomerServiceOp) Create(customer Customer) (*Customer, error) {
	path := fmt.Sprintf("%s.json", customersBasePath)
	wrappedData := CustomerResource{Customer: &customer}
	resource := new(CustomerResource)
	err := s.client.Post(path, wrappedData, resource)
	return resource.Customer, err
}

// Update an existing customer
func (s *CustomerServiceOp) Update(customer Customer) (*Customer, error) {
	path := fmt.Sprintf("%s/%d.json", customersBasePath, customer.ID)
	wrappedData := CustomerResource{Customer: &customer}
	resource := new(CustomerResource)
	err := s.client.Put(path, wrappedData, resource)
	return resource.Customer, err
}

// Delete an existing customer
func (s *CustomerServiceOp) Delete(customerID int64) error {
	path := fmt.Sprintf("%s/%d.json", customersBasePath, customerID)
	return s.client.Delete(path)
}

// Search customers
func (s *CustomerServiceOp) Search(options interface{}) ([]Customer, error) {
	path := fmt.Sprintf("%s/search.json", customersBasePath)
	resource := new(CustomersResource)
	err := s.client.Get(path, resource, options)
	return resource.Customers, err
}

// ListOrders retrieves all orders from a customer
func (s *CustomerServiceOp) ListOrders(customerID int64, options interface{}) ([]Order, error) {
	path := fmt.Sprintf("%s/%d/orders.json", customersBasePath, customerID)
	resource := new(OrdersResource)
	err := s.client.Get(path, resource, options)
	return resource.Orders, err
}

// ListTags retrieves all unique tags across all customers
func (s *CustomerServiceOp) ListTags(options interface{}) ([]string, error) {
	path := fmt.Sprintf("%s/tags.json", customersBasePath)
	resource := new(CustomerTagsResource)
	err := s.client.Get(path, resource, options)
	return resource.Tags, err
}

// List metafields for a customer
func (s *CustomerServiceOp) ListMetafields(customerID int64, options interface{}) ([]Metafield, error) {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.List(options)
}

// Count metafields for a customer
func (s *CustomerServiceOp) CountMetafields(customerID int64, options interface{}) (int, error) {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.Count(options)
}

// Get individual metafield for a customer
func (s *CustomerServiceOp) GetMetafield(customerID int64, metafieldID int64, options interface{}) (*Metafield, error) {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.Get(metafieldID, options)
}

// Create a new metafield for a customer
func (s *CustomerServiceOp) CreateMetafield(customerID int64, metafield Metafield) (*Metafield, error) {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.Create(metafield)
}

// Update an existing metafield for a customer
func (s *CustomerServiceOp) UpdateMetafield(customerID int64, metafield Metafield) (*Metafield, error) {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.Update(metafield)
}

// // Delete an existing metafield for a customer
func (s *CustomerServiceOp) DeleteMetafield(customerID int64, metafieldID int64) error {
	metafieldService := &MetafieldServiceOp{client: s.client, resource: customersResourceName, resourceID: customerID}
	return metafieldService.Delete(metafieldID)
}

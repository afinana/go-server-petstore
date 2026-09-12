package petstore

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// MaxBodyBytes limits request body size to 1MB to prevent DoS via massive payloads
	MaxBodyBytes = 1 << 20

	// MaxAdoptionsPerUser defines the maximum number of pets a user can adopt
	MaxAdoptionsPerUser = 3

	// Validation constraints
	MaxNameLength     = 100
	MaxURLLength      = 2048
	MaxPhotoURLs      = 20
	MaxTags           = 20
	MaxTagLength      = 50
	MaxCategoryLength = 50
	MaxUsernameLength = 50
	MinUsernameLength = 3
)

var (
	validPetStatuses   = map[string]bool{"available": true, "pending": true, "sold": true}
	validOrderStatuses = map[string]bool{"placed": true, "approved": true, "delivered": true, "cancelled": true}
	usernameRegex      = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// ValidatePet performs deep input validation on a Pet object
func ValidatePet(p *Pet) error {
	trimmedName := strings.TrimSpace(p.Name)
	if trimmedName == "" {
		return errors.New("pet name is required and cannot be empty or whitespace")
	}
	if len(trimmedName) > MaxNameLength {
		return fmt.Errorf("pet name cannot exceed %d characters", MaxNameLength)
	}
	// Disallow ASCII control characters in pet name
	for _, c := range trimmedName {
		if c < 32 {
			return errors.New("pet name contains invalid control characters")
		}
	}
	p.Name = trimmedName

	// Validate status
	if p.Status != "" {
		p.Status = strings.ToLower(strings.TrimSpace(p.Status))
		if !validPetStatuses[p.Status] {
			return fmt.Errorf("invalid pet status %q; allowed values: available, pending, sold", p.Status)
		}
	} else {
		p.Status = "available"
	}

	// Validate Category
	if p.Category != nil {
		p.Category.Name = strings.TrimSpace(p.Category.Name)
		if len(p.Category.Name) > MaxCategoryLength {
			return fmt.Errorf("category name cannot exceed %d characters", MaxCategoryLength)
		}
	}

	// Validate PhotoUrls
	if len(p.PhotoUrls) > MaxPhotoURLs {
		return fmt.Errorf("cannot provide more than %d photo URLs", MaxPhotoURLs)
	}
	for i, u := range p.PhotoUrls {
		u = strings.TrimSpace(u)
		if len(u) > MaxURLLength {
			return fmt.Errorf("photo URL at index %d exceeds maximum length of %d", i, MaxURLLength)
		}
		if u != "" {
			parsedURL, err := url.ParseRequestURI(u)
			if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				return fmt.Errorf("photo URL at index %d must be a valid http or https URL", i)
			}
		}
		p.PhotoUrls[i] = u
	}

	// Validate Tags
	if len(p.Tags) > MaxTags {
		return fmt.Errorf("cannot provide more than %d tags", MaxTags)
	}
	for i, tag := range p.Tags {
		tagName := strings.TrimSpace(tag.Name)
		if len(tagName) > MaxTagLength {
			return fmt.Errorf("tag name at index %d exceeds maximum length of %d", i, MaxTagLength)
		}
		p.Tags[i].Name = tagName
	}

	return nil
}

// ValidateOrder validates order payload and parameters
func ValidateOrder(o *Order) error {
	if o.PetId <= 0 {
		return errors.New("petId must be a positive integer")
	}

	if o.Quantity <= 0 {
		return errors.New("quantity must be at least 1")
	}
	if o.Quantity > int32(MaxAdoptionsPerUser) {
		return fmt.Errorf("quantity cannot exceed the maximum adoption limit of %d pets per order", MaxAdoptionsPerUser)
	}

	if o.Status != "" {
		o.Status = strings.ToLower(strings.TrimSpace(o.Status))
		if !validOrderStatuses[o.Status] {
			return fmt.Errorf("invalid order status %q; allowed: placed, approved, delivered, cancelled", o.Status)
		}
	} else {
		o.Status = "placed"
	}

	return nil
}

// ValidateID checks if an ID string is a valid positive int64 or 24-char hex ObjectID
func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("resource ID cannot be empty")
	}

	// Check if valid numeric ID
	if num, err := strconv.ParseInt(id, 10, 64); err == nil && num > 0 {
		return nil
	}

	// Check if valid MongoDB ObjectID
	if _, err := primitive.ObjectIDFromHex(id); err == nil {
		return nil
	}

	return fmt.Errorf("invalid ID format %q: must be a positive integer or 24-character hexadecimal ObjectID", id)
}

// LimitRequestBody wraps request body with MaxBytesReader
func LimitRequestBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
}

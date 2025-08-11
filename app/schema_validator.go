package app

import (
	"fmt"
	"reflect"
)

// EventSchema defines the schema for analytics events
type EventSchema struct {
	RequiredFields []string
	FieldTypes     map[string]string
	CustomRules    map[string]ValidationRule
}

// ValidationRule defines a custom validation rule
type ValidationRule func(value interface{}) error

// SchemaValidator handles event schema validation
type SchemaValidator struct {
	schemas map[string]*EventSchema
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	validator := &SchemaValidator{
		schemas: make(map[string]*EventSchema),
	}

	// Register default schemas
	validator.registerDefaultSchemas()

	return validator
}

// registerDefaultSchemas registers default event schemas
func (s *SchemaValidator) registerDefaultSchemas() {
	// Page view event schema
	s.RegisterSchema("page_view", &EventSchema{
		RequiredFields: []string{"event_type", "user_id"},
		FieldTypes: map[string]string{
			"event_type": "string",
			"user_id":    "string",
			"page":       "string",
			"properties": "map",
			"session_id": "string",
			"ip_address": "string",
		},
		CustomRules: map[string]ValidationRule{
			"event_type": func(value interface{}) error {
				if str, ok := value.(string); ok {
					if str != "page_view" {
						return fmt.Errorf("event_type must be 'page_view' for page_view events")
					}
				}
				return nil
			},
		},
	})

	// Conversion event schema
	s.RegisterSchema("conversion", &EventSchema{
		RequiredFields: []string{"event_type", "user_id"},
		FieldTypes: map[string]string{
			"event_type": "string",
			"user_id":    "string",
			"page":       "string",
			"properties": "map",
			"amount":     "float64",
			"currency":   "string",
		},
		CustomRules: map[string]ValidationRule{
			"event_type": func(value interface{}) error {
				if str, ok := value.(string); ok {
					if str != "conversion" {
						return fmt.Errorf("event_type must be 'conversion' for conversion events")
					}
				}
				return nil
			},
		},
	})

	// Generic event schema
	s.RegisterSchema("generic", &EventSchema{
		RequiredFields: []string{"event_type", "user_id"},
		FieldTypes: map[string]string{
			"event_type": "string",
			"user_id":    "string",
			"page":       "string",
			"properties": "map",
			"session_id": "string",
			"ip_address": "string",
		},
	})
}

// RegisterSchema registers a new event schema
func (s *SchemaValidator) RegisterSchema(eventType string, schema *EventSchema) {
	s.schemas[eventType] = schema
}

// ValidateEvent validates an event against its schema
func (s *SchemaValidator) ValidateEvent(eventData map[string]interface{}) error {
	// Determine event type
	eventType, ok := eventData["event_type"].(string)
	if !ok {
		return fmt.Errorf("event_type is required and must be a string")
	}

	// Get schema for this event type
	schema, exists := s.schemas[eventType]
	if !exists {
		// Use generic schema if specific schema doesn't exist
		schema = s.schemas["generic"]
	}

	if schema == nil {
		return fmt.Errorf("no schema found for event type: %s", eventType)
	}

	// Validate required fields
	if err := s.validateRequiredFields(eventData, schema.RequiredFields); err != nil {
		return err
	}

	// Validate field types
	if err := s.validateFieldTypes(eventData, schema.FieldTypes); err != nil {
		return err
	}

	// Apply custom validation rules
	if err := s.applyCustomRules(eventData, schema.CustomRules); err != nil {
		return err
	}

	return nil
}

// validateRequiredFields checks that all required fields are present
func (s *SchemaValidator) validateRequiredFields(eventData map[string]interface{}, requiredFields []string) error {
	for _, field := range requiredFields {
		if _, exists := eventData[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// validateFieldTypes checks that fields have the correct types
func (s *SchemaValidator) validateFieldTypes(eventData map[string]interface{}, fieldTypes map[string]string) error {
	for field, expectedType := range fieldTypes {
		if value, exists := eventData[field]; exists {
			if err := s.validateFieldType(field, value, expectedType); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateFieldType validates a single field's type
func (s *SchemaValidator) validateFieldType(fieldName string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field '%s' must be a string, got %s", fieldName, reflect.TypeOf(value))
		}
	case "float64":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("field '%s' must be a float64, got %s", fieldName, reflect.TypeOf(value))
		}
	case "map":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("field '%s' must be a map, got %s", fieldName, reflect.TypeOf(value))
		}
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return fmt.Errorf("field '%s' must be an array, got %s", fieldName, reflect.TypeOf(value))
		}
	default:
		// For unknown types, just check if the value is not nil
		if value == nil {
			return fmt.Errorf("field '%s' cannot be nil", fieldName)
		}
	}
	return nil
}

// applyCustomRules applies custom validation rules
func (s *SchemaValidator) applyCustomRules(eventData map[string]interface{}, customRules map[string]ValidationRule) error {
	for field, rule := range customRules {
		if value, exists := eventData[field]; exists {
			if err := rule(value); err != nil {
				return fmt.Errorf("custom validation failed for field '%s': %w", field, err)
			}
		}
	}
	return nil
}

// GetValidationErrors returns detailed validation errors
func (s *SchemaValidator) GetValidationErrors(eventData map[string]interface{}) []string {
	var errors []string

	// Check required fields
	eventType, ok := eventData["event_type"].(string)
	if !ok {
		errors = append(errors, "event_type is required and must be a string")
		return errors
	}

	schema, exists := s.schemas[eventType]
	if !exists {
		schema = s.schemas["generic"]
	}

	if schema == nil {
		errors = append(errors, fmt.Sprintf("no schema found for event type: %s", eventType))
		return errors
	}

	// Check required fields
	for _, field := range schema.RequiredFields {
		if _, exists := eventData[field]; !exists {
			errors = append(errors, fmt.Sprintf("required field '%s' is missing", field))
		}
	}

	// Check field types
	for field, expectedType := range schema.FieldTypes {
		if value, exists := eventData[field]; exists {
			if err := s.validateFieldType(field, value, expectedType); err != nil {
				errors = append(errors, err.Error())
			}
		}
	}

	return errors
}

package protoxls

import (
	"fmt"
	"strconv"
	"strings"
)

// CellType represents the expected data type for validation
type CellType string

const (
	CellTypeNumber CellType = "number"
	CellTypeBool   CellType = "bool"
	CellTypeString CellType = "string"
)

// ValidateCellType checks if cell value matches expected type
func ValidateCellType(cellValue string, expectedType CellType) bool {
	switch expectedType {
	case CellTypeNumber:
		_, err := strconv.ParseFloat(cellValue, 64)
		return err == nil
	case CellTypeBool:
		return isBooleanValue(cellValue)
	case CellTypeString:
		return true // string类型不校验
	default:
		return false
	}
}

// isBooleanValue checks if the string represents a valid boolean value
func isBooleanValue(value string) bool {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	validValues := []string{"1", "0", "true", "false"}

	for _, valid := range validValues {
		if normalizedValue == valid {
			return true
		}
	}
	return false
}

// ParseNumber safely parses a string to number
func ParseNumber(value string) (interface{}, error) {
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		// Check if it's actually an integer
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal, nil
		}
		return floatVal, nil
	}
	return nil, fmt.Errorf("invalid number format: %s", value)
}

// ParseBoolean safely parses a string to boolean
func ParseBoolean(value string) (bool, error) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	switch normalizedValue {
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean format: %s", value)
	}
}

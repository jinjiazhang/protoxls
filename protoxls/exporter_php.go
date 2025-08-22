package protoxls

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// PhpExporter exports data to PHP format
type PhpExporter struct {
	OutputDir     string
	CompactFormat bool // Whether to compress each data entry to a single line
}

// ExportResult exports the table store data to a PHP file
func (e *PhpExporter) ExportResult(store *TableStore) error {
	// Create output file using the common helper
	file, err := CreateOutputFile(store, e.OutputDir, "php")
	if err != nil {
		return fmt.Errorf("failed to create PHP file: %v", err)
	}
	defer file.Close()

	// Get table name for variable naming
	tableName := GetTableName(store)
	variableName := strings.ToLower(tableName)

	// Write PHP opening tag and variable declaration
	_, err = file.WriteString("<?php\n\n")
	if err != nil {
		return fmt.Errorf("failed to write PHP opening tag: %v", err)
	}

	// Export data to PHP format as a complete array with each key-value pair on one line
	if store.HasChildStores() {
		// Export as array structure with formatted output
		keys := store.GetAllKeys()

		// Write variable name and opening bracket
		if _, err := file.WriteString(fmt.Sprintf("$%s = [\n", variableName)); err != nil {
			return fmt.Errorf("failed to write variable declaration: %v", err)
		}

		for i, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				keyStr := e.formatPhpKey(key)
				childCode := e.generatePhpCode(childStore, 0)

				// Write key-value pair with proper formatting (each pair on one line)
				lineCode := fmt.Sprintf("    %s => %s", keyStr, childCode)
				if i < len(keys)-1 {
					lineCode += ","
				}
				lineCode += "\n"

				if _, err := file.WriteString(lineCode); err != nil {
					return fmt.Errorf("failed to write PHP code: %v", err)
				}
			}
		}

		// Write closing bracket
		if _, err := file.WriteString("];"); err != nil {
			return fmt.Errorf("failed to write closing bracket: %v", err)
		}
	} else {
		// Export each message as one line in an array
		messages := store.GetAllMessages()

		// Write variable name and opening bracket
		if _, err := file.WriteString(fmt.Sprintf("$%s = [\n", variableName)); err != nil {
			return fmt.Errorf("failed to write variable declaration: %v", err)
		}

		for i, message := range messages {
			messageCode := e.generatePhpMessage(message, 0)

			// Write this message as one line with proper formatting
			lineCode := fmt.Sprintf("    %s", messageCode)
			if i < len(messages)-1 {
				lineCode += ","
			}
			lineCode += "\n"

			if _, err := file.WriteString(lineCode); err != nil {
				return fmt.Errorf("failed to write PHP code: %v", err)
			}
		}

		// Write closing bracket
		if _, err := file.WriteString("];"); err != nil {
			return fmt.Errorf("failed to write closing bracket: %v", err)
		}
	}

	return nil
}

// generatePhpCode generates PHP code for the configuration data
func (e *PhpExporter) generatePhpCode(store *TableStore, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	var result strings.Builder

	if store.HasChildStores() {
		if e.CompactFormat {
			// Compact format: all on one line
			result.WriteString("[")
			keys := store.GetAllKeys()
			for i, key := range keys {
				childStore := store.GetChildStore(key)
				if childStore != nil {
					result.WriteString(fmt.Sprintf("%s => ", e.formatPhpKey(key)))
					result.WriteString(e.generatePhpCode(childStore, indentLevel+1))
					if i < len(keys)-1 {
						result.WriteString(", ")
					}
				}
			}
			result.WriteString("]")
		} else {
			// Multi-line format
			result.WriteString("[\n")
			keys := store.GetAllKeys()
			for i, key := range keys {
				childStore := store.GetChildStore(key)
				if childStore != nil {
					result.WriteString(fmt.Sprintf("%s    %s => ", indent, e.formatPhpKey(key)))
					result.WriteString(e.generatePhpCode(childStore, indentLevel+1))
					if i < len(keys)-1 {
						result.WriteString(",")
					}
					result.WriteString("\n")
				}
			}
			result.WriteString(fmt.Sprintf("%s]", indent))
		}
	} else {
		message := store.GetFirstMessage()
		if message != nil {
			result.WriteString(e.generatePhpMessage(message, indentLevel))
		} else {
			result.WriteString("[]")
		}
	}

	return result.String()
}

// generatePhpMessage generates PHP code for a protobuf message with consistent field order
func (e *PhpExporter) generatePhpMessage(msg *dynamic.Message, indentLevel int) string {
	var result strings.Builder

	result.WriteString("[")

	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()
	
	// Sort fields by field number to maintain proto definition order (consistent with JSON exporter)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].GetNumber() < fields[j].GetNumber()
	})

	fieldCount := 0
	if e.CompactFormat {
		// Compact format: all fields on one line
		for _, field := range fields {
			if fieldCount > 0 {
				result.WriteString(", ")
			}
			value := msg.GetField(field)
			result.WriteString(fmt.Sprintf("'%s' => %s", field.GetName(), e.formatPhpValue(value, field, indentLevel+1)))
			fieldCount++
		}
	} else {
		// Multi-line format: each field on separate line
		indent := strings.Repeat("    ", indentLevel+1)
		result.WriteString("\n")
		for _, field := range fields {
			value := msg.GetField(field)
			result.WriteString(fmt.Sprintf("%s'%s' => %s", indent, field.GetName(), e.formatPhpValue(value, field, indentLevel+1)))
			if fieldCount < len(fields)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
			fieldCount++
		}
		result.WriteString(strings.Repeat("    ", indentLevel))
	}

	result.WriteString("]")
	return result.String()
}

// formatPhpValue formats a field value for PHP output
func (e *PhpExporter) formatPhpValue(value interface{}, field *desc.FieldDescriptor, indentLevel int) string {
	if field.IsRepeated() {
		return e.formatPhpArray(value, field, indentLevel)
	}

	switch field.GetType().String() {
	case "TYPE_STRING":
		// Escape quotes and backslashes in strings
		escaped := strings.ReplaceAll(value.(string), "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "\\'")
		return fmt.Sprintf("'%s'", escaped)
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		return fmt.Sprintf("%d", value.(int32))
	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		return fmt.Sprintf("%d", value.(int64))
	case "TYPE_FLOAT", "TYPE_DOUBLE":
		return fmt.Sprintf("%g", value)
	case "TYPE_BOOL":
		if value.(bool) {
			return "true"
		}
		return "false"
	case "TYPE_MESSAGE":
		if msg, ok := value.(*dynamic.Message); ok {
			return e.generatePhpMessage(msg, indentLevel)
		}
	case "TYPE_ENUM":
		return fmt.Sprintf("%d", value.(int32))
	}

	return "null"
}

// formatPhpArray formats array values for PHP output
func (e *PhpExporter) formatPhpArray(value interface{}, field *desc.FieldDescriptor, indentLevel int) string {
	v, ok := value.([]interface{})
	if !ok {
		return "[]"
	}

	if len(v) == 0 {
		return "[]"
	}

	// Determine element type and format accordingly
	fieldType := field.GetType().String()
	var result strings.Builder

	switch fieldType {
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		// Primitive types: use inline format [val1, val2, val3]
		result.WriteString("[")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int32)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		result.WriteString("[")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int64)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_FLOAT", "TYPE_DOUBLE":
		result.WriteString("[")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%g", item))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_STRING":
		result.WriteString("[")
		for i, item := range v {
			escaped := strings.ReplaceAll(item.(string), "\\", "\\\\")
			escaped = strings.ReplaceAll(escaped, "'", "\\'")
			result.WriteString(fmt.Sprintf("'%s'", escaped))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_BOOL":
		result.WriteString("[")
		for i, item := range v {
			if item.(bool) {
				result.WriteString("true")
			} else {
				result.WriteString("false")
			}
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_ENUM":
		result.WriteString("[")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int32)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")

	case "TYPE_MESSAGE":
		if e.CompactFormat {
			result.WriteString("[")
			for i, item := range v {
				if i > 0 {
					result.WriteString(", ")
				}
				if msg, ok := item.(*dynamic.Message); ok {
					result.WriteString(e.generatePhpMessage(msg, indentLevel+1))
				} else {
					result.WriteString("null")
				}
			}
			result.WriteString("]")
		} else {
			// Multi-line format for message arrays
			indent := strings.Repeat("    ", indentLevel)
			result.WriteString("[\n")
			for i, item := range v {
				result.WriteString(fmt.Sprintf("%s    ", indent))
				if msg, ok := item.(*dynamic.Message); ok {
					result.WriteString(e.generatePhpMessage(msg, indentLevel+1))
				} else {
					result.WriteString("null")
				}
				if i < len(v)-1 {
					result.WriteString(",")
				}
				result.WriteString("\n")
			}
			result.WriteString(fmt.Sprintf("%s]", indent))
		}

	default:
		// Fallback: treat as strings
		result.WriteString("[")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("'%v'", item))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("]")
	}

	return result.String()
}

// formatPhpKey formats a store key for PHP output
func (e *PhpExporter) formatPhpKey(key StoreKey) string {
	if key.KeyType == KeyTypeInteger {
		return fmt.Sprintf("'%d'", key.IntegerValue)
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(key.StringValue, "'", "\\'"))
}
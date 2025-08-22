package protoxls

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// PhpExporter exports data to PHP format
type PhpExporter struct {
	OutputDir string
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

	// Convert store data to interface for PHP export
	data, err := e.exportStoreToInterface(store)
	if err != nil {
		return fmt.Errorf("failed to convert store data: %v", err)
	}

	// Write the data as a PHP array
	phpCode := fmt.Sprintf("$%s = %s;\n", variableName, e.convertToPhpArray(data, 0))
	_, err = file.WriteString(phpCode)
	if err != nil {
		return fmt.Errorf("failed to write PHP data: %v", err)
	}

	return nil
}

// exportStoreToInterface converts TableStore to interface{} for PHP export
func (e *PhpExporter) exportStoreToInterface(store *TableStore) (interface{}, error) {
	if store.HasChildStores() {
		result := make(map[string]interface{})
		keys := store.GetAllKeys()
		for _, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				childData, err := e.exportStoreToInterface(childStore)
				if err != nil {
					return nil, err
				}
				result[key.String()] = childData
			}
		}
		return result, nil
	} else {
		messages := store.GetAllMessages()
		if len(messages) == 1 {
			// Single message, return the message data directly
			return e.convertMessageToMap(messages[0]), nil
		} else {
			// Multiple messages, return as array
			result := make([]interface{}, len(messages))
			for i, message := range messages {
				result[i] = e.convertMessageToMap(message)
			}
			return result, nil
		}
	}
}

// convertMessageToMap converts a dynamic message to map for PHP serialization
func (e *PhpExporter) convertMessageToMap(msg *dynamic.Message) map[string]interface{} {
	result := make(map[string]interface{})
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()

	for _, field := range fields {
		value := msg.GetField(field)
		fieldName := field.GetName()

		if field.IsRepeated() {
			result[fieldName] = e.convertRepeatedFieldValue(value, field)
		} else {
			result[fieldName] = e.convertSingleFieldValue(value, field)
		}
	}

	return result
}

// convertSingleFieldValue converts a single field value for PHP serialization
func (e *PhpExporter) convertSingleFieldValue(value interface{}, field *desc.FieldDescriptor) interface{} {
	if value == nil {
		return nil
	}

	switch field.GetType().String() {
	case "TYPE_MESSAGE":
		if dmsg, ok := value.(*dynamic.Message); ok {
			return e.convertMessageToMap(dmsg)
		}
	case "TYPE_ENUM":
		// Return enum as number
		return value
	default:
		return value
	}

	return value
}

// convertRepeatedFieldValue converts repeated field values for PHP serialization
func (e *PhpExporter) convertRepeatedFieldValue(value interface{}, field *desc.FieldDescriptor) []interface{} {
	// Handle slice values
	switch v := value.(type) {
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = e.convertSingleFieldValue(item, field)
		}
		return result
	default:
		// If it's not a slice, wrap it in a slice
		return []interface{}{e.convertSingleFieldValue(value, field)}
	}
}

// convertToPhpArray converts Go data structures to PHP array syntax
func (e *PhpExporter) convertToPhpArray(data interface{}, indent int) string {
	indentStr := strings.Repeat("    ", indent)
	nextIndentStr := strings.Repeat("    ", indent+1)

	switch v := data.(type) {
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case string:
		// Escape quotes and backslashes in strings
		escaped := strings.ReplaceAll(v, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "'", "\\'")
		return fmt.Sprintf("'%s'", escaped)
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		
		var elements []string
		for _, item := range v {
			elements = append(elements, nextIndentStr+e.convertToPhpArray(item, indent+1))
		}
		
		return fmt.Sprintf("[\n%s\n%s]", strings.Join(elements, ",\n"), indentStr)
	case map[string]interface{}:
		if len(v) == 0 {
			return "[]"
		}
		
		var elements []string
		for key, value := range v {
			phpKey := fmt.Sprintf("'%s'", strings.ReplaceAll(key, "'", "\\'"))
			phpValue := e.convertToPhpArray(value, indent+1)
			elements = append(elements, fmt.Sprintf("%s%s => %s", nextIndentStr, phpKey, phpValue))
		}
		
		return fmt.Sprintf("[\n%s\n%s]", strings.Join(elements, ",\n"), indentStr)
	default:
		// Handle other types using reflection
		rv := reflect.ValueOf(data)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			var elements []string
			for i := 0; i < rv.Len(); i++ {
				elements = append(elements, nextIndentStr+e.convertToPhpArray(rv.Index(i).Interface(), indent+1))
			}
			if len(elements) == 0 {
				return "[]"
			}
			return fmt.Sprintf("[\n%s\n%s]", strings.Join(elements, ",\n"), indentStr)
		case reflect.Map:
			var elements []string
			for _, key := range rv.MapKeys() {
				keyStr := fmt.Sprintf("'%v'", key.Interface())
				value := rv.MapIndex(key).Interface()
				phpValue := e.convertToPhpArray(value, indent+1)
				elements = append(elements, fmt.Sprintf("%s%s => %s", nextIndentStr, keyStr, phpValue))
			}
			if len(elements) == 0 {
				return "[]"
			}
			return fmt.Sprintf("[\n%s\n%s]", strings.Join(elements, ",\n"), indentStr)
		default:
			// Convert to string as fallback
			return fmt.Sprintf("'%v'", data)
		}
	}
}
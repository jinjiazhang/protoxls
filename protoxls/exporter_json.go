package protoxls

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// JsonExporter exports configuration data to JSON format
type JsonExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to JSON format
func (je *JsonExporter) ExportResult(store *TableStore) error {
	// Create output file using shared function
	file, err := CreateOutputFile(store, je.OutputDir, "JSON")
	if err != nil {
		return err
	}
	defer file.Close()

	// Export data to JSON format as a complete object with each key-value pair on one line
	if store.HasChildStores() {
		// Export as map structure with formatted output
		keys := store.GetAllKeys()

		// Write opening brace
		if _, err := file.WriteString("{\n"); err != nil {
			return fmt.Errorf("failed to write opening brace: %v", err)
		}

		for i, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				childData, err := je.exportStoreToInterface(childStore)
				if err != nil {
					return err
				}

				// Marshal the value as compact JSON
				valueBytes, err := json.Marshal(childData)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %v", err)
				}
				formattedValue := string(valueBytes)

				// Write key-value pair with proper formatting (space around colon)
				keyStr := fmt.Sprintf(`    "%s": %s`, key.String(), formattedValue)
				if i < len(keys)-1 {
					keyStr += ","
				}
				keyStr += "\n"

				if _, err := file.WriteString(keyStr); err != nil {
					return fmt.Errorf("failed to write JSON: %v", err)
				}
			}
		}

		// Write closing brace
		if _, err := file.WriteString("}"); err != nil {
			return fmt.Errorf("failed to write closing brace: %v", err)
		}
	} else {
		// Export each message as one line in an array
		messages := store.GetAllMessages()

		// Write opening bracket
		if _, err := file.WriteString("[\n"); err != nil {
			return fmt.Errorf("failed to write opening bracket: %v", err)
		}

		for i, message := range messages {
			// Convert dynamic message to JSON by converting to ordered map
			messageData := je.convertMessageToMap(message)

			// Marshal the message as compact JSON
			jsonBytes, err := json.Marshal(messageData)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %v", err)
			}

			// Write this message as one line with proper formatting
			lineStr := fmt.Sprintf("    %s", string(jsonBytes))
			if i < len(messages)-1 {
				lineStr += ","
			}
			lineStr += "\n"

			if _, err := file.WriteString(lineStr); err != nil {
				return fmt.Errorf("failed to write JSON: %v", err)
			}
		}

		// Write closing bracket
		if _, err := file.WriteString("]"); err != nil {
			return fmt.Errorf("failed to write closing bracket: %v", err)
		}
	}

	return nil
}

func (je *JsonExporter) exportStoreToInterface(store *TableStore) (interface{}, error) {
	if store.HasChildStores() {
		result := make(map[string]interface{})
		keys := store.GetAllKeys()
		for _, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				childData, err := je.exportStoreToInterface(childStore)
				if err != nil {
					return nil, err
				}
				result[key.String()] = childData
			}
		}
		return result, nil
	} else {
		message := store.GetFirstMessage()
		if message != nil {
			return je.convertMessageToMap(message), nil
		}
		return nil, nil
	}
}

// OrderedMap represents an ordered map to maintain field order
type OrderedMap struct {
	Keys   []string
	Values map[string]interface{}
}

// MarshalJSON implements json.Marshaler to maintain order during JSON serialization
func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	var parts []string
	for _, key := range om.Keys {
		value := om.Values[key]
		valueBytes, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		parts = append(parts, fmt.Sprintf(`"%s": %s`, key, string(valueBytes)))
	}
	return []byte("{" + strings.Join(parts, ", ") + "}"), nil
}

// convertMessageToMap converts a dynamic message to ordered map for JSON serialization
func (je *JsonExporter) convertMessageToMap(msg *dynamic.Message) *OrderedMap {
	result := &OrderedMap{
		Keys:   make([]string, 0),
		Values: make(map[string]interface{}),
	}
	descriptor := msg.GetMessageDescriptor()

	// Get fields and sort them by field number to maintain proto definition order
	fields := descriptor.GetFields()
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].GetNumber() < fields[j].GetNumber()
	})

	for _, field := range fields {
		if !msg.HasField(field) {
			continue
		}

		value := msg.GetField(field)
		fieldName := field.GetName()

		if field.IsRepeated() {
			result.Values[fieldName] = je.convertRepeatedFieldValue(value, field)
		} else {
			result.Values[fieldName] = je.convertSingleFieldValue(value, field)
		}
		result.Keys = append(result.Keys, fieldName)
	}

	return result
}

// convertSingleFieldValue converts a single field value for JSON serialization
func (je *JsonExporter) convertSingleFieldValue(value interface{}, field *desc.FieldDescriptor) interface{} {
	if value == nil {
		return nil
	}

	switch field.GetType().String() {
	case "TYPE_MESSAGE":
		if dmsg, ok := value.(*dynamic.Message); ok {
			return je.convertMessageToMap(dmsg)
		}
	case "TYPE_ENUM":
		// Return enum as number
		return value
	default:
		return value
	}

	return value
}

// convertRepeatedFieldValue converts repeated field values for JSON serialization
func (je *JsonExporter) convertRepeatedFieldValue(value interface{}, field *desc.FieldDescriptor) []interface{} {
	// Handle slice values
	switch v := value.(type) {
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = je.convertSingleFieldValue(item, field)
		}
		return result
	default:
		// If it's not a slice, wrap it in a slice
		return []interface{}{je.convertSingleFieldValue(value, field)}
	}
}

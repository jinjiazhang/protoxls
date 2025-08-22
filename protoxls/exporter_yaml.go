package protoxls

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"gopkg.in/yaml.v3"
)

// YamlExporter exports data to YAML format
type YamlExporter struct {
	OutputDir string
}

// ExportResult exports the table store data to a YAML file
func (e *YamlExporter) ExportResult(store *TableStore) error {
	// Create output file using the common helper
	file, err := CreateOutputFile(store, e.OutputDir, "yaml")
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %v", err)
	}
	defer file.Close()

	// Convert store data to interface for YAML export
	data, err := e.exportStoreToInterface(store)
	if err != nil {
		return fmt.Errorf("failed to convert store data: %v", err)
	}

	// Convert data to YAML format
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to YAML: %v", err)
	}

	// Write YAML data to file
	_, err = file.Write(yamlData)
	if err != nil {
		return fmt.Errorf("failed to write YAML data: %v", err)
	}

	return nil
}

// exportStoreToInterface converts TableStore to interface{} for YAML export
func (e *YamlExporter) exportStoreToInterface(store *TableStore) (interface{}, error) {
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

// convertMessageToMap converts a dynamic message to map for YAML serialization
func (e *YamlExporter) convertMessageToMap(msg *dynamic.Message) map[string]interface{} {
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

// convertSingleFieldValue converts a single field value for YAML serialization
func (e *YamlExporter) convertSingleFieldValue(value interface{}, field *desc.FieldDescriptor) interface{} {
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

// convertRepeatedFieldValue converts repeated field values for YAML serialization
func (e *YamlExporter) convertRepeatedFieldValue(value interface{}, field *desc.FieldDescriptor) []interface{} {
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
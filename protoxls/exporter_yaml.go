package protoxls

import (
	"fmt"
	"sort"

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

// convertMessageToMap converts a dynamic message to ordered YAML node for serialization
func (e *YamlExporter) convertMessageToMap(msg *dynamic.Message) *yaml.Node {
	result := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()

	// Sort fields by field number to maintain proto definition order (consistent with JSON and PHP exporters)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].GetNumber() < fields[j].GetNumber()
	})

	for _, field := range fields {
		value := msg.GetField(field)
		fieldName := field.GetName()

		// Add field name as key
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: fieldName,
		}
		result.Content = append(result.Content, keyNode)

		// Add field value
		var fieldValue interface{}
		if field.IsRepeated() {
			fieldValue = e.convertRepeatedFieldValue(value, field)
		} else {
			fieldValue = e.convertSingleFieldValue(value, field)
		}
		
		valueNode := e.convertValueToYamlNode(fieldValue)
		result.Content = append(result.Content, valueNode)
	}

	return result
}

// convertValueToYamlNode converts a value to yaml.Node
func (e *YamlExporter) convertValueToYamlNode(value interface{}) *yaml.Node {
	node := &yaml.Node{}
	node.Encode(value)
	return node
}

// convertSingleFieldValue converts a single field value for YAML serialization
func (e *YamlExporter) convertSingleFieldValue(value interface{}, field *desc.FieldDescriptor) interface{} {
	if value == nil {
		return nil
	}

	switch field.GetType().String() {
	case "TYPE_MESSAGE":
		if dmsg, ok := value.(*dynamic.Message); ok {
			// For nested messages, we need to convert them to regular map for proper YAML encoding
			return e.convertMessageToOrderedMap(dmsg)
		}
	case "TYPE_ENUM":
		// Return enum as number
		return value
	default:
		return value
	}

	return value
}

// convertMessageToOrderedMap converts a dynamic message to a yaml.Node with ordered keys for nested messages
func (e *YamlExporter) convertMessageToOrderedMap(msg *dynamic.Message) *yaml.Node {
	result := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()

	// Sort fields by field number to maintain proto definition order
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].GetNumber() < fields[j].GetNumber()
	})

	for _, field := range fields {
		value := msg.GetField(field)
		fieldName := field.GetName()

		// Add field name as key
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: fieldName,
		}
		result.Content = append(result.Content, keyNode)

		// Add field value
		var fieldValue interface{}
		if field.IsRepeated() {
			fieldValue = e.convertRepeatedFieldValue(value, field)
		} else {
			fieldValue = e.convertSingleFieldValueForNested(value, field)
		}
		
		valueNode := e.convertValueToYamlNode(fieldValue)
		result.Content = append(result.Content, valueNode)
	}

	return result
}

// convertSingleFieldValueForNested converts a single field value for nested YAML serialization
func (e *YamlExporter) convertSingleFieldValueForNested(value interface{}, field *desc.FieldDescriptor) interface{} {
	if value == nil {
		return nil
	}

	switch field.GetType().String() {
	case "TYPE_MESSAGE":
		if dmsg, ok := value.(*dynamic.Message); ok {
			// For nested messages, we need to convert them to ordered yaml.Node
			return e.convertMessageToOrderedMap(dmsg)
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
			if field.GetType().String() == "TYPE_MESSAGE" {
				if dmsg, ok := item.(*dynamic.Message); ok {
					result[i] = e.convertMessageToOrderedMap(dmsg)
				} else {
					result[i] = item
				}
			} else {
				result[i] = e.convertSingleFieldValue(item, field)
			}
		}
		return result
	default:
		// If it's not a slice, wrap it in a slice
		return []interface{}{e.convertSingleFieldValue(value, field)}
	}
}
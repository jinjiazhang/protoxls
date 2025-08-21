package protoxls

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// JsonExporter exports configuration data to JSON format
type JsonExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to JSON format
func (je *JsonExporter) ExportResult(store *DataStore) error {
	descriptor := store.GetMessageDescriptor()
	fileName := fmt.Sprintf("%s.json", strings.ToLower(descriptor.GetName()))
	
	// Use custom output directory or default
	outputDir := je.OutputDir
	if outputDir == "" {
		outputDir = DefaultOutputDir
	}
	
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	filePath := filepath.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create json file: %v", err)
	}
	defer file.Close()

	// Export data to JSON format
	var jsonData interface{}
	
	if store.HasChildStores() {
		// Export as map structure
		result := make(map[string]interface{})
		keys := store.GetAllKeys()
		for _, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				childData, err := je.exportStoreToInterface(childStore)
				if err != nil {
					return err
				}
				result[key.String()] = childData
			}
		}
		jsonData = result
	} else {
		// Export as array of messages
		messages := store.GetAllMessages()
		result := make([]interface{}, 0, len(messages))
		
		for _, message := range messages {
			// Convert dynamic message to JSON by converting to map
			messageData := je.convertMessageToMap(message)
			result = append(result, messageData)
		}
		jsonData = result
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jsonData); err != nil {
		return fmt.Errorf("failed to encode JSON: %v", err)
	}

	fmt.Printf("Exported JSON file: %s\n", filePath)
	return nil
}

func (je *JsonExporter) exportStoreToInterface(store *DataStore) (interface{}, error) {
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

// convertMessageToMap converts a dynamic message to map for JSON serialization
func (je *JsonExporter) convertMessageToMap(msg *dynamic.Message) map[string]interface{} {
	result := make(map[string]interface{})
	descriptor := msg.GetMessageDescriptor()
	
	for _, field := range descriptor.GetFields() {
		if !msg.HasField(field) {
			continue
		}
		
		value := msg.GetField(field)
		fieldName := field.GetName()
		
		if field.IsRepeated() {
			result[fieldName] = je.convertRepeatedFieldValue(value, field)
		} else {
			result[fieldName] = je.convertSingleFieldValue(value, field)
		}
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
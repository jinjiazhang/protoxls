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

// LuaExporter exports configuration data to Lua format
type LuaExporter struct{}

func (le *LuaExporter) ExportResult(store *ConfigStore) error {
	descriptor := store.GetDescriptor()
	fileName := fmt.Sprintf("%s.lua", strings.ToLower(descriptor.GetName()))
	
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	filePath := filepath.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create lua file: %v", err)
	}
	defer file.Close()

	luaCode := le.generateCode(store, 0)
	_, err = file.WriteString(luaCode)
	if err != nil {
		return fmt.Errorf("failed to write lua code: %v", err)
	}

	fmt.Printf("Exported Lua file: %s\n", filePath)
	return nil
}

func (le *LuaExporter) generateCode(store *ConfigStore, layer int) string {
	indent := strings.Repeat("    ", layer)
	var result strings.Builder

	if store.HasChildren() {
		result.WriteString("{\n")
		keys := store.ExportKeys()
		for i, key := range keys {
			childStore := store.GetConfig(key.String())
			if childStore != nil {
				result.WriteString(fmt.Sprintf("%s    [%s] = ", indent, le.formatKey(key)))
				result.WriteString(le.generateCode(childStore, layer+1))
				if i < len(keys)-1 {
					result.WriteString(",")
				}
				result.WriteString("\n")
			}
		}
		result.WriteString(fmt.Sprintf("%s}", indent))
	} else {
		data := store.GetData()
		if data != nil {
			result.WriteString(le.generateMessage(data, layer))
		} else {
			result.WriteString("{}")
		}
	}

	return result.String()
}

func (le *LuaExporter) generateMessage(msg *dynamic.Message, layer int) string {
	indent := strings.Repeat("    ", layer)
	var result strings.Builder
	
	result.WriteString("{\n")
	
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()
	
	for i, field := range fields {
		if msg.HasField(field) {
			value := msg.GetField(field)
			result.WriteString(fmt.Sprintf("%s    %s = ", indent, field.GetName()))
			result.WriteString(le.formatValue(value, field, layer+1))
			
			if i < len(fields)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}
	}
	
	result.WriteString(fmt.Sprintf("%s}", indent))
	return result.String()
}

func (le *LuaExporter) formatValue(value interface{}, field *desc.FieldDescriptor, layer int) string {
	if field.IsRepeated() {
		return le.formatArray(value, field, layer)
	}
	
	switch field.GetType().String() {
	case "TYPE_STRING":
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(value.(string), `"`, `\"`))
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		return fmt.Sprintf("%d", value.(int32))
	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		return fmt.Sprintf("%d", value.(int64))
	case "TYPE_FLOAT", "TYPE_DOUBLE":
		return fmt.Sprintf("%f", value)
	case "TYPE_BOOL":
		if value.(bool) {
			return "true"
		}
		return "false"
	case "TYPE_MESSAGE":
		if msg, ok := value.(*dynamic.Message); ok {
			return le.generateMessage(msg, layer)
		}
	case "TYPE_ENUM":
		return fmt.Sprintf("%d", value.(int32))
	}
	
	return "nil"
}

func (le *LuaExporter) formatArray(value interface{}, field *desc.FieldDescriptor, layer int) string {
	indent := strings.Repeat("    ", layer)
	var result strings.Builder
	
	result.WriteString("{\n")
	
	// Handle slice/array values
	switch v := value.(type) {
	case []interface{}:
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%s    ", indent))
			result.WriteString(le.formatValue(item, field, layer+1))
			if i < len(v)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}
	}
	
	result.WriteString(fmt.Sprintf("%s}", indent))
	return result.String()
}

func (le *LuaExporter) formatKey(key StoreKey) string {
	if key.KeyType == KeyTypeInteger {
		return fmt.Sprintf("%d", key.NumKey)
	}
	return fmt.Sprintf(`"%s"`, key.StrKey)
}

// BinExporter exports configuration data to binary format
type BinExporter struct{}

func (be *BinExporter) ExportResult(store *ConfigStore) error {
	descriptor := store.GetDescriptor()
	fileName := fmt.Sprintf("%s.bin", strings.ToLower(descriptor.GetName()))
	
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	filePath := filepath.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create binary file: %v", err)
	}
	defer file.Close()

	// Export all data messages to binary format
	datas := store.ExportDatas()
	for _, data := range datas {
		bytes, err := data.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal message: %v", err)
		}
		
		// Write message size first (4 bytes)
		size := uint32(len(bytes))
		sizeBytes := []byte{
			byte(size >> 24),
			byte(size >> 16),
			byte(size >> 8),
			byte(size),
		}
		
		if _, err := file.Write(sizeBytes); err != nil {
			return fmt.Errorf("failed to write message size: %v", err)
		}
		
		if _, err := file.Write(bytes); err != nil {
			return fmt.Errorf("failed to write message data: %v", err)
		}
	}

	fmt.Printf("Exported binary file: %s\n", filePath)
	return nil
}

// JsonExporter exports configuration data to JSON format
type JsonExporter struct{}

func (je *JsonExporter) ExportResult(store *ConfigStore) error {
	descriptor := store.GetDescriptor()
	fileName := fmt.Sprintf("%s.json", strings.ToLower(descriptor.GetName()))
	
	// Create output directory if it doesn't exist
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
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
	
	if store.HasChildren() {
		// Export as map structure
		result := make(map[string]interface{})
		keys := store.ExportKeys()
		for _, key := range keys {
			childStore := store.GetConfig(key.String())
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
		datas := store.ExportDatas()
		result := make([]interface{}, 0, len(datas))
		
		for _, data := range datas {
			// Convert dynamic message to JSON by converting to map
			msgData := je.dynamicMessageToMap(data)
			result = append(result, msgData)
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

func (je *JsonExporter) exportStoreToInterface(store *ConfigStore) (interface{}, error) {
	if store.HasChildren() {
		result := make(map[string]interface{})
		keys := store.ExportKeys()
		for _, key := range keys {
			childStore := store.GetConfig(key.String())
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
		data := store.GetData()
		if data != nil {
			return je.dynamicMessageToMap(data), nil
		}
		return nil, nil
	}
}

// Helper function to convert dynamic message to map for JSON serialization
func (je *JsonExporter) dynamicMessageToMap(msg *dynamic.Message) map[string]interface{} {
	result := make(map[string]interface{})
	descriptor := msg.GetMessageDescriptor()
	
	for _, field := range descriptor.GetFields() {
		if !msg.HasField(field) {
			continue
		}
		
		value := msg.GetField(field)
		fieldName := field.GetName()
		
		if field.IsRepeated() {
			result[fieldName] = je.convertRepeatedValue(value, field)
		} else {
			result[fieldName] = je.convertSingleValue(value, field)
		}
	}
	
	return result
}

func (je *JsonExporter) convertSingleValue(value interface{}, field *desc.FieldDescriptor) interface{} {
	if value == nil {
		return nil
	}
	
	switch field.GetType().String() {
	case "TYPE_MESSAGE":
		if dmsg, ok := value.(*dynamic.Message); ok {
			return je.dynamicMessageToMap(dmsg)
		}
	case "TYPE_ENUM":
		// Return enum as number
		return value
	default:
		return value
	}
	
	return value
}

func (je *JsonExporter) convertRepeatedValue(value interface{}, field *desc.FieldDescriptor) []interface{} {
	// Handle slice values
	switch v := value.(type) {
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = je.convertSingleValue(item, field)
		}
		return result
	default:
		// If it's not a slice, wrap it in a slice
		return []interface{}{je.convertSingleValue(value, field)}
	}
}
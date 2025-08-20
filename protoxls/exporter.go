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

const (
	// DefaultOutputDir is the default directory for exported files
	DefaultOutputDir = "output"
	// DefaultFilePermissions for created files and directories
	DefaultFilePermissions = 0755
)

// Exporter defines the interface for exporting configuration data
type Exporter interface {
	ExportResult(store *ConfigStore) error
}

// LuaExporter exports configuration data to Lua format
type LuaExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to Lua format
func (le *LuaExporter) ExportResult(store *ConfigStore) error {
	descriptor := store.GetMessageDescriptor()
	fileName := fmt.Sprintf("%s.lua", strings.ToLower(descriptor.GetName()))
	
	// Use custom output directory or default
	outputDir := le.OutputDir
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
		return fmt.Errorf("failed to create lua file: %v", err)
	}
	defer file.Close()

	luaCode := le.generateLuaCode(store, 0)
	if _, err = file.WriteString(luaCode); err != nil {
		return fmt.Errorf("failed to write lua code: %v", err)
	}

	fmt.Printf("Exported Lua file: %s\n", filePath)
	return nil
}

// generateLuaCode generates Lua code for the configuration data
func (le *LuaExporter) generateLuaCode(store *ConfigStore, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	var result strings.Builder

	if store.HasChildStores() {
		result.WriteString("{\n")
		keys := store.GetAllKeys()
		for i, key := range keys {
			childStore := store.GetChildStore(key.String())
			if childStore != nil {
				result.WriteString(fmt.Sprintf("%s    [%s] = ", indent, le.formatLuaKey(key)))
				result.WriteString(le.generateLuaCode(childStore, indentLevel+1))
				if i < len(keys)-1 {
					result.WriteString(",")
				}
				result.WriteString("\n")
			}
		}
		result.WriteString(fmt.Sprintf("%s}", indent))
	} else {
		message := store.GetFirstMessage()
		if message != nil {
			result.WriteString(le.generateLuaMessage(message, indentLevel))
		} else {
			result.WriteString("{}")
		}
	}

	return result.String()
}

// generateLuaMessage generates Lua code for a protobuf message
func (le *LuaExporter) generateLuaMessage(msg *dynamic.Message, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	var result strings.Builder
	
	result.WriteString("{\n")
	
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()
	
	for i, field := range fields {
		if msg.HasField(field) {
			value := msg.GetField(field)
			result.WriteString(fmt.Sprintf("%s    %s = ", indent, field.GetName()))
			result.WriteString(le.formatLuaValue(value, field, indentLevel+1))
			
			if i < len(fields)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}
	}
	
	result.WriteString(fmt.Sprintf("%s}", indent))
	return result.String()
}

// formatLuaValue formats a field value for Lua output
func (le *LuaExporter) formatLuaValue(value interface{}, field *desc.FieldDescriptor, indentLevel int) string {
	if field.IsRepeated() {
		return le.formatLuaArray(value, field, indentLevel)
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
			return le.generateLuaMessage(msg, indentLevel)
		}
	case "TYPE_ENUM":
		return fmt.Sprintf("%d", value.(int32))
	}
	
	return "nil"
}

// formatLuaArray formats array values for Lua output
func (le *LuaExporter) formatLuaArray(value interface{}, field *desc.FieldDescriptor, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	var result strings.Builder
	
	result.WriteString("{\n")
	
	// Handle slice/array values
	switch v := value.(type) {
	case []interface{}:
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%s    ", indent))
			result.WriteString(le.formatLuaValue(item, field, indentLevel+1))
			if i < len(v)-1 {
				result.WriteString(",")
			}
			result.WriteString("\n")
		}
	}
	
	result.WriteString(fmt.Sprintf("%s}", indent))
	return result.String()
}

// formatLuaKey formats a store key for Lua output
func (le *LuaExporter) formatLuaKey(key StoreKey) string {
	if key.KeyType == KeyTypeInteger {
		return fmt.Sprintf("%d", key.IntegerValue)
	}
	return fmt.Sprintf(`"%s"`, key.StringValue)
}

// BinExporter exports configuration data to binary format
type BinExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to binary format
func (be *BinExporter) ExportResult(store *ConfigStore) error {
	descriptor := store.GetMessageDescriptor()
	fileName := fmt.Sprintf("%s.bin", strings.ToLower(descriptor.GetName()))
	
	// Use custom output directory or default
	outputDir := be.OutputDir
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
		return fmt.Errorf("failed to create binary file: %v", err)
	}
	defer file.Close()

	// Export all data messages to binary format
	messages := store.GetAllMessages()
	for i, message := range messages {
		messageBytes, err := message.Marshal()
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %v", i, err)
		}
		
		// Write message size first (4 bytes, big-endian)
		size := uint32(len(messageBytes))
		sizeBytes := []byte{
			byte(size >> 24),
			byte(size >> 16),
			byte(size >> 8),
			byte(size),
		}
		
		if _, err := file.Write(sizeBytes); err != nil {
			return fmt.Errorf("failed to write message %d size: %v", i, err)
		}
		
		if _, err := file.Write(messageBytes); err != nil {
			return fmt.Errorf("failed to write message %d data: %v", i, err)
		}
	}

	fmt.Printf("Exported binary file: %s\n", filePath)
	return nil
}

// JsonExporter exports configuration data to JSON format
type JsonExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to JSON format
func (je *JsonExporter) ExportResult(store *ConfigStore) error {
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
			childStore := store.GetChildStore(key.String())
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

func (je *JsonExporter) exportStoreToInterface(store *ConfigStore) (interface{}, error) {
	if store.HasChildStores() {
		result := make(map[string]interface{})
		keys := store.GetAllKeys()
		for _, key := range keys {
			childStore := store.GetChildStore(key.String())
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
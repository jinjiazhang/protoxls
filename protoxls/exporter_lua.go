package protoxls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// LuaExporter exports configuration data to Lua format
type LuaExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to Lua format
func (le *LuaExporter) ExportResult(store *TableStore) error {
	fileName := GetExportName(store, ".lua")
	
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

	// Export data to Lua format as a complete table with each key-value pair on one line
	// Get table name using shared function
	tableName := GetTableName(store)
	
	if store.HasChildStores() {
		// Export as table structure with formatted output
		keys := store.GetAllKeys()
		
		// Write table name and opening brace
		if _, err := file.WriteString(fmt.Sprintf("%s = {\n", tableName)); err != nil {
			return fmt.Errorf("failed to write table declaration: %v", err)
		}
		
		for i, key := range keys {
			childStore := store.GetChildStore(key)
			if childStore != nil {
				keyStr := le.formatLuaKey(key)
				childCode := le.generateLuaCode(childStore, 0)
				
				// Write key-value pair with proper formatting
				lineCode := fmt.Sprintf("    [%s] = %s", keyStr, childCode)
				if i < len(keys)-1 {
					lineCode += ","
				}
				lineCode += "\n"
				
				if _, err := file.WriteString(lineCode); err != nil {
					return fmt.Errorf("failed to write lua code: %v", err)
				}
			}
		}
		
		// Write closing brace
		if _, err := file.WriteString("}"); err != nil {
			return fmt.Errorf("failed to write closing brace: %v", err)
		}
	} else {
		// Export each message as one line in a table
		messages := store.GetAllMessages()
		
		// Write table name and opening brace
		if _, err := file.WriteString(fmt.Sprintf("%s = {\n", tableName)); err != nil {
			return fmt.Errorf("failed to write table declaration: %v", err)
		}
		
		for i, message := range messages {
			messageCode := le.generateLuaMessage(message, 0)
			
			// Write this message as one line with proper formatting
			lineCode := fmt.Sprintf("    %s", messageCode)
			if i < len(messages)-1 {
				lineCode += ","
			}
			lineCode += "\n"
			
			if _, err := file.WriteString(lineCode); err != nil {
				return fmt.Errorf("failed to write lua code: %v", err)
			}
		}
		
		// Write closing brace
		if _, err := file.WriteString("}"); err != nil {
			return fmt.Errorf("failed to write closing brace: %v", err)
		}
	}

	fmt.Printf("Exported Lua file: %s\n", filePath)
	return nil
}

// generateLuaCode generates Lua code for the configuration data
func (le *LuaExporter) generateLuaCode(store *TableStore, indentLevel int) string {
	indent := strings.Repeat("    ", indentLevel)
	var result strings.Builder

	if store.HasChildStores() {
		result.WriteString("{\n")
		keys := store.GetAllKeys()
		for i, key := range keys {
			childStore := store.GetChildStore(key)
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
	var result strings.Builder
	
	result.WriteString("{")
	
	descriptor := msg.GetMessageDescriptor()
	fields := descriptor.GetFields()
	fieldCount := 0
	
	for _, field := range fields {
		if msg.HasField(field) {
			if fieldCount > 0 {
				result.WriteString(", ")
			}
			value := msg.GetField(field)
			result.WriteString(fmt.Sprintf("%s = %s", field.GetName(), le.formatLuaValue(value, field, indentLevel+1)))
			fieldCount++
		}
	}
	
	result.WriteString("}")
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
	var result strings.Builder
	
	v, ok := value.([]interface{})
	if !ok {
		return "{}"
	}
	
	if len(v) == 0 {
		return "{}"
	}
	
	// Determine element type and format accordingly
	fieldType := field.GetType().String()
	
	switch fieldType {
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		// Primitive types: use inline format {val1, val2, val3}
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int32)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
		
	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int64)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
		
	case "TYPE_FLOAT", "TYPE_DOUBLE":
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%f", item))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
		
	case "TYPE_STRING":
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf(`"%s"`, strings.ReplaceAll(item.(string), `"`, `\"`)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
		
	case "TYPE_BOOL":
		result.WriteString("{")
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
		result.WriteString("}")
		
	case "TYPE_ENUM":
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf("%d", item.(int32)))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
		
	case "TYPE_MESSAGE":
		result.WriteString("{")
		for i, item := range v {
			if i > 0 {
				result.WriteString(", ")
			}
			if msg, ok := item.(*dynamic.Message); ok {
				result.WriteString(le.generateLuaMessage(msg, indentLevel+1))
			} else {
				result.WriteString("nil")
			}
		}
		result.WriteString("}")
		
	default:
		// Fallback: treat as strings
		result.WriteString("{")
		for i, item := range v {
			result.WriteString(fmt.Sprintf(`"%v"`, item))
			if i < len(v)-1 {
				result.WriteString(", ")
			}
		}
		result.WriteString("}")
	}
	
	return result.String()
}

// formatLuaKey formats a store key for Lua output
func (le *LuaExporter) formatLuaKey(key StoreKey) string {
	if key.KeyType == KeyTypeInteger {
		return fmt.Sprintf("%d", key.IntegerValue)
	}
	return fmt.Sprintf(`"%s"`, key.StringValue)
}


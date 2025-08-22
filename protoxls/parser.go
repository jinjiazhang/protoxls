package protoxls

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultArraySeparator is the default separator for array values in cells
	DefaultArraySeparator = ","
	// ColumnNameSeparator is the separator used between base prefix and column name
	ColumnNameSeparator = "."
)

// buildFieldColumnName builds the column name for a field with optional base prefix
func buildFieldColumnName(field *desc.FieldDescriptor, basePrefix string) string {
	options := field.GetFieldOptions()
	columnName := ""
	if options != nil {
		if ext, ok := proto.GetExtension(options, E_Text).(string); ok && ext != "" {
			columnName = ext
		}
	}
	if columnName == "" {
		columnName = field.GetName()
	}
	if basePrefix != "" {
		return basePrefix + ColumnNameSeparator + columnName
	}
	return columnName
}

// buildArrayElementColumnName builds column name for array elements like "field[1]"
func buildArrayElementColumnName(basePrefix string, index int) string {
	return fmt.Sprintf("%s[%d]", basePrefix, index)
}

// columnExists checks if column exists in header map
func columnExists(headerMap map[string]int, columnName string) bool {
	_, exists := headerMap[columnName]
	return exists
}

// parseFieldValue parses a single field value from Excel row
func parseFieldValue(message *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, rowIndex int, basePrefix string) error {
	columnName := buildFieldColumnName(field, basePrefix)
	colIndex, ok := headerMap[columnName]
	if !ok && field.GetType().String() != "TYPE_MESSAGE" {
		return fmt.Errorf("column not found: %s", columnName)
	}
	if colIndex >= len(row) {
		return fmt.Errorf("column index out of bounds for column %s: index %d >= row length %d", columnName, colIndex, len(row))
	}
	cellValue := row[colIndex]
	if cellValue == "" {
		return nil // 允许为空
	}

	// Validate cell type
	switch field.GetType().String() {
	case "TYPE_DOUBLE", "TYPE_FLOAT", "TYPE_INT32", "TYPE_UINT32", "TYPE_INT64", "TYPE_UINT64":
		if !ValidateCellType(cellValue, CellTypeNumber) {
			return fmt.Errorf("invalid number format at row %d, column %d (%s): %s", rowIndex+2, colIndex+1, columnName, cellValue)
		}
	case "TYPE_BOOL":
		if !ValidateCellType(cellValue, CellTypeBool) {
			return fmt.Errorf("invalid boolean format at row %d, column %d (%s): %s", rowIndex+2, colIndex+1, columnName, cellValue)
		}
	case "TYPE_ENUM", "TYPE_STRING":
		// string/enum类型不需要预校验
	}

	// Convert value to appropriate type
	fieldValue, err := convertCellValue(cellValue, field)
	if err != nil {
		return fmt.Errorf("failed to convert value at row %d, column %d (%s): %v", rowIndex+2, colIndex+1, columnName, err)
	}

	// Handle nested messages
	if field.GetType().String() == "TYPE_MESSAGE" {
		nestedMessage := dynamic.NewMessage(field.GetMessageType())
		if err := parseMessage(nestedMessage, field.GetMessageType(), row, headerMap, rowIndex, columnName); err != nil {
			return err
		}
		fieldValue = nestedMessage
	}

	message.SetField(field, fieldValue)
	return nil
}

// convertCellValue converts cell string value to appropriate Go type based on field type
func convertCellValue(cellValue string, field *desc.FieldDescriptor) (interface{}, error) {
	switch field.GetType().String() {
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		if intVal, err := strconv.ParseInt(cellValue, 10, 32); err == nil {
			return int32(intVal), nil
		}
		return nil, fmt.Errorf("invalid int32 value: %s", cellValue)

	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		if intVal, err := strconv.ParseInt(cellValue, 10, 64); err == nil {
			return intVal, nil
		}
		return nil, fmt.Errorf("invalid int64 value: %s", cellValue)

	case "TYPE_FLOAT":
		if floatVal, err := strconv.ParseFloat(cellValue, 32); err == nil {
			return float32(floatVal), nil
		}
		return nil, fmt.Errorf("invalid float value: %s", cellValue)

	case "TYPE_DOUBLE":
		if floatVal, err := strconv.ParseFloat(cellValue, 64); err == nil {
			return floatVal, nil
		}
		return nil, fmt.Errorf("invalid double value: %s", cellValue)

	case "TYPE_BOOL":
		if boolVal, err := ParseBoolean(cellValue); err == nil {
			return boolVal, nil
		}
		return nil, fmt.Errorf("invalid boolean value: %s", cellValue)

	case "TYPE_ENUM":
		enumValue, err := parseEnumValue(cellValue, field)
		if err != nil {
			return nil, err
		}
		return enumValue, nil

	case "TYPE_STRING":
		return cellValue, nil

	default:
		return cellValue, nil
	}
}

// parseEnumValue parses enum value from string
func parseEnumValue(cellValue string, field *desc.FieldDescriptor) (int32, error) {
	enumDesc := field.GetEnumType()
	for _, enumVal := range enumDesc.GetValues() {
		customName := ""
		if opts := enumVal.GetEnumValueOptions(); opts != nil {
			if ext, ok := proto.GetExtension(opts, E_Alias).(string); ok && ext != "" {
				customName = ext
			}
		}
		if customName == cellValue || enumVal.GetName() == cellValue {
			return enumVal.GetNumber(), nil
		}
	}
	return -1, fmt.Errorf("enum value not found: %s for field %s", cellValue, field.GetName())
}

// parseRepeatedFieldValue parses repeated field values from Excel row
func parseRepeatedFieldValue(message *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, _ int, basePrefix string) error {
	columnName := buildFieldColumnName(field, basePrefix)

	// Try parsing as separator-delimited array first
	if columnExists(headerMap, columnName) {
		if err := parseDelimitedArray(message, field, row, headerMap, columnName); err != nil {
			return err
		}
		return nil
	}

	// Try parsing as indexed columns: name[1], name[2], etc.
	return parseIndexedArray(message, field, row, headerMap, columnName)
}

// parseDelimitedArray parses array values separated by delimiter
func parseDelimitedArray(message *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, columnName string) error {
	colIndex := headerMap[columnName]
	cellValue := row[colIndex]
	if cellValue == "" {
		return nil
	}

	values := strings.Split(cellValue, DefaultArraySeparator)
	for _, val := range values {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}

		convertedValue, err := convertCellValue(val, field)
		if err != nil {
			return fmt.Errorf("failed to convert array element '%s': %v", val, err)
		}
		message.AddRepeatedField(field, convertedValue)
	}
	return nil
}

// parseIndexedArray parses array values from indexed columns
func parseIndexedArray(message *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, baseColumnName string) error {
	for index := 1; ; index++ {
		elementColumnName := buildArrayElementColumnName(baseColumnName, index)
		
		// For message types, check if any sub-field exists for this index
		if field.GetType().String() == "TYPE_MESSAGE" {
			// Check if any field of this message exists with this index
			hasAnyField := false
			for _, subField := range field.GetMessageType().GetFields() {
				subColumnName := elementColumnName + "." + buildFieldColumnName(subField, "")
				if columnExists(headerMap, subColumnName) {
					hasAnyField = true
					break
				}
			}
			if !hasAnyField {
				break
			}
			
			// Create nested message for this index
			nestedMessage := dynamic.NewMessage(field.GetMessageType())
			if err := parseMessage(nestedMessage, field.GetMessageType(), row, headerMap, 0, elementColumnName); err != nil {
				return fmt.Errorf("failed to parse nested message at index %d: %v", index, err)
			}
			message.AddRepeatedField(field, nestedMessage)
		} else {
			// Handle primitive types
			if !columnExists(headerMap, elementColumnName) {
				break
			}

			colIndex := headerMap[elementColumnName]
			cellValue := row[colIndex]
			if cellValue == "" {
				continue
			}

			convertedValue, err := convertCellValue(cellValue, field)
			if err != nil {
				return fmt.Errorf("failed to convert array element '%s' at index %d: %v", cellValue, index, err)
			}
			message.AddRepeatedField(field, convertedValue)
		}
	}
	return nil
}

// parseMessage parses a complete message from Excel row data
func parseMessage(message *dynamic.Message, msgDesc *desc.MessageDescriptor, row []string, headerMap map[string]int, rowIndex int, basePrefix string) error {
	for _, field := range msgDesc.GetFields() {
		if field.IsMap() {
			// TODO: Map fields are not yet supported
			continue
		} else if field.IsRepeated() {
			if err := parseRepeatedFieldValue(message, field, row, headerMap, rowIndex, basePrefix); err != nil {
				return err
			}
		} else {
			if err := parseFieldValue(message, field, row, headerMap, rowIndex, basePrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportConfig holds configuration for different export formats
type ExportConfig struct {
	LuaOutput  string // Output directory for Lua files
	JsonOutput string // Output directory for JSON files
	BinOutput  string // Output directory for Binary files
}

// ParseProtoFiles parses proto files and generates configuration tables with custom export configuration
func ParseProtoFiles(protoFile string, importPaths []string, exportConfig *ExportConfig) error {
	parser := protoparse.Parser{
		ImportPaths: importPaths,
	}

	fileDescriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		return fmt.Errorf("failed to parse proto file %s: %v", protoFile, err)
	}

	var configStores []*TableStore

	for _, fd := range fileDescriptors {
		for _, md := range fd.GetMessageTypes() {
			// Only process messages that have Excel options
			options := md.GetMessageOptions()
			if options == nil {
				continue
			}
			
			// Check if message has excel option
			if _, ok := proto.GetExtension(options, E_Excel).(string); !ok {
				continue
			}
			
			store, err := parseExcelToTableStore(md)
			if err != nil {
				return fmt.Errorf("failed to generate table for message %s: %v", md.GetName(), err)
			}
			configStores = append(configStores, store)
		}
	}

	// Export results to specified formats
	return ExportTableStores(configStores, exportConfig)
}

// parseExcelToTableStore parses Excel file data into TableStore
func parseExcelToTableStore(msgDesc *desc.MessageDescriptor) (*TableStore, error) {
	// Parse table configuration from message options
	options := msgDesc.GetMessageOptions()
	if options == nil {
		return nil, fmt.Errorf("message %s has no options", msgDesc.GetName())
	}

	// Extract Excel file path
	excelPath, ok := proto.GetExtension(options, E_Excel).(string)
	if !ok || excelPath == "" {
		return nil, fmt.Errorf("message %s missing excel option", msgDesc.GetName())
	}

	// Extract sheet name
	sheetName, ok := proto.GetExtension(options, E_Sheet).(string)
	if !ok || sheetName == "" {
		return nil, fmt.Errorf("message %s missing sheet option", msgDesc.GetName())
	}

	// Extract optional key configuration
	keysConfig := ""
	if keys, ok := proto.GetExtension(options, E_Keys).(string); ok {
		keysConfig = keys
	}

	excelFile, err := excelize.OpenFile(excelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file %s: %v", excelPath, err)
	}
	defer excelFile.Close()

	rows, err := excelFile.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet %s: %v", sheetName, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet %s has insufficient data (need at least header + 1 data row)", sheetName)
	}

	// Build header map
	headers := rows[0]
	headerMap := make(map[string]int)
	for index, header := range headers {
		headerMap[header] = index
	}

	// Create config store
	store := NewTableStore(msgDesc)

	// Parse data rows
	messages := make([]*dynamic.Message, 0, len(rows)-1)
	for rowIndex, row := range rows[1:] {
		message := dynamic.NewMessage(msgDesc)
		if err := parseMessage(message, msgDesc, row, headerMap, rowIndex, ""); err != nil {
			return nil, fmt.Errorf("failed to parse row %d: %v", rowIndex+2, err)
		}
		messages = append(messages, message)
	}

	// Import data into store
	store.AddMessages(messages)

	// Build store with keys if specified
	if keysConfig != "" {
		keyNames := strings.Split(keysConfig, ";")
		if err := store.BuildHierarchicalStore(keyNames); err != nil {
			return nil, fmt.Errorf("failed to build store with keys: %v", err)
		}
	}

	log.Printf("Successfully parsed %d rows for message %s", len(messages), msgDesc.GetName())
	return store, nil
}

// ExportTableStores exports configuration stores to specified formats
func ExportTableStores(stores []*TableStore, exportConfig *ExportConfig) error {
	var exporters []Exporter

	// Add exporters based on configuration
	if exportConfig.LuaOutput != "" {
		exporters = append(exporters, &LuaExporter{OutputDir: exportConfig.LuaOutput})
	}
	if exportConfig.JsonOutput != "" {
		exporters = append(exporters, &JsonExporter{OutputDir: exportConfig.JsonOutput})
	}
	if exportConfig.BinOutput != "" {
		exporters = append(exporters, &BinExporter{OutputDir: exportConfig.BinOutput})
	}

	// If no exporters specified, default to JSON
	if len(exporters) == 0 {
		exporters = append(exporters, &JsonExporter{OutputDir: DefaultOutputDir})
	}

	for _, store := range stores {
		for _, exporter := range exporters {
			if err := exporter.ExportResult(store); err != nil {
				log.Printf("Failed to export %T for %s: %v", exporter, store.GetMessageDescriptor().GetName(), err)
			}
		}
	}

	return nil
}

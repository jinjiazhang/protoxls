package protoxls

import (
	"strings"
	"google.golang.org/protobuf/proto"
)

const (
	// DefaultOutputDir is the default directory for exported files
	DefaultOutputDir = "output"
	// DefaultFilePermissions for created files and directories
	DefaultFilePermissions = 0755
)

// Exporter defines the interface for exporting configuration data
type Exporter interface {
	ExportResult(store *TableStore) error
}

// GetExportName returns the filename for export, prioritizing table option
func GetExportName(store *TableStore, extension string) string {
	tableName := GetTableName(store)
	return tableName + extension
}

// GetTableName returns the preferred table name, prioritizing table option
func GetTableName(store *TableStore) string {
	descriptor := store.GetMessageDescriptor()
	options := descriptor.GetMessageOptions()
	
	// Try to get table option first
	if options != nil {
		if tableName, ok := proto.GetExtension(options, E_Table).(string); ok && tableName != "" {
			return tableName
		}
	}
	
	// Fallback to descriptor name
	return strings.ToLower(descriptor.GetName())
}
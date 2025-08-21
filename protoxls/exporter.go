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

// GetPreferredFileName returns the preferred filename for export, prioritizing table option
func GetPreferredFileName(store *TableStore, extension string) string {
	descriptor := store.GetMessageDescriptor()
	options := descriptor.GetMessageOptions()
	
	// Try to get table option first
	if options != nil {
		if tableName, ok := proto.GetExtension(options, E_Table).(string); ok && tableName != "" {
			return tableName + extension
		}
	}
	
	// Fallback to descriptor name
	return strings.ToLower(descriptor.GetName()) + extension
}
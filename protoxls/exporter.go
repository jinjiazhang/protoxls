package protoxls

import (
	"fmt"
	"os"
	"path/filepath"
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

// CreateOutputFile creates an output file with the given parameters and returns the file handle
// This centralizes the common file creation logic used by all exporters
func CreateOutputFile(store *TableStore, outputDir, fileType string) (*os.File, error) {
	// Generate filename based on table name and file type
	tableName := GetTableName(store)
	var extension string
	switch strings.ToLower(fileType) {
	case "json":
		extension = ".json"
	case "lua":
		extension = ".lua"
	case "binary", "bin":
		extension = ".bin"
	default:
		extension = "." + strings.ToLower(fileType)
	}
	fileName := tableName + extension

	// Use default output directory if not specified
	if outputDir == "" {
		outputDir = DefaultOutputDir
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, DefaultFilePermissions); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create the file
	filePath := filepath.Join(outputDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s file: %v", fileType, err)
	}

	// Print export success message
	fmt.Printf("Exported %s file: %s\n", fileType, filePath)

	return file, nil
}

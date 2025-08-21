package protoxls

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BinExporter exports configuration data to binary format
type BinExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to binary format
func (be *BinExporter) ExportResult(store *TableStore) error {
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
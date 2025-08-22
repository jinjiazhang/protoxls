package protoxls

import (
	"fmt"
)

// BinExporter exports configuration data to binary format
type BinExporter struct {
	OutputDir string // Custom output directory, defaults to DefaultOutputDir if empty
}

// ExportResult exports configuration data to binary format
func (be *BinExporter) ExportResult(store *TableStore) error {
	// Create output file using shared function
	file, err := CreateOutputFile(store, be.OutputDir, "binary")
	if err != nil {
		return err
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

	return nil
}

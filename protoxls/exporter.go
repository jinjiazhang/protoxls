package protoxls

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
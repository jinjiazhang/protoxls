package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"protoxls/protoxls"
	"strings"
)

func main() {
	// Proto file and import paths
	protoFilePath := flag.String("proto", "scheme.proto", "Path to the .proto file to parse")
	importPaths := flag.String("I", ".", "Import paths for .proto files (colon-separated)")

	// Output format flags (similar to protoc)
	luaOut := flag.String("lua_out", "", "Generate Lua files in the specified directory")
	jsonOut := flag.String("json_out", "", "Generate JSON files in the specified directory")
	binOut := flag.String("bin_out", "", "Generate binary files in the specified directory")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] -proto <proto_file>\n\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "Protocol buffer configuration table generator.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto                         # Generate JSON files (default)\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -json_out=./output      # Generate JSON files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -lua_out=./output       # Generate Lua files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -lua_out=./output -json_out=./output  # Generate both formats\n", "protoxls")
	}

	flag.Parse()

	// Validate required arguments
	if *protoFilePath == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: -proto flag is required\n\n")
		flag.Usage()
		return
	}

	// Parse import paths
	var parsedImportPaths []string
	if *importPaths != "" {
		if strings.Contains(*importPaths, ":") {
			// Unix-style colon-separated paths
			parsedImportPaths = strings.Split(*importPaths, ":")
		} else {
			// Use system-specific path list separator
			parsedImportPaths = filepath.SplitList(*importPaths)
		}
	} else {
		parsedImportPaths = []string{"."}
	}

	// Configure export options
	exportConfig := &protoxls.ExportConfig{
		LuaOutput:  *luaOut,
		JsonOutput: *jsonOut,
		BinOutput:  *binOut,
	}

	// If no output format specified, default to JSON
	if exportConfig.LuaOutput == "" && exportConfig.JsonOutput == "" && exportConfig.BinOutput == "" {
		exportConfig.JsonOutput = "output"
		fmt.Println("No output format specified, defaulting to JSON output in './output' directory")
	}

	// Parse proto files and generate tables
	err := protoxls.ParseProtoFiles(*protoFilePath, parsedImportPaths, exportConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Print success message
	fmt.Printf("Successfully processed %s\n", *protoFilePath)
}

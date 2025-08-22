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
	yamlOut := flag.String("yaml_out", "", "Generate YAML files in the specified directory")
	phpOut := flag.String("php_out", "", "Generate PHP files in the specified directory")
	allOut := flag.String("all_out", "", "Generate all format files in the specified directory")

	// Format options
	compactFormat := flag.Bool("compact", false, "Compress each data entry to a single line (applies to lua, json, php formats)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] -proto <proto_file>\n\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "Protocol buffer configuration table generator.\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -all_out=./output       # Generate all format files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -json_out=./output      # Generate JSON files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -lua_out=./output       # Generate Lua files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -yaml_out=./output      # Generate YAML files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -php_out=./output       # Generate PHP files in ./output\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -lua_out=./output -json_out=./output  # Generate multiple formats\n", "protoxls")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -proto config.proto -all_out=./output -compact           # Generate all formats compactly\n", "protoxls")
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
		CompactFormat: *compactFormat,
	}

	// Handle all_out option
	if *allOut != "" {
		exportConfig.LuaOutput = *allOut
		exportConfig.JsonOutput = *allOut
		exportConfig.BinOutput = *allOut
		exportConfig.YamlOutput = *allOut
		exportConfig.PhpOutput = *allOut
	} else {
		// Use individual format options
		exportConfig.LuaOutput = *luaOut
		exportConfig.JsonOutput = *jsonOut
		exportConfig.BinOutput = *binOut
		exportConfig.YamlOutput = *yamlOut
		exportConfig.PhpOutput = *phpOut
	}

	// Check if any output format is specified
	if exportConfig.LuaOutput == "" && exportConfig.JsonOutput == "" && exportConfig.BinOutput == "" && exportConfig.YamlOutput == "" && exportConfig.PhpOutput == "" {
		fmt.Fprintf(flag.CommandLine.Output(), "Error: No output format specified. Use one of: -lua_out, -json_out, -bin_out, -yaml_out, -php_out, or -all_out\n\n")
		flag.Usage()
		return
	}

	// Parse proto files and generate tables
	err := protoxls.ParseProtoFiles(*protoFilePath, parsedImportPaths, exportConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Print success message
	fmt.Printf("Successfully processed %s\n", *protoFilePath)
}

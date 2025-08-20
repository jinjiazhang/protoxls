package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"protoxls/protoxls"
)

func main() {
	protoFilePath := flag.String("proto", "", "Path to the .proto file to parse")
	importPaths := flag.String("I", ".", "import paths for .proto files")

	flag.Parse()
	if *protoFilePath == "" {
		fmt.Println("Usage: protoxls -proto <path_to_proto_file> [-I <import_paths>]")
		flag.PrintDefaults()
		return
	}

	var parsedImportPaths []string
	if *importPaths != "" {
		parsedImportPaths = filepath.SplitList(*importPaths)
	} else {
		parsedImportPaths = []string{"."}
	}

	err := protoxls.ParseProtoFiles(*protoFilePath, parsedImportPaths)
	if err != nil {
		log.Fatal(err)
	}
}

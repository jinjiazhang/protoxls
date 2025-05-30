package protoxls

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

func GenerateTables(protoFile string, importPaths []string) error {
	parser := protoparse.Parser{
		ImportPaths: importPaths,
	}

	fileDescriptors, err := parser.ParseFiles(protoFile)
	if err != nil {
		return fmt.Errorf("failed to parse proto, file:%v, err: %v", protoFile, err)
	}

	for _, fd := range fileDescriptors {
		for _, md := range fd.GetMessageTypes() {
			config := parseXlsConfig(md)
			if config == nil {
				continue
			}

			err = GenerateTable(config, md)
			if err != nil {
				return fmt.Errorf("failed to generate table, message:%v, err: %v", md.GetName(), err)
			}
		}
	}

	return nil
}

func GenerateTable(config *XlsConfig, md *desc.MessageDescriptor) error {
	return nil
}

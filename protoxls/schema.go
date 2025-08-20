package protoxls

import (
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/proto"
)

type TableSchema struct {
	Excel  string
	Sheet  string
	Key    string
	Prefix string
	Suffix string
}

func parseTableSchema(msgDesc *desc.MessageDescriptor) *TableSchema {
	options := msgDesc.GetMessageOptions()
	if options == nil {
		return nil
	}

	config := &TableSchema{}
	
	// Safely extract extensions
	if excel, ok := proto.GetExtension(options, E_Excel).(string); ok && excel != "" {
		config.Excel = excel
	} else {
		return nil // Excel is required
	}

	if sheet, ok := proto.GetExtension(options, E_Sheet).(string); ok && sheet != "" {
		config.Sheet = sheet
	} else {
		return nil // Sheet is required
	}

	if key, ok := proto.GetExtension(options, E_Key).(string); ok {
		config.Key = key
	}

	if prefix, ok := proto.GetExtension(options, E_Prefix).(string); ok {
		config.Prefix = prefix
	}

	if suffix, ok := proto.GetExtension(options, E_Suffix).(string); ok {
		config.Suffix = suffix
	}

	return config
}

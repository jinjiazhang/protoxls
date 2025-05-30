package protoxls

import (
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/proto"
)

type XlsConfig struct {
	ExcelName string
	SheetName string
	KeyField  string
	Prefix    string
	Suffix    string
}

func parseXlsConfig(md *desc.MessageDescriptor) *XlsConfig {
	config := &XlsConfig{}
	options := md.GetMessageOptions()
	config.ExcelName = proto.GetExtension(options, E_Excel).(string)
	config.SheetName = proto.GetExtension(options, E_Sheet).(string)
	config.KeyField = proto.GetExtension(options, E_Key).(string)
	config.Prefix = proto.GetExtension(options, E_Prefix).(string)
	config.Suffix = proto.GetExtension(options, E_Suffix).(string)
	return config
}

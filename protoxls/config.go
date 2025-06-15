package protoxls

import (
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/proto"
)

type XlsConfig struct {
	Excel  string
	Sheet  string
	Key    string
	Prefix string
	Suffix string
}

func parseXlsConfig(msgDesc *desc.MessageDescriptor) *XlsConfig {
	config := &XlsConfig{}
	options := msgDesc.GetMessageOptions()
	config.Excel = proto.GetExtension(options, E_Excel).(string)
	config.Sheet = proto.GetExtension(options, E_Sheet).(string)
	config.Key = proto.GetExtension(options, E_Key).(string)
	config.Prefix = proto.GetExtension(options, E_Prefix).(string)
	config.Suffix = proto.GetExtension(options, E_Suffix).(string)
	return config
}

package main

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/protobuf/proto"
)

// ParseScheme 解析pb协议文件
func ParseScheme(file string) (*desc.FileDescriptor, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var fileSet descriptor.FileDescriptorSet
	proto.Unmarshal(bs, &fileSet)

	fileDesc, err := desc.CreateFileDescriptorFromSet(&fileSet)
	if err != nil {
		return nil, err
	}

	return fileDesc, nil
}

// LoadXlsStore 从xls文件中加载数据
func LoadXlsStore(meta *desc.MessageDescriptor) error {
	options := meta.GetMessageOptions()
	optExcel := proto.GetExtension(options, E_Excel).(string)
	optSheet := proto.GetExtension(options, E_Sheet).(string)
	fmt.Println(optExcel, optSheet)
	return nil
}

// SaveLuaTable 保存为lua配置
func SaveLuaTable(configs []*proto.Message) error {
	return nil
}

// SavePhpTable 保存为php配置
func SavePhpTable(configs []*proto.Message) error {
	return nil
}

// SaveYmlTable 保存为yml配置
func SaveYmlTable(configs []*proto.Message) error {
	return nil
}

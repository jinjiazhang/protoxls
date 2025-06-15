package protoxls

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
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

func GenerateTable(config *XlsConfig, msgDesc *desc.MessageDescriptor) error {
	f, err := excelize.OpenFile(config.Excel)
	if err != nil {
		return fmt.Errorf("failed to open excel file: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows(config.Sheet)
	if err != nil {
		return fmt.Errorf("failed to get sheet: %v", err)
	}

	if len(rows) < 2 {
		return fmt.Errorf("sheet %s has no data", config.Sheet)
	}

	headers := rows[0]
	headerMap := make(map[string]int)
	for idx, h := range headers {
		headerMap[h] = idx
	}
	for _, row := range rows[1:] {
		data := dynamic.NewMessage(msgDesc)
		for _, field := range msgDesc.GetFields() {
			if field.IsRepeated() || field.IsMap() {
				continue
			}

			options := field.GetFieldOptions()
			column := proto.GetExtension(options, E_Text).(string)
			if column == "" {
				column = field.GetName()
			}

			colIdx, ok := headerMap[column]
			if !ok || colIdx >= len(row) {
				return fmt.Errorf("excel column missing or out of range: %s", column)
			}
			cellValue := row[colIdx]
			if cellValue == "" {
				return fmt.Errorf("excel column %s row %d is empty", column, colIdx+1)
			}

			var v interface{} = cellValue
			switch field.GetType().String() {
			case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
				var iv int32
				_, err := fmt.Sscanf(cellValue, "%d", &iv)
				if err == nil {
					v = iv
				}
			case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
				var iv int64
				_, err := fmt.Sscanf(cellValue, "%d", &iv)
				if err == nil {
					v = iv
				}
			case "TYPE_FLOAT", "TYPE_DOUBLE":
				var fv float64
				_, err := fmt.Sscanf(cellValue, "%f", &fv)
				if err == nil {
					v = fv
				}
			case "TYPE_BOOL":
				if cellValue == "1" || cellValue == "true" || cellValue == "TRUE" {
					v = true
				} else {
					v = false
				}
			case "TYPE_ENUM", "TYPE_MESSAGE":
				continue
			}

			data.SetField(field, v)
		}
		// TODO: 收集data到datas，后续保存为json
	}
	return nil
}

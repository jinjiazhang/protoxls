package protoxls

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
)

func cellTypeCheck(cell string, expect string) bool {
	if expect == "number" {
		_, err := fmt.Sscanf(cell, "%f", new(float64))
		return err == nil
	}
	if expect == "bool" {
		return cell == "1" || cell == "0" || cell == "true" || cell == "false" || cell == "TRUE" || cell == "FALSE"
	}
	return true // string类型不校验
}

func getFieldText(field *desc.FieldDescriptor, base string) string {
	options := field.GetFieldOptions()
	column := ""
	if options != nil {
		if ext, ok := proto.GetExtension(options, E_Text).(string); ok && ext != "" {
			column = ext
		}
	}
	if column == "" {
		column = field.GetName()
	}
	if base != "" {
		return base + column
	}
	return column
}

func getElementText(base string, idx int) string {
	return fmt.Sprintf("%s[%d]", base, idx)
}

func hasColumn(headerMap map[string]int, name string) bool {
	_, ok := headerMap[name]
	return ok
}

func parseSingleField(data *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, rowIdx int, base string) error {
	textName := getFieldText(field, base)
	colIdx, ok := headerMap[textName]
	if !ok || colIdx >= len(row) {
		return fmt.Errorf("ParseSingle column not found, name=%s", textName)
	}
	cellValue := row[colIdx]
	if cellValue == "" {
		return nil // 允许为空
	}

	switch field.GetType().String() {
	case "TYPE_DOUBLE", "TYPE_FLOAT", "TYPE_INT32", "TYPE_UINT32", "TYPE_INT64", "TYPE_UINT64":
		if !cellTypeCheck(cellValue, "number") {
			return fmt.Errorf("ParseSingle cell type error, expect=number, row=%d, col=%d, text=%s", rowIdx+2, colIdx+1, textName)
		}
	case "TYPE_BOOL":
		if !cellTypeCheck(cellValue, "bool") {
			return fmt.Errorf("ParseSingle cell type error, expect=bool, row=%d, col=%d, text=%s", rowIdx+2, colIdx+1, textName)
		}
	case "TYPE_ENUM", "TYPE_STRING":
		// string/enum不校验
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
	case "TYPE_ENUM":
		enumDesc := field.GetEnumType()
		var enumVal int32 = -1
		for _, v := range enumDesc.GetValues() {
			cname := ""
			if opts := v.GetEnumValueOptions(); opts != nil {
				if ext, ok := proto.GetExtension(opts, E_Cname).(string); ok && ext != "" {
					cname = ext
				}
			}
			if cname == cellValue || v.GetName() == cellValue {
				enumVal = v.GetNumber()
				break
			}
		}
		if enumVal == -1 {
			return fmt.Errorf("enum value not found: %s for field %s", cellValue, field.GetName())
		}
		v = enumVal
	case "TYPE_STRING":
		v = cellValue
	case "TYPE_MESSAGE":
		// 嵌套message递归
		msg := dynamic.NewMessage(field.GetMessageType())
		err := parseMessage(msg, field.GetMessageType(), row, headerMap, rowIdx, textName)
		if err != nil {
			return err
		}
		v = msg
	}
	data.SetField(field, v)
	return nil
}

func parseRepeatedField(data *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, rowIdx int, base string) error {
	textName := getFieldText(field, base)
	// 1. 尝试分隔符数组
	if hasColumn(headerMap, textName) {
		colIdx := headerMap[textName]
		cellValue := row[colIdx]
		if cellValue != "" {
			// 默认分号分隔
			arr := cellValue
			values := strings.Split(arr, ";")
			for _, val := range values {
				val = strings.TrimSpace(val)
				if val == "" {
					continue
				}
				var v interface{} = val
				switch field.GetType().String() {
				case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
					var iv int32
					_, err := fmt.Sscanf(val, "%d", &iv)
					if err == nil {
						v = iv
					} else {
						return fmt.Errorf("repeated int32 parse error: %v, value=%s", err, val)
					}
				case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
					var iv int64
					_, err := fmt.Sscanf(val, "%d", &iv)
					if err == nil {
						v = iv
					} else {
						return fmt.Errorf("repeated int64 parse error: %v, value=%s", err, val)
					}
				case "TYPE_FLOAT", "TYPE_DOUBLE":
					var fv float64
					_, err := fmt.Sscanf(val, "%f", &fv)
					if err == nil {
						v = fv
					} else {
						return fmt.Errorf("repeated float parse error: %v, value=%s", err, val)
					}
				case "TYPE_BOOL":
					if val == "1" || val == "true" || val == "TRUE" {
						v = true
					} else {
						v = false
					}
				}
				data.AddRepeatedField(field, v)
			}
			return nil
		}
	}
	// 2. 多列 name[1], name[2] ...
	for idx := 1; ; idx++ {
		elemText := getElementText(textName, idx)
		if !hasColumn(headerMap, elemText) {
			break
		}
		colIdx := headerMap[elemText]
		cellValue := row[colIdx]
		if cellValue == "" {
			continue
		}
		var v interface{} = cellValue
		switch field.GetType().String() {
		case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
			var iv int32
			_, err := fmt.Sscanf(cellValue, "%d", &iv)
			if err == nil {
				v = iv
			} else {
				return fmt.Errorf("repeated int32 parse error: %v, value=%s", err, cellValue)
			}
		case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
			var iv int64
			_, err := fmt.Sscanf(cellValue, "%d", &iv)
			if err == nil {
				v = iv
			} else {
				return fmt.Errorf("repeated int64 parse error: %v, value=%s", err, cellValue)
			}
		case "TYPE_FLOAT", "TYPE_DOUBLE":
			var fv float64
			_, err := fmt.Sscanf(cellValue, "%f", &fv)
			if err == nil {
				v = fv
			} else {
				return fmt.Errorf("repeated float parse error: %v, value=%s", err, cellValue)
			}
		case "TYPE_BOOL":
			if cellValue == "1" || cellValue == "true" || cellValue == "TRUE" {
				v = true
			} else {
				v = false
			}
		}
		data.AddRepeatedField(field, v)
	}
	return nil
}

func parseMapField(data *dynamic.Message, field *desc.FieldDescriptor, row []string, headerMap map[string]int, rowIdx int, base string) error {
	// map字段一般为message，包含key/value
	// 这里只做简单处理，实际可递归
	return nil // TODO: map支持
}

func parseMessage(data *dynamic.Message, msgDesc *desc.MessageDescriptor, row []string, headerMap map[string]int, rowIdx int, base string) error {
	for _, field := range msgDesc.GetFields() {
		if field.IsMap() {
			err := parseMapField(data, field, row, headerMap, rowIdx, base)
			if err != nil {
				return err
			}
		} else if field.IsRepeated() {
			err := parseRepeatedField(data, field, row, headerMap, rowIdx, base)
			if err != nil {
				return err
			}
		} else {
			err := parseSingleField(data, field, row, headerMap, rowIdx, base)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

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
	datas := make([]*dynamic.Message, 0, len(rows)-1)
	for i, row := range rows[1:] {
		data := dynamic.NewMessage(msgDesc)
		err := parseMessage(data, msgDesc, row, headerMap, i, "")
		if err != nil {
			return err
		}
		datas = append(datas, data)
	}
	// TODO: SaveJsonTable(config, datas)
	return nil
}

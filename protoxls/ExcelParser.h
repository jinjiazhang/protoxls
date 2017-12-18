#ifndef _JINJIAZHANG_EXCELPARSER_H_
#define _JINJIAZHANG_EXCELPARSER_H_

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"

#include "libxl.h"
using namespace libxl;

#include <google/protobuf/descriptor.h>
#include <google/protobuf/dynamic_message.h>
#include <google/protobuf/util/json_util.h>
using namespace google::protobuf;

class ExcelParser
{
public:
    ExcelParser(MessageFactory* factory);
    ~ExcelParser();

public:
    bool LoadSheet(string excel_name, string sheet_name);
    bool ParseData(const Descriptor* descriptor, vector<Message*>& datas);

private:
    bool ReadColumns();
    bool ParseMessage(Message* message, const Descriptor* descriptor, int row, string base);
    bool ParseField(Message* message, const FieldDescriptor* field, int row, string base);
    bool ParseSingle(Message* message, const FieldDescriptor* field, int row, string base);
    bool ParseRepeated(Message* message, const FieldDescriptor* field, int row, string base);
    bool ParseTable(Message* message, const FieldDescriptor* field, int row, string base);

private:
    string GetFiledText(const FieldDescriptor* field, string base);

private:
    Book* book_;
    Sheet* sheet_;
    string excel_name_;
    string sheet_name_;
    std::map<string, int> columns_;
    MessageFactory* factory_;
};


#endif
#ifndef _JINJIAZHANG_EXCELPARSER_H_
#define _JINJIAZHANG_EXCELPARSER_H_

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"

#include "libxl.h"
using namespace libxl;

#include <google/protobuf/descriptor.h>
using namespace google::protobuf;

class ExcelParser
{
public:
    ExcelParser();
    ~ExcelParser();

public:
    bool LoadSheet(string excel_name, string sheet_name);
    bool ParserData(const Descriptor* descriptor, vector<Message*>& datas);

private:
    bool ReadColumns();

private:
    Book* book_;
    Sheet* sheet_;
    string excel_name_;
    string sheet_name_;
    std::map<string, int> columns_;
};


#endif
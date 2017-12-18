#include "ExcelParser.h"

ExcelParser::ExcelParser()
{
    book_ = NULL;
    sheet_ = NULL;
}

ExcelParser::~ExcelParser()
{
    if (book_) {
        book_->release();
        book_ = NULL;
        sheet_ = NULL;
    }
}

bool ExcelParser::LoadSheet(string excel_name, string sheet_name)
{
    string::size_type pos = excel_name.rfind('.');
    string ext_name = excel_name.substr(pos == string::npos ? excel_name.length() : pos + 1);
    if (ext_name.compare("xls") != 0) {
        proto_error("only xls file supported, excel=%s", excel_name.c_str());
        return false;
    }

    book_ = xlCreateBook();
    if (!book_->load(excel_name.c_str())) {
        proto_error("load excel fail, excel=%s, error=%s", excel_name.c_str(), book_->errorMessage());
        book_->release();
        return false;
    }

    for (int i = 0; i < book_->sheetCount(); i++)
    {
        Sheet* sheet = book_->getSheet(i);
        if (sheet_name.compare(sheet->name()) == 0) 
        {
            sheet_ = sheet;
            break;
        }
    }

    if (sheet_ == NULL) {
        proto_error("sheet not found, excel=%s, sheet=%s", excel_name.c_str(), sheet_name.c_str());
        book_->release();
        return false;
    }

    excel_name_ = excel_name;
    sheet_name_ = sheet_name;
    return true;
}

bool ExcelParser::ParserData(const Descriptor* descriptor, vector<Message*>& datas)
{
    return true;
}

bool ExcelParser::ParserColumns(void* work_sheet)
{
    return true;
}
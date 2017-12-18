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
    if (ext_name.compare("xls") == 0)
    {
        book_ = xlCreateBook();
    }
    else if (ext_name.compare("xlsx") == 0)
    {
        book_ = xlCreateXMLBook();
    }
    else
    {
        proto_error("only xls file supported, excel=%s\n", excel_name.c_str());
        return false;
    }

    book_->setKey("protoxls", "windows-27262a0805c8e4046cbd6661ael7mahf");
    if (!book_->load(excel_name.c_str()))
    {
        proto_error("load excel fail, excel=%s, error=%s\n", excel_name.c_str(), book_->errorMessage());
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

    if (sheet_ == NULL)
    {
        proto_error("sheet not found, excel=%s, sheet=%s\n", excel_name.c_str(), sheet_name.c_str());
        book_->release();
        return false;
    }

    excel_name_ = excel_name;
    sheet_name_ = sheet_name;
    return true;
}

bool ExcelParser::ParserData(const Descriptor* descriptor, vector<Message*>& datas)
{
    PROTO_ASSERT(sheet_ != NULL);
    PROTO_DO(ReadColumns());
    return true;
}

bool ExcelParser::ReadColumns()
{
    PROTO_ASSERT(sheet_ != NULL);
    PROTO_ASSERT(sheet_->firstRow() == 0);
    PROTO_ASSERT(sheet_->firstCol() == 0);
    PROTO_ASSERT(sheet_->lastRow() > 0);
    PROTO_ASSERT(sheet_->lastCol() > 0);
    
    for (int col = 0; col < sheet_->lastCol(); col++)
    {
        PROTO_ASSERT(sheet_->cellType(0, col) == CELLTYPE_STRING);
        const char* text = sheet_->readStr(0, col);
        columns_.insert(std::make_pair(string(text), col));
    }
    return true;
}

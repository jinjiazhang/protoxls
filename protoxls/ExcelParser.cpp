#include "ExcelParser.h"
#include "ParseHelper.h"

#define EXPECT_CELLTYPE(row, col, cell_type, text_name) \
    if (sheet_->cellType(row, col) != (cell_type)) { \
        proto_error("ParseFiled cell type error, expect=%s, row=%d, col=%d, text=%s\n", \
            #cell_type, row, col, text_name.c_str()); \
        return false; \
    }

ExcelParser::ExcelParser(MessageFactory* factory)
{
    book_ = NULL;
    sheet_ = NULL;
    factory_ = factory;
    field_format_ = "%s%s";
    array_format_ = "%s[%d]";
    key_format_ = "%skey";
    value_format_ = "%svalue";
    index_start_ = 1;
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
        proto_error("LoadSheet only xls/xlsx file supported, excel=%s\n", excel_name.c_str());
        return false;
    }

#ifdef _WIN32
    book_->setKey("protoxls", "windows-27262a0805c8e4046cbd6661ael7mahf");
#else
    book_->setKey("protoxls", "linux-e7d61a7895a8a4140c0d26314em7sacf");
#endif

    if (!book_->load(excel_name.c_str()))
    {
        proto_error("LoadSheet load excel fail, excel=%s, error=%s\n", excel_name.c_str(), book_->errorMessage());
        book_->release();
        book_ = NULL;
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
        proto_error("LoadSheet sheet not found, excel=%s, sheet=%s\n", excel_name.c_str(), sheet_name.c_str());
        book_->release();
        book_ = NULL;
        return false;
    }

    excel_name_ = excel_name;
    sheet_name_ = sheet_name;
    return true;
}

bool ExcelParser::ParseData(const Descriptor* descriptor, vector<Message*>& datas)
{
    PROTO_ASSERT(sheet_ != NULL);
    if (sheet_->firstRow() >= sheet_->lastRow()) {
        proto_error("ParseData sheet row empyt, first=%d, last=%d\n", sheet_->firstRow(), sheet_->lastRow());
        return false;
    }

    if (sheet_->firstCol() >= sheet_->lastCol()) {
        proto_error("ParseData sheet col empyt, first=%d, last=%d\n", sheet_->firstRow(), sheet_->lastRow());
        return false;
    }

    int row = sheet_->firstRow();
    for (int col = sheet_->firstCol(); col < sheet_->lastCol(); col++)
    {
        if (sheet_->cellType(row, col) == CELLTYPE_STRING)
        {
            const char* text = sheet_->readStr(row, col);
            columns_.insert(std::make_pair(text, col));
        }
    }

    for (int row = sheet_->firstRow() + 1; row < sheet_->lastRow(); row++)
    {
        Message* message = factory_->GetPrototype(descriptor)->New();
        if (!ParseMessage(message, descriptor, row, "")) {
            proto_error("ParseData parse row fail, row=%d\n", row);
            return false;
        }
        datas.push_back(message);
    }
    return true;
}

string ExcelParser::GetFiledText(const FieldDescriptor* field, string base)
{
    const FieldOptions& option = field->options();
    string text_name = utf82ansi(option.GetExtension(text));
    if (text_name.empty()) {
        text_name = field->name();
    }
    if (base.empty()) {
        return text_name;
    }

    char full_name[256];
    sprintf(full_name, field_format_.c_str(), base.c_str(), text_name.c_str());
    return string(full_name);
}

string ExcelParser::GetElementText(string text_name, int index)
{
    char full_name[256];
    sprintf(full_name, array_format_.c_str(), text_name.c_str(), index);
    return string(full_name);
}

bool ExcelParser::HasFiled(const FieldDescriptor* field, int row, string base)
{
    string text_name = GetFiledText(field, base);
    if (columns_.find(text_name) == columns_.end()) {
        return false;
    }

    int col = columns_[text_name];
    CellType cell_type = sheet_->cellType(row, col);
    if (cell_type == CELLTYPE_EMPTY || cell_type == CELLTYPE_BLANK) {
        return false;
    }
    return true;
}

bool ExcelParser::HasMessage(const FieldDescriptor* field, int row, string base)
{
    const Descriptor* descriptor = field->message_type();
    for (int i = 0; i < descriptor->field_count(); i++)
    {
        const FieldDescriptor* subfield = descriptor->field(i);
        if (subfield->is_repeated())
        {
            if (HasElement(subfield, index_start_, row, base)) {
                return true;
            }
        }
        else
        {
            if (HasFiled(subfield, row, base)) {
                return true;
            }
        }
    }
    return false;
}

bool ExcelParser::HasElement(const FieldDescriptor* field, int index, int row, string base)
{
    string text_name = GetFiledText(field, base);
    string element_text = GetElementText(text_name, index);
    if (field->cpp_type() == FieldDescriptor::CPPTYPE_MESSAGE) {
        return HasMessage(field, row, element_text);
    }

    if (columns_.find(element_text) == columns_.end()) {
        return false;
    }

    int col = columns_[element_text];
    CellType cell_type = sheet_->cellType(row, col);
    if (cell_type == CELLTYPE_EMPTY || cell_type == CELLTYPE_BLANK) {
        return false;
    }
    return true;
}

bool ExcelParser::UnixTimestamp(int row, int col)
{
    if (!sheet_->isDate(row, col)) {
        return true;
    }

    struct tm stm;
    memset(&stm, 0, sizeof(stm));
    double value = sheet_->readNum(row, col);
    bool result = book_->dateUnpack(value, &stm.tm_year, &stm.tm_mon, &stm.tm_mday, &stm.tm_hour, &stm.tm_min, &stm.tm_sec);
    if (!result) {
        proto_error("UnixTimestamp date unpack fail, value=%f, row=%d, col=%d\n", value, row, col);
        return false;
    }

    stm.tm_year -= 1900;
    stm.tm_mon -= 1;
    time_t time = mktime(&stm) ;
    if (time < 0) {
        proto_error("UnixTimestamp make time fail, value=%f, row=%d, col=%d\n", value, row, col);
        return false;
    }

    sheet_->writeNum(row, col, (double)time);
    return true;
}

bool ExcelParser::ParseMessage(Message* message, const Descriptor* descriptor, int row, string base)
{
    for (int i = 0; i < descriptor->field_count(); i++)
    {
        const FieldDescriptor* field = descriptor->field(i);
        if (!ParseField(message, field, row, base)) {
            proto_error("ParseMessage parse field fail, row=%d, field=%s\n", row, field->full_name().c_str());
            return false;
        }
    }
    return true;
}

bool ExcelParser::ParseField(Message* message, const FieldDescriptor* field, int row, string base)
{
    if (field->is_map())
        return ParseMultiple(message, field, row, base);
    else if (field->is_required())
        return ParseSingle(message, field, row, base);
    else if (field->is_optional())
        return ParseSingle(message, field, row, base);
    else if (field->is_repeated())
        return ParseRepeated(message, field, row, base);
    else
        return false;
    return true;
}

bool ExcelParser::ParseSingle(Message* message, const FieldDescriptor* field, int row, string base)
{
    string text_name = GetFiledText(field, base);
    if (field->cpp_type() == FieldDescriptor::CPPTYPE_MESSAGE)
    {
        const Reflection* reflection = message->GetReflection();
        Message* submessage = reflection->MutableMessage(message, field);
        return ParseMessage(submessage, field->message_type(), row, text_name);
    }

    if (columns_.find(text_name) == columns_.end())
    {
        proto_error("ParseSingle column not found, name=%s\n", text_name.c_str());
        return false;
    }

    int col = columns_[text_name];
    CellType cell_type = sheet_->cellType(row, col);
    if (cell_type == CELLTYPE_EMPTY || cell_type == CELLTYPE_BLANK)
    {
        // proto_warn("ParseSingle cell empty, name=%s, row=%d\n", text_name.c_str(), row);
        return true;
    }

    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
    case FieldDescriptor::CPPTYPE_FLOAT:
    case FieldDescriptor::CPPTYPE_INT32:
    case FieldDescriptor::CPPTYPE_UINT32:
    case FieldDescriptor::CPPTYPE_INT64:
    case FieldDescriptor::CPPTYPE_UINT64:
        PROTO_DO(UnixTimestamp(row, col));
        EXPECT_CELLTYPE(row, col, CELLTYPE_NUMBER, text_name);
        ParseHelper::SetNumberField(message, field, sheet_->readNum(row, col));
        break;
    case FieldDescriptor::CPPTYPE_BOOL:
        EXPECT_CELLTYPE(row, col, CELLTYPE_BOOLEAN, text_name);
        ParseHelper::SetBoolField(message, field, sheet_->readBool(row, col));
        break;
    case FieldDescriptor::CPPTYPE_ENUM:
        EXPECT_CELLTYPE(row, col, CELLTYPE_STRING, text_name);
        ParseHelper::SetEnumField(message, field, sheet_->readStr(row, col));
        break;
    case FieldDescriptor::CPPTYPE_STRING:
        EXPECT_CELLTYPE(row, col, CELLTYPE_STRING, text_name);
        ParseHelper::SetStringField(message, field, sheet_->readStr(row, col));
        break;
    default:
        proto_error("ParseSingle field unknow type, field=%s\n", field->full_name().c_str());
        return false;
    }
    return true;
}

bool ExcelParser::ParseMultiple(Message* message, const FieldDescriptor* field, int row, string base)
{
    string text_name = GetFiledText(field, base);
    if (field->cpp_type() == FieldDescriptor::CPPTYPE_MESSAGE)
    {
        int index = index_start_;
        while (HasElement(field, index, row, base))
        {
            string element_base = GetElementText(text_name, index);
            const Reflection* reflection = message->GetReflection();
            Message* submessage = reflection->AddMessage(message, field);
            PROTO_DO(ParseMessage(submessage, field->message_type(), row, element_base));
            index += 1;
        }
        return true;
    }

    int index = index_start_;
    while (HasElement(field, index, row, base))
    {
        string element_text = GetElementText(text_name, index);
        int col = columns_[element_text];
        CellType cell_type = sheet_->cellType(row, col);

        switch (field->cpp_type())
        {
        case FieldDescriptor::CPPTYPE_DOUBLE:
        case FieldDescriptor::CPPTYPE_FLOAT:
        case FieldDescriptor::CPPTYPE_INT32:
        case FieldDescriptor::CPPTYPE_UINT32:
        case FieldDescriptor::CPPTYPE_INT64:
        case FieldDescriptor::CPPTYPE_UINT64:
            PROTO_DO(UnixTimestamp(row, col));
            EXPECT_CELLTYPE(row, col, CELLTYPE_NUMBER, element_text);
            ParseHelper::AddNumberField(message, field, sheet_->readNum(row, col));
            break;
        case FieldDescriptor::CPPTYPE_BOOL:
            EXPECT_CELLTYPE(row, col, CELLTYPE_BOOLEAN, element_text);
            ParseHelper::AddBoolField(message, field, sheet_->readBool(row, col));
            break;
        case FieldDescriptor::CPPTYPE_ENUM:
            EXPECT_CELLTYPE(row, col, CELLTYPE_STRING, element_text);
            ParseHelper::AddEnumField(message, field, sheet_->readStr(row, col));
            break;
        case FieldDescriptor::CPPTYPE_STRING:
            EXPECT_CELLTYPE(row, col, CELLTYPE_STRING, element_text);
            ParseHelper::AddStringField(message, field, sheet_->readStr(row, col));
            break;
        default:
            proto_error("ParseMultiple field unknow type, field=%s\n", field->full_name().c_str());
            return false;
        }

        index += 1;
    }
    return true;
}

bool ExcelParser::ParseArray(Message* message, const FieldDescriptor* field, int row, string base)
{
    string text_name = GetFiledText(field, base);
    if (columns_.find(text_name) == columns_.end())
    {
        proto_error("ParseArray column not found, name=%s\n", text_name.c_str());
        return false;
    }

    int col = columns_[text_name];
    CellType cell_type = sheet_->cellType(row, col);
    if (cell_type == CELLTYPE_EMPTY || cell_type == CELLTYPE_BLANK)
    {
        // proto_warn("ParseArray cell empty, name=%s, row=%d\n", text_name.c_str(), row);
        return true;
    }

    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
    case FieldDescriptor::CPPTYPE_FLOAT:
    case FieldDescriptor::CPPTYPE_INT32:
    case FieldDescriptor::CPPTYPE_UINT32:
    case FieldDescriptor::CPPTYPE_INT64:
    case FieldDescriptor::CPPTYPE_UINT64:
        break;
    default:
        proto_error("ParseArray only number supported, field=%s\n", field->full_name().c_str());
        return false;
    }

    if (cell_type == CELLTYPE_NUMBER) 
    {
        ParseHelper::AddNumberField(message, field, sheet_->readNum(row, col));
        return true;
    }

    EXPECT_CELLTYPE(row, col, CELLTYPE_STRING, text_name);
    return ParseHelper::FillNumberArray(message, field, sheet_->readStr(row, col));
}

bool ExcelParser::ParseRepeated(Message* message, const FieldDescriptor* field, int row, string base)
{
    string text_name = GetFiledText(field, base);
    if (columns_.find(text_name) != columns_.end())
    {
        // parse number array split by ";"
        return ParseArray(message, field, row, base);
    }
    else
    {
        // parse multiple columns by index
        return ParseMultiple(message, field, row, base);
    }
}
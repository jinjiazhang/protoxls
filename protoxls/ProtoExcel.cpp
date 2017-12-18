#include <algorithm>
#include "ProtoExcel.h"
#include "ExcelParser.h"

ProtoExcel::ProtoExcel()
:importer_(&sourceTree_, &errorCollector_)
{
    sourceTree_.MapPath("", "./");
    sourceTree_.MapPath("google/protobuf/", "./");
}

ProtoExcel::~ProtoExcel()
{

}

bool ProtoExcel::ParseScheme(const char* file)
{
    const FileDescriptor* parsed_file = importer_.Import(file);
    if (parsed_file == NULL)
    {
        proto_error("ParseScheme import proto fail, file=%s\n", file);
        return false;
    }

    for (int i = 0; i < parsed_file->message_type_count(); i++)
    {
        const Descriptor* descriptor = parsed_file->message_type(i);
        PROTO_DO(ParseConfig(descriptor));
    }
    return true;
}

bool ProtoExcel::ParseConfig(const Descriptor* descriptor)
{
    const MessageOptions& option = descriptor->options();
    vector<string> excel_names = Split(utf82ansi(option.GetExtension(excel)), ";");
    vector<string> sheet_names = Split(utf82ansi(option.GetExtension(sheet)), ";");
    if (excel_names.size() > 1 && sheet_names.size() > 1 && excel_names.size() != sheet_names.size())
    {
        proto_error("ParseConfig excel count not equal to sheet count, message name=%s\n", descriptor->name().c_str());
        return false;
    }

    size_t name_size = (std::max)(excel_names.size(), sheet_names.size());
    for (int i = excel_names.size(); i < name_size; i++)
    {
        string back_name = excel_names.back();
        excel_names.push_back(back_name);
    }

    for (int i = sheet_names.size(); i < name_size; i++)
    {
        string back_name = sheet_names.back();
        sheet_names.push_back(back_name);
    }

    vector<Message*> datas;
    for (int i = 0; i < name_size; i++)
    {
        ExcelParser parser;
        PROTO_DO(parser.LoadSheet(excel_names[i], sheet_names[i]));
        PROTO_DO(parser.ParserData(descriptor, datas));
    }
    return true;
}

bool ProtoExcel::ExportResult()
{
    return true;
}
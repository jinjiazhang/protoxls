#include "ProtoExcel.h"

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
    if (parsed_file == NULL) {
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
    vector<string> excel_names = Split(option.GetExtension(excel), ";");
    vector<string> sheet_names = Split(option.GetExtension(sheet), ";");
    if (excel_names.size() > 1 && sheet_names.size() > 1 && excel_names.size() != sheet_names.size()) {
        proto_error("ParseConfig excel count not equal to sheet count, message name = ", descriptor->name().c_str());
        return false;
    }

    size_t name_size = std::max(excel_names.size(), sheet_names.size());
    return true;
}

bool ProtoExcel::ExportResult()
{
    return true;
}
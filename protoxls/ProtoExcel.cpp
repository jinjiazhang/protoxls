#include <algorithm>
#include "ProtoExcel.h"
#include "ExcelParser.h"
#include "ConfigStore.h"
#include "LuaExporter.h"
#include "BinExporter.h"

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
        if (!ParseConfig(descriptor))
        {
            proto_error("ParseScheme parse config fail, scheme=%s\n", descriptor->full_name().c_str());
            continue;
        }
    }
    return true;
}

bool ProtoExcel::ParseConfig(const Descriptor* descriptor)
{
    const MessageOptions& option = descriptor->options();
    vector<string> excel_names = Split(utf82ansi(option.GetExtension(excel)), ";");
    vector<string> sheet_names = Split(utf82ansi(option.GetExtension(sheet)), ";");
    vector<string> key_names = Split(utf82ansi(option.GetExtension(key)), ";");
    if (excel_names.size() == 0 || sheet_names.size() == 0)
    {
        proto_error("ParseConfig excel/sheet option miss, scheme=%s\n", descriptor->full_name().c_str());
        return false;
    }

    if (excel_names.size() > 1 && sheet_names.size() > 1 && excel_names.size() != sheet_names.size())
    {
        proto_error("ParseConfig excel/sheet size unmatch, scheme=%s\n", descriptor->full_name().c_str());
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
        ExcelParser parser(&factory_);
        if (!parser.LoadSheet(excel_names[i], sheet_names[i])) {
            proto_error("ParseConfig load sheet fail, excel=%s, sheet=%s\n", excel_names[i].c_str(), sheet_names[i].c_str());
            return false;
        }

        if (!parser.ParseData(descriptor, datas)) {
            proto_error("ParseConfig parse data fail, excel=%s, sheet=%s\n", excel_names[i].c_str(), sheet_names[i].c_str());
            return false;
        }
    }

    ConfigStore* store = new ConfigStore(descriptor);
    store->ImportData(datas);
    if (!store->BuildStore(key_names)) {
        proto_error("ParseConfig build store fail, scheme=%s\n", descriptor->full_name().c_str());
        return false;
    }

    parseds_.push_back(store);
    return true;
}

bool ProtoExcel::ExportResult()
{
    for (int i = 0; i < parseds_.size(); i++)
    {
        ConfigStore* store = parseds_[i];
        LuaExporter::ExportResult(store);
        BinExporter::ExportResult(store);
    }
    return true;
}
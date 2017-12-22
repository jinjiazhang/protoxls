#include "LuaExporter.h"

LuaExporter::LuaExporter()
{

}

LuaExporter::~LuaExporter()
{

}

bool LuaExporter::ExportResult(const Descriptor* descriptor, vector<Message*> datas)
{
    const MessageOptions& option = descriptor->options();
    keys_ = Split(option.GetExtension(key), ";");
    output_ = option.GetExtension(output);
    prefix_ = option.GetExtension(prefix);
    suffix_ = option.GetExtension(suffix);

    string result = "";
    if (!prefix_.empty()) {
        result += prefix_;
    }

    string code;
    if (!GenerateCode(code)) {
        proto_error("ExportResult generate code fail, scheme=%s\n", descriptor->full_name().c_str());
        return false;
    }

    result += code;

    if (!suffix_.empty()) {
        result += suffix_;
    }

    if (!WriteToFile(result)) {
        proto_error("ExportResult write to file fail, scheme=%s\n", descriptor->full_name().c_str());
        return false;
    }

    return true;
}

bool LuaExporter::GenerateCode(string& code)
{
    code = "{\n}\n";
    return true;
}

bool LuaExporter::WriteToFile(string& result)
{
    FILE* fp = fopen(output_.c_str(), "w");
    if (fp == NULL) {
        proto_error("WriteToFile open file fail, file=%s\n", output_.c_str());
        return false;
    }

    fprintf(fp, result.c_str());
    fclose(fp);
    return true;
}
#ifndef _JINJIAZHANG_PROTOEXCEL_H_
#define _JINJIAZHANG_PROTOEXCEL_H_

#include <google/protobuf/descriptor.h>
#include <google/protobuf/dynamic_message.h>
#include <google/protobuf/compiler/importer.h>
#include <google/protobuf/stubs/strutil.h>

using namespace google::protobuf;
using namespace google::protobuf::io;
using namespace google::protobuf::compiler;
using namespace google::protobuf::internal;

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"

class ProtoErrorCollector : public MultiFileErrorCollector
{
    virtual void AddError(const std::string& filename, int line, int column, const std::string& message)
    {
        proto_error("[file]%s line %d, column %d : %s\n", filename.c_str(), line, column, message.c_str());
    }
};

class ProtoExcel
{
public:
    ProtoExcel();
    ~ProtoExcel();

public:
    bool ParseScheme(const char* proto);
    bool ExportResult();

private:
    bool ParseConfig(const Descriptor* descriptor);
    bool ParseExcel(const Descriptor* descriptor, vector<const Message*>& datas, string excel_name, string sheet_name);

private:
    ProtoErrorCollector errorCollector_;
    DiskSourceTree sourceTree_;
    Importer importer_;
    DynamicMessageFactory factory_;
};

#endif
#ifndef _JINJIAZHANG_LUAEXPORTER_H_
#define _JINJIAZHANG_LUAEXPORTER_H_

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"

#include <google/protobuf/descriptor.h>
#include <google/protobuf/dynamic_message.h>
#include <google/protobuf/compiler/importer.h>
#include <google/protobuf/stubs/strutil.h>

using namespace google::protobuf;
using namespace google::protobuf::io;
using namespace google::protobuf::compiler;
using namespace google::protobuf::internal;

class LuaExporter
{
public:
    LuaExporter();
    ~LuaExporter();

public:
    bool ExportResult(const Descriptor* descriptor, vector<Message*> datas);

private:
    bool GenerateCode(string& code);
    bool WriteToFile(string& result);

private:
    vector<string> keys_;
    string output_;
    string prefix_;
    string suffix_;
};


#endif
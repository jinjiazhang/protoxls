#ifndef _JINJIAZHANG_LUAEXPORTER_H_
#define _JINJIAZHANG_LUAEXPORTER_H_

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"
#include "ConfigStore.h"

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
    static bool ExportResult(ConfigStore* store);
    static string GenerateCode(ConfigStore* store, int layer=0);
    static string GenerateCode(const Message& message, int layer=0);
    static string GenerateField(const Message& message, const FieldDescriptor* field, int layer=0);
    static string GenerateTable(const Message& message, const FieldDescriptor* field, int layer=0);
    static string GenerateSingle(const Message& message, const FieldDescriptor* field, int layer=0);
    static string GenerateMultiple(const Message& message, const FieldDescriptor* field, int layer=0);
};


#endif
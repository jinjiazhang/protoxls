#ifndef _JINJIAZHANG_BINEXPORTER_H_
#define _JINJIAZHANG_BINEXPORTER_H_

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

class BinExporter
{
public:
    static bool ExportResult(ConfigStore* store);
};


#endif
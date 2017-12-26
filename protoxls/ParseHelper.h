#ifndef _JINJIAZHANG_PARSEHELPER_H_
#define _JINJIAZHANG_PARSEHELPER_H_

#include <google/protobuf/descriptor.h>
#include <google/protobuf/dynamic_message.h>
#include <google/protobuf/util/json_util.h>
#include <google/protobuf/stubs/strutil.h>
using namespace google::protobuf;

#include "logger.h"
#include "strconv.h"
#include "option.pb.h"

class ParseHelper
{
public:
    static void SetNumberField(Message* message, const FieldDescriptor* field, double value);
    static void SetBoolField(Message* message, const FieldDescriptor* field, bool value);
    static void SetEnumField(Message* message, const FieldDescriptor* field, const char* text);
    static void SetStringField(Message* message, const FieldDescriptor* field, const char* value);
    
    static void AddNumberField(Message* message, const FieldDescriptor* field, double value);
    static void AddBoolField(Message* message, const FieldDescriptor* field, bool value);
    static void AddEnumField(Message* message, const FieldDescriptor* field, const char* text);
    static void AddStringField(Message* message, const FieldDescriptor* field, const char* value);

    static bool GetEnumValue(const FieldDescriptor* field, const char* text, int* value);
    static bool FillNumberArray(Message* message, const FieldDescriptor* field, const char* text);
    static bool GetNumberField(const Message& message, const FieldDescriptor* field, int64* value);
    static bool GetStringField(const Message& message, const FieldDescriptor* field, string* value);
};

#endif
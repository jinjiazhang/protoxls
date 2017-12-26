#include "ParseHelper.h"

void ParseHelper::SetNumberField(Message* message, const FieldDescriptor* field, double value)
{
    const Reflection* reflection = message->GetReflection();
    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
        reflection->SetDouble(message, field, (double)value);
        break;
    case FieldDescriptor::CPPTYPE_FLOAT:
        reflection->SetFloat(message, field, (float)value);
        break;
    case FieldDescriptor::CPPTYPE_INT32:
        reflection->SetInt32(message, field, (int32)value);
        break;
    case FieldDescriptor::CPPTYPE_UINT32:
        reflection->SetUInt32(message, field, (uint32)value);
        break;
    case FieldDescriptor::CPPTYPE_INT64:
        reflection->SetInt64(message, field, (int64)value);
        break;
    case FieldDescriptor::CPPTYPE_UINT64:
        reflection->SetUInt64(message, field, (uint64)value); 
        break;
    default:
        proto_error("SetNumberField field unknow type, field=%s\n", field->full_name().c_str());
        break;
    }
}

void ParseHelper::SetBoolField(Message* message, const FieldDescriptor* field, bool value)
{
    const Reflection* reflection = message->GetReflection();
    reflection->SetBool(message, field, value);
}

void ParseHelper::SetEnumField(Message* message, const FieldDescriptor* field, const char* text)
{
    int value = field->default_value_enum()->number();
    if (!GetEnumValue(field, text, &value)) {
        proto_error("SetEnumField enum not found, field=%s, text=%s\n", field->full_name().c_str(), text);
    }
    const Reflection* reflection = message->GetReflection();
    reflection->SetEnumValue(message, field, value);
}

void ParseHelper::SetStringField(Message* message, const FieldDescriptor* field, const char* value)
{
    const Reflection* reflection = message->GetReflection();
    reflection->SetString(message, field, ansi2utf8(value));
}

void ParseHelper::AddNumberField(Message* message, const FieldDescriptor* field, double value)
{
    const Reflection* reflection = message->GetReflection();
    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
        reflection->AddDouble(message, field, (double)value);
        break;
    case FieldDescriptor::CPPTYPE_FLOAT:
        reflection->AddFloat(message, field, (float)value);
        break;
    case FieldDescriptor::CPPTYPE_INT32:
        reflection->AddInt32(message, field, (int32)value);
        break;
    case FieldDescriptor::CPPTYPE_UINT32:
        reflection->AddUInt32(message, field, (uint32)value);
        break;
    case FieldDescriptor::CPPTYPE_INT64:
        reflection->AddInt64(message, field, (int64)value);
        break;
    case FieldDescriptor::CPPTYPE_UINT64:
        reflection->AddUInt64(message, field, (uint64)value); 
        break;
    default:
        proto_error("SetNumberField field unknow type, field=%s\n", field->full_name().c_str());
        break;
    }
}

void ParseHelper::AddBoolField(Message* message, const FieldDescriptor* field, bool value)
{
    const Reflection* reflection = message->GetReflection();
    reflection->AddBool(message, field, value);
}

void ParseHelper::AddEnumField(Message* message, const FieldDescriptor* field, const char* text)
{
    int value = field->default_value_enum()->number();
    if (!GetEnumValue(field, text, &value)) {
        proto_error("AddEnumField enum not found, field=%s, text=%s\n", field->full_name().c_str(), text);
    }
    const Reflection* reflection = message->GetReflection();
    reflection->AddEnumValue(message, field, value);
}

void ParseHelper::AddStringField(Message* message, const FieldDescriptor* field, const char* value)
{
    const Reflection* reflection = message->GetReflection();
    reflection->AddString(message, field, ansi2utf8(value));
}

bool ParseHelper::GetEnumValue(const FieldDescriptor* field, const char* text, int* value)
{
    string enum_text = ansi2utf8(text);
    const EnumDescriptor* enum_type = field->enum_type();
    for (int i = 0; i < enum_type->value_count(); i++)
    {
        const EnumValueDescriptor* enumerate = enum_type->value(i);
        if (enum_text.compare(enumerate->name()) == 0) {
            *value = enumerate->number();
            return true;
        }

        string text_name = enumerate->options().GetExtension(cname);
        if (!text_name.empty() && enum_text.compare(text_name) == 0) {
            *value = enumerate->number();
            return true;
        }
    }
    return false;
}

bool ParseHelper::FillNumberArray(Message* message, const FieldDescriptor* field, const char* text)
{
    vector<string> values = Split(ansi2utf8(text), ";");
    for (size_t i = 0; i < values.size(); i++)
    {
        double value = atof(values[i].c_str());
        AddNumberField(message, field, value);
    }
    return true;
}

bool ParseHelper::GetNumberField(const Message& message, const FieldDescriptor* field, int64* value)
{
    const Reflection* reflection = message.GetReflection();
    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
        *value = (int64)reflection->GetDouble(message, field);
        break;
    case FieldDescriptor::CPPTYPE_FLOAT:
        *value = (int64)reflection->GetFloat(message, field);
        break;
    case FieldDescriptor::CPPTYPE_INT32:
        *value = (int64)reflection->GetInt32(message, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT32:
        *value = (int64)reflection->GetUInt32(message, field);
        break;
    case FieldDescriptor::CPPTYPE_INT64:
        *value = (int64)reflection->GetInt64(message, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT64:
        *value = (int64)reflection->GetUInt64(message, field);
        break;
    case FieldDescriptor::CPPTYPE_ENUM:
        *value = (int64)reflection->GetEnumValue(message, field);
        break;
    default:
        proto_error("GetNumberField field unknow type, field=%s\n", field->full_name().c_str());
        return false;
    }
    return true;
}

bool ParseHelper::GetStringField(const Message& message, const FieldDescriptor* field, string* value)
{
    const Reflection* reflection = message.GetReflection();
    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_STRING:
        *value = reflection->GetString(message, field);
        break;
    default:
        proto_error("GetStringField field unknow type, field=%s\n", field->full_name().c_str());
        return false;
    }
    return true;
}
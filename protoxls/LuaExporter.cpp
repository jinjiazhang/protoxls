#include <sstream>
#include "LuaExporter.h"
#include "ParseHelper.h"

bool LuaExporter::ExportResult(ConfigStore* store)
{
    const Descriptor* descriptor = store->GetDescriptor();
    const MessageOptions& option = descriptor->options();
    string prefix_text = option.GetExtension(prefix);
    string suffix_text = option.GetExtension(suffix);

    string result;
    result += prefix_text;
    result += GenerateCode(store);
    result += suffix_text;

    string output_text = option.GetExtension(output);
    string file_name = output_text + ".lua";

    FILE* fp = fopen(file_name.c_str(), "w");
    if (fp == NULL) {
        proto_error("ExportResult open file fail, file=%s\n", file_name.c_str());
        return false;
    }

    fprintf(fp, result.c_str());
    fclose(fp);
    return true;
}

#define FLUSH_LAYER(code, layer) do {for (int i = 0; i < (layer); i++) { code << "\t";}} while(0)

string LuaExporter::GenerateCode(ConfigStore* store, int layer)
{
    std::stringstream code;
    code << "{\n";
    
    vector<StoreKey> store_keys;
    store->ExportKeys(store_keys);
    std::sort(store_keys.begin(), store_keys.end());

    for (int i = 0; i < store_keys.size(); i++)
    {
        FLUSH_LAYER(code, layer + 1);
        StoreKey store_key = store_keys[i];
        if (store_key.key_type == KEY_TINTEGER)
            code << "[" << store_key.num_key << "]";
        else
            code << "['" << store_key.str_key << "']"; 

        code << " = ";
        ConfigStore* sub_store = store->GetConfig(store_key);
        if (sub_store->HasStoreMap())
        {
            code << GenerateCode(sub_store, layer + 1);
        }
        else
        {
            code << GenerateCode(*sub_store->GetData());
        }

        code << ",\n";
    }

    FLUSH_LAYER(code, layer);
    code << "}";
    return code.str();
}

string LuaExporter::GenerateCode(const Message& message, int layer)
{
    std::stringstream code;
    code << "{";
    
    const Descriptor* descriptor = message.GetDescriptor();
    for (int i = 0; i < descriptor->field_count(); i++)
    {
        const FieldDescriptor* field = descriptor->field(i);
        code << field->name();
        code << "=";
        code << GenerateField(message, field, layer + 1);
        if (i < descriptor->field_count() - 1) {
            code << ", ";
        }
    }

    code << "}";
    return code.str();
}

string LuaExporter::GenerateField(const Message& message, const FieldDescriptor* field, int layer)
{
    if (field->is_map())
        return GenerateTable(message, field, layer);
    else if (field->is_required())
        return GenerateSingle(message, field, layer);
    else if (field->is_optional())
        return GenerateSingle(message, field, layer);
    else if (field->is_repeated())
        return GenerateMultiple(message, field, layer);
    else
        return "";
}

string LuaExporter::GenerateTable(const Message& message, const FieldDescriptor* field, int layer)
{
    std::stringstream code;
    code << "{";

    const Reflection* reflection = message.GetReflection();
    int field_size = reflection->FieldSize(message, field);

    const Descriptor* descriptor = field->message_type();
    const FieldDescriptor* key = descriptor->field(0);
    const FieldDescriptor* value = descriptor->field(1);

    for (int index = 0; index < field_size; index++)
    {
        const Message& submessage = reflection->GetRepeatedMessage(message, field, index);
        StoreKey store_key;
        ConfigStore::GetKeyVal(submessage, key->name(), &store_key);
        if (store_key.key_type == KEY_TINTEGER)
            code << "[" << store_key.num_key << "]";
        else
            code << "['" << store_key.str_key << "']";

        code << " = ";
        code << GenerateField(submessage, value, layer);

        if (index < field_size - 1) {
            code << ", ";
        }
    }

    code << "}";
    return code.str();
}

string LuaExporter::GenerateSingle(const Message& message, const FieldDescriptor* field, int layer)
{
    std::stringstream code;
    const Reflection* reflection = message.GetReflection();
    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_DOUBLE:
        code << reflection->GetDouble(message, field);
        break;
    case FieldDescriptor::CPPTYPE_FLOAT:
        code << reflection->GetFloat(message, field);
        break;
    case FieldDescriptor::CPPTYPE_INT32:
        code << reflection->GetInt32(message, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT32:
        code << reflection->GetUInt32(message, field);
        break;
    case FieldDescriptor::CPPTYPE_INT64:
        code << reflection->GetInt64(message, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT64:
        code << reflection->GetUInt64(message, field);
        break;
    case FieldDescriptor::CPPTYPE_ENUM:
        code << reflection->GetEnumValue(message, field);
        break;
    case FieldDescriptor::CPPTYPE_BOOL:
        code << reflection->GetBool(message, field);
        break;
    case FieldDescriptor::CPPTYPE_STRING:
        code << "'" << reflection->GetString(message, field) << "'";
        break;
    case FieldDescriptor::CPPTYPE_MESSAGE:
        code << GenerateCode(reflection->GetMessage(message, field), layer + 1);
        break;
    default:
        proto_error("GenerateSingle field unknow type, field=%s\n", field->full_name().c_str());
        break;
    }
    return code.str();
}

string LuaExporter::GenerateMultiple(const Message& message, const FieldDescriptor* field, int layer)
{
    std::stringstream code;
    code << "{";

    const Reflection* reflection = message.GetReflection();
    int field_size = reflection->FieldSize(message, field);

    for (int index = 0; index < field_size; index++)
    {
        const Reflection* reflection = message.GetReflection();
        switch (field->cpp_type())
        {
        case FieldDescriptor::CPPTYPE_DOUBLE:
            code << reflection->GetRepeatedDouble(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_FLOAT:
            code << reflection->GetRepeatedFloat(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_INT32:
            code << reflection->GetRepeatedInt32(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_UINT32:
            code << reflection->GetRepeatedUInt32(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_INT64:
            code << reflection->GetRepeatedInt64(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_UINT64:
            code << reflection->GetRepeatedUInt64(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_ENUM:
            code << reflection->GetRepeatedEnumValue(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_BOOL:
            code << reflection->GetRepeatedBool(message, field, index);
            break;
        case FieldDescriptor::CPPTYPE_STRING:
            code << "'" << reflection->GetRepeatedString(message, field, index) << "'";
            break;
        case FieldDescriptor::CPPTYPE_MESSAGE:
            code << GenerateCode(reflection->GetRepeatedMessage(message, field, index), layer + 1);
            break;
        default:
            proto_error("GenerateMultiple field unknow type, field=%s\n", field->full_name().c_str());
            break;
        }

        if (index < field_size - 1) {
            code << ", ";
        }
    }

    code << "}";
    return code.str();
}
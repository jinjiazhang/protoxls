#include <sstream>
#include "LuaExporter.h"

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
        {
            code << "[";
            code << store_key.num_key;
            code << "]";
        }
        else
        {
            code << "['";
            code << store_key.str_key;
            code << "']"; 
        }

        code << " = ";
        ConfigStore* sub_store = store->GetConfig(store_key);
        if (sub_store->HasStoreMap())
        {
            code << GenerateCode(sub_store, layer + 1);
        }
        else
        {
            code << GenerateCode(sub_store->GetData());
        }

        code << ",\n";
    }

    FLUSH_LAYER(code, layer);
    code << "}";
    return code.str();
}

string LuaExporter::GenerateCode(Message* data, int layer)
{
    std::stringstream code;
    code << "{...}";
    return code.str();
}
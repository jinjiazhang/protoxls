#include <fstream>
#include "BinExporter.h"

bool BinExporter::ExportResult(ConfigStore* store)
{
    const Descriptor* descriptor = store->GetDescriptor();
    const MessageOptions& option = descriptor->options();

    StoreScheme result;
    result.set_magic(164442955);
    result.set_scheme(descriptor->name());
    
    vector<string> key_names = store->GetKeyNames();
    for (int i = 0; i < key_names.size(); i++) {
        result.add_keys(key_names[i]);
    }

    vector<Message*> datas;
    store->ExportDatas(datas);
    for (int i = 0; i < datas.size(); i++)
    {
        string data = datas[i]->SerializeAsString();
        result.add_datas(data);
    }

    result.set_hash("");

    string output_text = option.GetExtension(output);
    if (output_text.empty()) {
        output_text = descriptor->name();
    }

    string file_name = output_text + ".bytes";
    std::ofstream file(file_name.c_str(), std::ios::binary);
    if (file.bad()) {
        proto_error("ExportResult open file fail, file=%s\n", file_name.c_str());
        return false;
    }

    result.SerializeToOstream(&file);
    file.close();
    return true;
}
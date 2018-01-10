#include "ConfigStore.h"


ConfigStore::ConfigStore(const Descriptor* descriptor)
{
    stores_ = NULL;
    key_names_ = NULL;
    descriptor_ = descriptor;
}

ConfigStore::~ConfigStore()
{
    descriptor_ = NULL;
    if (stores_ != NULL)
    {
        ConfigStoreMap::iterator it = stores_->begin();
        for (; it != stores_->end(); ++it) {
            delete it->second;
            it->second = NULL;
        }
        delete stores_;
        stores_ = NULL;
    }
}

void ConfigStore::ImportData(Message* data)
{
    datas_.push_back(data);
}

void ConfigStore::ImportData(vector<Message*> datas)
{
    datas_.insert(datas_.end(), datas.begin(), datas.end());
}

vector<string> ConfigStore::GetKeyNames()
{
    return *key_names_;
}

const Descriptor* ConfigStore::GetDescriptor()
{
    return descriptor_;
}

Message* ConfigStore::GetData()
{
    return datas_.front();    
}

ConfigStore* ConfigStore::GetConfig(int num_key)
{
    StoreKey store_key;
    store_key.key_type = KEY_TINTEGER;
    store_key.num_key = num_key;
    return GetConfig(store_key);
}

ConfigStore* ConfigStore::GetConfig(string str_key)
{
    StoreKey store_key;
    store_key.key_type = KEY_TSTRING;
    store_key.str_key = str_key;
    return GetConfig(store_key);
}

ConfigStore* ConfigStore::GetConfig(StoreKey store_key)
{
    ConfigStoreMap::iterator it = stores_->find(store_key);
    if (it == stores_->end()) {
        return NULL;
    }
    return it->second;
}

bool ConfigStore::HasChildren()
{
    return stores_ != NULL;
}

void ConfigStore::ExportKeys(vector<StoreKey>& keys)
{
    ConfigStoreMap::iterator it = stores_->begin();
    for (; it != stores_->end(); ++it)
    {
        keys.push_back(it->first);
    }
}

void ConfigStore::ExportDatas(vector<Message*>& datas)
{
    datas.insert(datas.end(), datas_.begin(), datas_.end());
}

bool ConfigStore::GetKeyVal(const Message& data, string key_name, StoreKey* store_key)
{
    const Reflection* reflection = data.GetReflection();
    const Descriptor* descriptor = data.GetDescriptor();
    const FieldDescriptor* field = descriptor->FindFieldByName(key_name);
    if (field == NULL) {
        proto_error("GetKeyVal key filed not found, key=%s, scheme=%s\n", key_name.c_str(), descriptor->full_name().c_str());
        return false;
    }

    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_INT32:
        store_key->key_type = KEY_TINTEGER;
        store_key->num_key = reflection->GetInt32(data, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT32:
        store_key->key_type = KEY_TINTEGER;
        store_key->num_key = reflection->GetUInt32(data, field);
        break;
    case FieldDescriptor::CPPTYPE_INT64:
        store_key->key_type = KEY_TINTEGER;
        store_key->num_key = reflection->GetInt64(data, field);
        break;
    case FieldDescriptor::CPPTYPE_UINT64:
        store_key->key_type = KEY_TINTEGER;
        store_key->num_key = reflection->GetUInt64(data, field);
        break;
    case FieldDescriptor::CPPTYPE_ENUM:
        store_key->key_type = KEY_TINTEGER;
        store_key->num_key = reflection->GetEnumValue(data, field);
        break;
    case FieldDescriptor::CPPTYPE_STRING:
        store_key->key_type = KEY_TSTRING;
        store_key->str_key = reflection->GetString(data, field);
        break;
    default:
        proto_error("GetKeyVal key filed unknow type, key=%s, scheme=%s\n", key_name.c_str(), descriptor->full_name().c_str());
        return false;
    }
    return true;
}

bool ConfigStore::BuildStore(vector<string> key_names)
{
    if (key_names.size() == 0) {
        proto_error("BuildStore keys empty");
        return false;
    }
    
    stores_ = new ConfigStoreMap();
    key_names_ = new vector<string>(key_names);

    string key_name = key_names.front();
    for (size_t i = 0; i < datas_.size(); i++)
    {
        Message* data = datas_[i];
        StoreKey store_key;
        if (!GetKeyVal(*data, key_name, &store_key)) {
            proto_error("BuildStore get keyval fail, key=%s\n", key_name.c_str());
            return false;
        }

        ConfigStoreMap::iterator it = stores_->find(store_key);
        if (it == stores_->end())
        {
            ConfigStore* store = new ConfigStore(descriptor_);
            store->ImportData(data);
            stores_->insert(std::make_pair(store_key, store));
        }
        else
        {
            it->second->ImportData(data);
        }
    }

    if (key_names.size() > 1) {
        vector<string> sub_keys;
        sub_keys.insert(sub_keys.end(), key_names.begin()+1, key_names.end());
        ConfigStoreMap::iterator it = stores_->begin();
        for (; it != stores_->end(); ++it)
        {
            if (!it->second->BuildStore(sub_keys)) {
                proto_error("BuildStore build sub store fail, key=%s\n", key_name.c_str());
                return false;
            }
        }
    }

    return true;
}
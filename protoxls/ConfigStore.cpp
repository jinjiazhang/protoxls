#include "ConfigStore.h"
#include "ParseHelper.h"


ConfigStore::ConfigStore(const Descriptor* descriptor)
{
    stores_ = NULL;
    descriptor_ = descriptor;
}

ConfigStore::~ConfigStore()
{
    descriptor_ = NULL;
    if (stores_ != NULL)
    {
        StoreMap::iterator it = stores_->begin();
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
    StoreMap::iterator it = stores_->find(store_key);
    if (it == stores_->end()) {
        return NULL;
    }
    return it->second;
}

bool ConfigStore::HasStoreMap()
{
    return stores_ != NULL;
}

void ConfigStore::ExportKeys(vector<StoreKey>& keys)
{
    StoreMap::iterator it = stores_->begin();
    for (; it != stores_->end(); ++it)
    {
        keys.push_back(it->first);
    }
}

bool ConfigStore::GetKeyVal(Message* data, string key_name, StoreKey* store_key)
{
    const Reflection* reflection = data->GetReflection();
    const Descriptor* descriptor = data->GetDescriptor();
    const FieldDescriptor* field = descriptor->FindFieldByName(key_name);
    if (field == NULL) {
        proto_error("GetKeyVal key filed not found, key=%s, scheme=%s\n", key_name.c_str(), descriptor->full_name().c_str());
        return false;
    }

    switch (field->cpp_type())
    {
    case FieldDescriptor::CPPTYPE_INT32:
    case FieldDescriptor::CPPTYPE_UINT32:
    case FieldDescriptor::CPPTYPE_INT64:
    case FieldDescriptor::CPPTYPE_UINT64:
    case FieldDescriptor::CPPTYPE_ENUM:
        store_key->key_type = KEY_TINTEGER;
        ParseHelper::GetNumberField(data, field, &store_key->num_key);
        break;
    case FieldDescriptor::CPPTYPE_STRING:
        store_key->key_type = KEY_TSTRING;
        ParseHelper::GetStringField(data, field, &store_key->str_key);
        break;
    default:
        proto_error("GetKeyVal key filed unknow type, key=%s, scheme=%s\n", key_name.c_str(), descriptor->full_name().c_str());
        return false;
    }
    return true;
}

bool ConfigStore::BuildStore(vector<string> keys)
{
    if (keys.size() == 0) {
        proto_error("BuildStore keys empty");
        return false;
    }
    
    stores_ = new StoreMap();
    string key_name = keys.front();
    for (int i = 0; i < datas_.size(); i++)
    {
        Message* data = datas_[i];
        StoreKey store_key;
        if (!GetKeyVal(data, key_name, &store_key)) {
            proto_error("BuildStore get keyval fail, key=%s\n", key_name.c_str());
            return false;
        }

        StoreMap::iterator it = stores_->find(store_key);
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

    if (keys.size() > 1) {
        vector<string> sub_keys;
        sub_keys.insert(sub_keys.end(), keys.begin()+1, keys.end());
        StoreMap::iterator it = stores_->begin();
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
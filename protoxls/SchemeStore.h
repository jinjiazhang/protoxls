#ifndef _JINJIAZHANG_SCHEMESTORE_H_
#define _JINJIAZHANG_SCHEMESTORE_H_

#include <fstream>
#include "option.pb.h"
#include "ConfigStore.h"

template<class SchemeClass>
class BaseStore {
public:
    bool LoadBytes(string file_name)
    {
        std::ifstream file(file_name.c_str(), std::ios::binary);
        if (file.bad()) {
            proto_error("LoadBytes open file fail, file=%s\n", file_name.c_str());
            return false;
        }

        StoreScheme result;
        result.ParseFromIstream(&file);
        file.close();

        store_ = new ConfigStore(sample_.GetDescriptor());
        for (int i = 0; i < result.datas_size(); i++)
        {
            SchemeClass* data = new SchemeClass();
            data->ParseFromString(result.datas(i));
            store_->ImportData(data);
        }

        vector<string> key_names;
        for (int i = 0; i < result.keys_size(); i++)
        {
            key_names.push_back(result.keys(i));
        }

        store_->BuildStore(key_names);
        return true;
    }

protected:
    ConfigStore* store_;
    SchemeClass sample_;
};

template<class SchemeClass, typename KeyType>
class SchemeStore1 : public BaseStore<SchemeClass>
{
public:
    const SchemeClass* GetConfig(KeyType key_val)
    {
        ConfigStore* sub_store = store_->GetConfig(key_val);
        if (sub_store == NULL) {
            return NULL;
        }
        return dynamic_cast<SchemeClass*>(sub_store->GetData());
    }
};

template<class SchemeClass, typename KeyType1, typename KeyType2>
class SchemeStore2 : public BaseStore<SchemeClass>
{
public:
    const SchemeClass* GetConfig(KeyType1 key_val1, KeyType2 key_val2)
    {
        ConfigStore* sub1_store = store_->GetConfig(key_val1);
        if (sub1_store == NULL) {
            return NULL;
        }
        ConfigStore* sub2_store = sub1_store->GetConfig(key_val2);
        if (sub2_store == NULL) {
            return NULL;
        }
        return dynamic_cast<SchemeClass*>(sub2_store->GetData());
    }
};

#endif
#ifndef _JINJIAZHANG_CONFIGSTORE_H_
#define _JINJIAZHANG_CONFIGSTORE_H_

#include "logger.h"
#include <google/protobuf/descriptor.h>
#include <google/protobuf/dynamic_message.h>

using namespace google::protobuf;
using namespace google::protobuf::io;
using namespace google::protobuf::compiler;
using namespace google::protobuf::internal;

#define KEY_TINTEGER 1
#define KEY_TSTRING 2

struct StoreKey {
    int key_type;
    int64 num_key;
    string str_key;

    bool operator <(const StoreKey& other) const
    {
        if (key_type == other.key_type)
        {
            if (key_type == KEY_TINTEGER)
                return num_key < other.num_key;
            else
                return str_key.compare(other.str_key) < 0;
        }
        return key_type < other.key_type;
    }
};

class ConfigStore {
    typedef std::map<StoreKey, ConfigStore*> StoreMap;
public:
    ConfigStore(const Descriptor* descriptor);
    ~ConfigStore();

public:
    void ImportData(Message* data);
    void ImportData(vector<Message*> datas);
    bool BuildStore(vector<string> keys);

public:
    Message* GetData();
    ConfigStore* GetConfig(int num_key);
    ConfigStore* GetConfig(string str_key);
    ConfigStore* GetConfig(StoreKey store_key);
    const Descriptor* GetDescriptor();
    bool  HasStoreMap();
    void  ExportKeys(vector<StoreKey>& keys);

private:
    bool GetKeyVal(Message* data, string key_name, StoreKey* store_key);

private:
    vector<Message*> datas_;
    StoreMap* stores_;
    const Descriptor* descriptor_;
};

#endif
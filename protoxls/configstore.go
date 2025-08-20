package protoxls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type StoreKeyType int

const (
	KeyTypeInteger StoreKeyType = iota + 1
	KeyTypeString
)

type StoreKey struct {
	KeyType  StoreKeyType
	NumKey   int64
	StrKey   string
}

func (k StoreKey) String() string {
	if k.KeyType == KeyTypeInteger {
		return fmt.Sprintf("%d", k.NumKey)
	}
	return k.StrKey
}

type ConfigStore struct {
	descriptor *desc.MessageDescriptor
	datas      []*dynamic.Message
	stores     map[StoreKey]*ConfigStore
	keyNames   []string
}

func NewConfigStore(descriptor *desc.MessageDescriptor) *ConfigStore {
	return &ConfigStore{
		descriptor: descriptor,
		datas:      make([]*dynamic.Message, 0),
		stores:     make(map[StoreKey]*ConfigStore),
		keyNames:   make([]string, 0),
	}
}

func (cs *ConfigStore) ImportData(data *dynamic.Message) {
	cs.datas = append(cs.datas, data)
}

func (cs *ConfigStore) ImportDatas(datas []*dynamic.Message) {
	cs.datas = append(cs.datas, datas...)
}

func (cs *ConfigStore) HasChildren() bool {
	return len(cs.stores) > 0
}

func (cs *ConfigStore) BuildStore(keyNames []string) error {
	if len(keyNames) == 0 {
		return nil
	}

	cs.keyNames = keyNames
	keyName := keyNames[0]
	remainingKeys := keyNames[1:]

	for _, data := range cs.datas {
		key, err := cs.getKeyVal(data, keyName)
		if err != nil {
			return fmt.Errorf("failed to get key value: %v", err)
		}

		store, exists := cs.stores[key]
		if !exists {
			store = NewConfigStore(cs.descriptor)
			cs.stores[key] = store
		}

		store.ImportData(data)
	}

	// Recursively build child stores
	for _, store := range cs.stores {
		if err := store.BuildStore(remainingKeys); err != nil {
			return err
		}
	}

	return nil
}

func (cs *ConfigStore) GetData() *dynamic.Message {
	if len(cs.datas) > 0 {
		return cs.datas[0]
	}
	return nil
}

func (cs *ConfigStore) GetConfig(key interface{}) *ConfigStore {
	var storeKey StoreKey
	switch v := key.(type) {
	case int:
		storeKey = StoreKey{KeyType: KeyTypeInteger, NumKey: int64(v)}
	case int32:
		storeKey = StoreKey{KeyType: KeyTypeInteger, NumKey: int64(v)}
	case int64:
		storeKey = StoreKey{KeyType: KeyTypeInteger, NumKey: v}
	case string:
		storeKey = StoreKey{KeyType: KeyTypeString, StrKey: v}
	default:
		return nil
	}

	return cs.stores[storeKey]
}

func (cs *ConfigStore) GetKeyNames() []string {
	return cs.keyNames
}

func (cs *ConfigStore) GetDescriptor() *desc.MessageDescriptor {
	return cs.descriptor
}

func (cs *ConfigStore) ExportKeys() []StoreKey {
	keys := make([]StoreKey, 0, len(cs.stores))
	for key := range cs.stores {
		keys = append(keys, key)
	}
	return keys
}

func (cs *ConfigStore) ExportDatas() []*dynamic.Message {
	return cs.datas
}

func (cs *ConfigStore) getKeyVal(data *dynamic.Message, keyName string) (StoreKey, error) {
	field := cs.descriptor.FindFieldByName(keyName)
	if field == nil {
		return StoreKey{}, fmt.Errorf("field %s not found", keyName)
	}

	value := data.GetField(field)
	if value == nil {
		return StoreKey{}, fmt.Errorf("field %s is nil", keyName)
	}

	switch field.GetType().String() {
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		if v, ok := value.(int32); ok {
			return StoreKey{KeyType: KeyTypeInteger, NumKey: int64(v)}, nil
		}
	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		if v, ok := value.(int64); ok {
			return StoreKey{KeyType: KeyTypeInteger, NumKey: v}, nil
		}
	case "TYPE_STRING":
		if v, ok := value.(string); ok {
			// Try to parse as number first
			if num, err := strconv.ParseInt(v, 10, 64); err == nil {
				return StoreKey{KeyType: KeyTypeInteger, NumKey: num}, nil
			}
			return StoreKey{KeyType: KeyTypeString, StrKey: v}, nil
		}
	}

	return StoreKey{}, fmt.Errorf("unsupported key type for field %s", keyName)
}

// Split utility function
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
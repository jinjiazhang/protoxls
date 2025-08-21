package protoxls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

// StoreKeyType defines the type of store key
type StoreKeyType int

const (
	// KeyTypeInteger represents integer keys
	KeyTypeInteger StoreKeyType = iota + 1
	// KeyTypeString represents string keys
	KeyTypeString
)

// StoreKey represents a key for indexing configuration data
type StoreKey struct {
	KeyType       StoreKeyType
	IntegerValue  int64
	StringValue   string
}

// String returns the string representation of the store key
func (k StoreKey) String() string {
	if k.KeyType == KeyTypeInteger {
		return fmt.Sprintf("%d", k.IntegerValue)
	}
	return k.StringValue
}

// DataStore manages configuration data with hierarchical storage
type DataStore struct {
	messageDescriptor *desc.MessageDescriptor
	messages          []*dynamic.Message
	childStores       map[StoreKey]*DataStore
	keyFieldNames     []string
}

// NewDataStore creates a new configuration store
func NewDataStore(messageDescriptor *desc.MessageDescriptor) *DataStore {
	return &DataStore{
		messageDescriptor: messageDescriptor,
		messages:          make([]*dynamic.Message, 0),
		childStores:       make(map[StoreKey]*DataStore),
		keyFieldNames:     make([]string, 0),
	}
}

// AddMessage adds a single message to the store
func (cs *DataStore) AddMessage(message *dynamic.Message) {
	cs.messages = append(cs.messages, message)
}

// AddMessages adds multiple messages to the store
func (cs *DataStore) AddMessages(messages []*dynamic.Message) {
	cs.messages = append(cs.messages, messages...)
}

// HasChildStores returns true if this store has child stores
func (cs *DataStore) HasChildStores() bool {
	return len(cs.childStores) > 0
}

// BuildHierarchicalStore builds a hierarchical store structure using the specified key fields
func (cs *DataStore) BuildHierarchicalStore(keyFieldNames []string) error {
	if len(keyFieldNames) == 0 {
		return nil
	}

	cs.keyFieldNames = keyFieldNames
	currentKeyField := keyFieldNames[0]
	remainingKeyFields := keyFieldNames[1:]

	for _, message := range cs.messages {
		key, err := cs.extractKeyFromMessage(message, currentKeyField)
		if err != nil {
			return fmt.Errorf("failed to extract key value from field %s: %v", currentKeyField, err)
		}

		childStore, exists := cs.childStores[key]
		if !exists {
			childStore = NewDataStore(cs.messageDescriptor)
			cs.childStores[key] = childStore
		}

		childStore.AddMessage(message)
	}

	// Recursively build child stores with remaining key fields
	for _, childStore := range cs.childStores {
		if err := childStore.BuildHierarchicalStore(remainingKeyFields); err != nil {
			return err
		}
	}

	return nil
}

// GetFirstMessage returns the first message in the store, or nil if empty
func (cs *DataStore) GetFirstMessage() *dynamic.Message {
	if len(cs.messages) > 0 {
		return cs.messages[0]
	}
	return nil
}

// GetChildStore returns a child store by key, or nil if not found
func (cs *DataStore) GetChildStore(key interface{}) *DataStore {
	storeKey, err := cs.convertToStoreKey(key)
	if err != nil {
		return nil
	}
	return cs.childStores[storeKey]
}

// convertToStoreKey converts various key types to StoreKey
func (cs *DataStore) convertToStoreKey(key interface{}) (StoreKey, error) {
	switch v := key.(type) {
	case StoreKey:
		return v, nil
	case int:
		return StoreKey{KeyType: KeyTypeInteger, IntegerValue: int64(v)}, nil
	case int32:
		return StoreKey{KeyType: KeyTypeInteger, IntegerValue: int64(v)}, nil
	case int64:
		return StoreKey{KeyType: KeyTypeInteger, IntegerValue: v}, nil
	case string:
		return StoreKey{KeyType: KeyTypeString, StringValue: v}, nil
	default:
		return StoreKey{}, fmt.Errorf("unsupported key type: %T", key)
	}
}

// GetKeyFieldNames returns the names of the key fields used for indexing
func (cs *DataStore) GetKeyFieldNames() []string {
	return cs.keyFieldNames
}

// GetMessageDescriptor returns the protobuf message descriptor
func (cs *DataStore) GetMessageDescriptor() *desc.MessageDescriptor {
	return cs.messageDescriptor
}

// GetAllKeys returns all store keys in this level
func (cs *DataStore) GetAllKeys() []StoreKey {
	keys := make([]StoreKey, 0, len(cs.childStores))
	for key := range cs.childStores {
		keys = append(keys, key)
	}
	return keys
}

// GetAllMessages returns all messages in this store
func (cs *DataStore) GetAllMessages() []*dynamic.Message {
	return cs.messages
}

// extractKeyFromMessage extracts a key value from a message field
func (cs *DataStore) extractKeyFromMessage(message *dynamic.Message, fieldName string) (StoreKey, error) {
	field := cs.messageDescriptor.FindFieldByName(fieldName)
	if field == nil {
		return StoreKey{}, fmt.Errorf("field %s not found in message descriptor", fieldName)
	}

	value := message.GetField(field)
	if value == nil {
		return StoreKey{}, fmt.Errorf("field %s has nil value", fieldName)
	}

	switch field.GetType().String() {
	case "TYPE_INT32", "TYPE_SINT32", "TYPE_SFIXED32", "TYPE_UINT32", "TYPE_FIXED32":
		if v, ok := value.(int32); ok {
			return StoreKey{KeyType: KeyTypeInteger, IntegerValue: int64(v)}, nil
		}
	case "TYPE_INT64", "TYPE_SINT64", "TYPE_SFIXED64", "TYPE_UINT64", "TYPE_FIXED64":
		if v, ok := value.(int64); ok {
			return StoreKey{KeyType: KeyTypeInteger, IntegerValue: v}, nil
		}
	case "TYPE_STRING":
		if v, ok := value.(string); ok {
			// Try to parse as number first for numeric string keys
			if num, err := strconv.ParseInt(v, 10, 64); err == nil {
				return StoreKey{KeyType: KeyTypeInteger, IntegerValue: num}, nil
			}
			return StoreKey{KeyType: KeyTypeString, StringValue: v}, nil
		}
	}

	return StoreKey{}, fmt.Errorf("unsupported key type %s for field %s", field.GetType().String(), fieldName)
}


// Split utility function for splitting delimited strings
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}
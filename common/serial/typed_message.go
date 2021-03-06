package serial

import (
	"errors"
	"reflect"

	"github.com/golang/protobuf/proto"
)

func ToTypedMessage(message proto.Message) *TypedMessage {
	if message == nil {
		return nil
	}
	settings, _ := proto.Marshal(message)
	return &TypedMessage{
		Type:  GetMessageType(message),
		Value: settings,
	}
}

func GetMessageType(message proto.Message) string {
	return proto.MessageName(message)
}

func GetInstance(messageType string) (interface{}, error) {
	mType := proto.MessageType(messageType).Elem()
	if mType == nil {
		return nil, errors.New("Unknown type: " + messageType)
	}
	return reflect.New(mType).Interface(), nil
}

func (v *TypedMessage) GetInstance() (interface{}, error) {
	instance, err := GetInstance(v.Type)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(v.Value, instance.(proto.Message)); err != nil {
		return nil, err
	}
	return instance, nil
}

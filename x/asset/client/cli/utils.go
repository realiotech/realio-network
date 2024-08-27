package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/gogo/protobuf/jsonpb"
)

// proposal defines the new Msg-based proposal.
type privilegeMsgContent struct {
	// Msg defines a sdk.Msgs proto-JSON-encoded as Any.
	Message     json.RawMessage `json:"message,omitempty"`
	MessageType string          `json:"message_type,omitempty"`
}

func parseMsgContent(path string) (*codectypes.Any, error) {
	var content privilegeMsgContent

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &content)
	if err != nil {
		return nil, err
	}

	return encodeJSONToProto(content.MessageType, content.Message)
}

func encodeJSONToProto(name string, jsonMsg []byte) (*codectypes.Any, error) {
	impl := proto.MessageType(name)
	if impl == nil {
		return nil, fmt.Errorf("message type %s not found", name)
	}
	msg := reflect.New(impl.Elem()).Interface().(proto.Message)
	err := jsonpb.Unmarshal(bytes.NewBuffer(jsonMsg), msg)
	if err != nil {
		return nil, fmt.Errorf("provided message is not valid %s: %w", jsonMsg, err)
	}
	return codectypes.NewAnyWithValue(msg)
}

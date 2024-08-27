package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/gogoproto/proto"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/realiotech/realio-network/x/asset/types"
)

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

type balances struct {
	Balances []json.RawMessage `json:"balances,omitempty"`
}

func parseBalances(cdc codec.Codec, path string) ([]types.Balance, error) {
	var rawBalances balances

	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, &rawBalances)
	if err != nil {
		return nil, err
	}

	balances := make([]types.Balance, len(rawBalances.Balances))
	for i, jsonMsg := range rawBalances.Balances {
		var balance types.Balance
		err := cdc.UnmarshalInterfaceJSON(jsonMsg, &balance)
		if err != nil {
			return nil, err
		}

		balances[i] = balance
	}

	return balances, nil
}

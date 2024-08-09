package keeper

import (
	"fmt"
	"reflect"
	"strings"

	types "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoiface"
)

type ProtoMsg = protoiface.MessageV1

func UnpackAnyMsg(anyPB *types.Any) (proto.Message, error) {
	split := strings.Split(anyPB.TypeUrl, "/")
	name := split[len(split)-1]
	typ := proto.MessageType(name)
	if typ == nil {
		return nil, fmt.Errorf("no message type found for %s", name)
	}
	to := reflect.New(typ.Elem()).Interface().(proto.Message)
	return to, UnpackAnyTo(anyPB, to)
}

func UnpackAnyTo(anyPB *types.Any, to ProtoMsg) error {
	return proto.Unmarshal(anyPB.Value, to)
}

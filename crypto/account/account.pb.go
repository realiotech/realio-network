// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ethermint/types/v1/account.proto

package account

import (
	types "github.com/cosmos/cosmos-sdk/x/auth/types"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// EthAccount implements the authtypes.AccountI interface and embeds an
// authtypes.BaseAccount type. It is compatible with the auth AccountKeeper.
type EthAccount struct {
	*types.BaseAccount `protobuf:"bytes,1,opt,name=base_account,json=baseAccount,proto3,embedded=base_account" json:"base_account,omitempty" yaml:"base_account"`
	CodeHash           string `protobuf:"bytes,2,opt,name=code_hash,json=codeHash,proto3" json:"code_hash,omitempty" yaml:"code_hash"`
}

func (m *EthAccount) Reset()      { *m = EthAccount{} }
func (*EthAccount) ProtoMessage() {}
func (*EthAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_4edc057d42a619ef, []int{0}
}
func (m *EthAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EthAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EthAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EthAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EthAccount.Merge(m, src)
}
func (m *EthAccount) XXX_Size() int {
	return m.Size()
}
func (m *EthAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_EthAccount.DiscardUnknown(m)
}

var xxx_messageInfo_EthAccount proto.InternalMessageInfo

func init() {
	proto.RegisterType((*EthAccount)(nil), "ethermint.types.v1.EthAccount")
}

func init() { proto.RegisterFile("ethermint/types/v1/account.proto", fileDescriptor_4edc057d42a619ef) }

var fileDescriptor_4edc057d42a619ef = []byte{
	// 319 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x51, 0x3f, 0x4f, 0xc2, 0x40,
	0x14, 0xef, 0x39, 0x18, 0x29, 0x0e, 0xa6, 0x32, 0x20, 0x9a, 0x6b, 0xd3, 0x89, 0x41, 0x7a, 0xa9,
	0x6c, 0x6c, 0x36, 0x31, 0xd1, 0xc1, 0x85, 0xd1, 0x05, 0xaf, 0xe7, 0x85, 0x23, 0x52, 0x1e, 0xe9,
	0x3d, 0x30, 0x7c, 0x03, 0x47, 0x47, 0x47, 0x3e, 0x84, 0x1f, 0xc2, 0xb8, 0xc8, 0xe8, 0x44, 0x0c,
	0x5d, 0x9c, 0xf9, 0x04, 0x86, 0xde, 0x89, 0x0e, 0x6e, 0xef, 0xf7, 0x7e, 0xff, 0x5e, 0xee, 0xdc,
	0x40, 0xa2, 0x92, 0x79, 0x36, 0x18, 0x21, 0xc3, 0xd9, 0x58, 0x6a, 0x36, 0x8d, 0x19, 0x17, 0x02,
	0x26, 0x23, 0x8c, 0xc6, 0x39, 0x20, 0x78, 0xde, 0x56, 0x11, 0x95, 0x8a, 0x68, 0x1a, 0x37, 0xa8,
	0x00, 0x9d, 0x81, 0x66, 0x7c, 0x82, 0x8a, 0x4d, 0xe3, 0x54, 0x22, 0x8f, 0x4b, 0x60, 0x3c, 0x8d,
	0x23, 0xc3, 0xf7, 0x4a, 0xc4, 0x0c, 0xb0, 0x54, 0xad, 0x0f, 0x7d, 0x30, 0xfb, 0xcd, 0x64, 0xb6,
	0xe1, 0x3b, 0x71, 0xdd, 0x0b, 0x54, 0xe7, 0xa6, 0xd9, 0xbb, 0x75, 0xf7, 0x53, 0xae, 0x65, 0xcf,
	0x5e, 0x52, 0x27, 0x01, 0x69, 0x56, 0xcf, 0x82, 0xc8, 0x26, 0x95, 0x4d, 0xb6, 0x36, 0x4a, 0xb8,
	0x96, 0xd6, 0x97, 0x1c, 0x2f, 0x96, 0x3e, 0x59, 0x2f, 0xfd, 0xc3, 0x19, 0xcf, 0x86, 0x9d, 0xf0,
	0x6f, 0x46, 0xd8, 0xad, 0xa6, 0xbf, 0x4a, 0x2f, 0x76, 0x2b, 0x02, 0xee, 0x64, 0x4f, 0x71, 0xad,
	0xea, 0x3b, 0x01, 0x69, 0x56, 0x92, 0xda, 0x7a, 0xe9, 0x1f, 0x18, 0xe3, 0x96, 0x0a, 0xbb, 0x7b,
	0x9b, 0xf9, 0x92, 0x6b, 0xd5, 0x39, 0x7d, 0x9c, 0xfb, 0xce, 0xf3, 0xdc, 0x77, 0xbe, 0xe6, 0xbe,
	0xf3, 0xf6, 0xd2, 0x3a, 0xf9, 0xef, 0x1a, 0x9b, 0x7f, 0x95, 0x5c, 0xbf, 0xae, 0x28, 0x59, 0xac,
	0x28, 0xf9, 0x5c, 0x51, 0xf2, 0x54, 0x50, 0x67, 0x51, 0x50, 0xe7, 0xa3, 0xa0, 0xce, 0x4d, 0xbb,
	0x3f, 0x40, 0x35, 0x49, 0x23, 0x01, 0x19, 0xcb, 0x25, 0x1f, 0x0e, 0x00, 0xa5, 0x50, 0x76, 0x6c,
	0x8d, 0x24, 0x3e, 0x40, 0x7e, 0xcf, 0x44, 0x3e, 0x1b, 0x23, 0xfc, 0xfc, 0x45, 0xba, 0x5b, 0xbe,
	0x53, 0xfb, 0x3b, 0x00, 0x00, 0xff, 0xff, 0xd8, 0xfb, 0x30, 0x3a, 0xb0, 0x01, 0x00, 0x00,
}

func (m *EthAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EthAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EthAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CodeHash) > 0 {
		i -= len(m.CodeHash)
		copy(dAtA[i:], m.CodeHash)
		i = encodeVarintAccount(dAtA, i, uint64(len(m.CodeHash)))
		i--
		dAtA[i] = 0x12
	}
	if m.BaseAccount != nil {
		{
			size, err := m.BaseAccount.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintAccount(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintAccount(dAtA []byte, offset int, v uint64) int {
	offset -= sovAccount(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EthAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BaseAccount != nil {
		l = m.BaseAccount.Size()
		n += 1 + l + sovAccount(uint64(l))
	}
	l = len(m.CodeHash)
	if l > 0 {
		n += 1 + l + sovAccount(uint64(l))
	}
	return n
}

func sovAccount(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozAccount(x uint64) (n int) {
	return sovAccount(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EthAccount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowAccount
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EthAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EthAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseAccount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAccount
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthAccount
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthAccount
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.BaseAccount == nil {
				m.BaseAccount = &types.BaseAccount{}
			}
			if err := m.BaseAccount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CodeHash", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowAccount
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthAccount
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthAccount
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CodeHash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipAccount(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthAccount
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipAccount(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowAccount
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAccount
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowAccount
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthAccount
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupAccount
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthAccount
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthAccount        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowAccount          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupAccount = fmt.Errorf("proto: unexpected end of group")
)

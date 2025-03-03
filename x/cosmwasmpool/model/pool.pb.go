// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/cosmwasmpool/v1beta1/model/pool.proto

package model

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/protobuf/types/known/timestamppb"
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

// CosmWasmPool represents the data serialized into state for each CW pool.
//
// Note: CW Pool has 2 pool models:
// - CosmWasmPool which is a proto-generated store model used for serialization
// into state.
// - Pool struct that encapsulates the CosmWasmPool and wasmKeeper for calling
// the contract.
//
// CosmWasmPool implements the poolmanager.PoolI interface but it panics on all
// methods. The reason is that access to wasmKeeper is required to call the
// contract.
//
// Instead, all interactions and poolmanager.PoolI methods are to be performed
// on the Pool struct. The reason why we cannot have a Pool struct only is
// because it cannot be serialized into state due to having a non-serializable
// wasmKeeper field.
type CosmWasmPool struct {
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty" yaml:"contract_address"`
	PoolId          uint64 `protobuf:"varint,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	CodeId          uint64 `protobuf:"varint,3,opt,name=code_id,json=codeId,proto3" json:"code_id,omitempty"`
	InstantiateMsg  []byte `protobuf:"bytes,4,opt,name=instantiate_msg,json=instantiateMsg,proto3" json:"instantiate_msg,omitempty" yaml:"instantiate_msg"`
}

func (m *CosmWasmPool) Reset()      { *m = CosmWasmPool{} }
func (*CosmWasmPool) ProtoMessage() {}
func (*CosmWasmPool) Descriptor() ([]byte, []int) {
	return fileDescriptor_a0cb64564a744af1, []int{0}
}
func (m *CosmWasmPool) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CosmWasmPool) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CosmWasmPool.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CosmWasmPool) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CosmWasmPool.Merge(m, src)
}
func (m *CosmWasmPool) XXX_Size() int {
	return m.Size()
}
func (m *CosmWasmPool) XXX_DiscardUnknown() {
	xxx_messageInfo_CosmWasmPool.DiscardUnknown(m)
}

var xxx_messageInfo_CosmWasmPool proto.InternalMessageInfo

func init() {
	proto.RegisterType((*CosmWasmPool)(nil), "osmosis.cosmwasmpool.v1beta1.CosmWasmPool")
}

func init() {
	proto.RegisterFile("osmosis/cosmwasmpool/v1beta1/model/pool.proto", fileDescriptor_a0cb64564a744af1)
}

var fileDescriptor_a0cb64564a744af1 = []byte{
	// 351 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x91, 0xcd, 0x4e, 0xf2, 0x40,
	0x14, 0x86, 0x3b, 0xdf, 0x87, 0x18, 0x1b, 0x22, 0xda, 0x18, 0x41, 0x34, 0x2d, 0xe9, 0x8a, 0x0d,
	0x9d, 0x80, 0x1b, 0xc3, 0x4e, 0x48, 0x4c, 0x58, 0x98, 0x98, 0x6e, 0x4c, 0xdc, 0x34, 0xd3, 0x1f,
	0x6b, 0x93, 0x0e, 0x87, 0x70, 0x06, 0xd4, 0x3b, 0x70, 0xe9, 0xd2, 0x25, 0x17, 0xe1, 0x45, 0x18,
	0x57, 0x2c, 0x5d, 0x11, 0x02, 0x77, 0xc0, 0x15, 0x98, 0xe9, 0x94, 0x04, 0xd9, 0xf5, 0x7d, 0xde,
	0xa7, 0x99, 0x99, 0x73, 0xf4, 0x26, 0x20, 0x07, 0x4c, 0x90, 0x06, 0x80, 0xfc, 0x99, 0x21, 0x1f,
	0x02, 0xa4, 0x74, 0xd2, 0xf2, 0x23, 0xc1, 0x5a, 0x94, 0x43, 0x18, 0xa5, 0x54, 0x22, 0x67, 0x38,
	0x02, 0x01, 0xc6, 0x45, 0xae, 0x3b, 0xdb, 0xba, 0x93, 0xeb, 0xb5, 0xb3, 0x20, 0xab, 0xbd, 0xcc,
	0xa5, 0x2a, 0xa8, 0x1f, 0x6b, 0x27, 0x31, 0xc4, 0xa0, 0xb8, 0xfc, 0xca, 0xa9, 0x15, 0x03, 0xc4,
	0x69, 0x44, 0xb3, 0xe4, 0x8f, 0x1f, 0xa9, 0x48, 0x78, 0x84, 0x82, 0xf1, 0xa1, 0x12, 0xec, 0x05,
	0xd1, 0x4b, 0x3d, 0x40, 0x7e, 0xcf, 0x90, 0xdf, 0x01, 0xa4, 0xc6, 0x8d, 0x7e, 0x14, 0xc0, 0x40,
	0x8c, 0x58, 0x20, 0x3c, 0x16, 0x86, 0xa3, 0x08, 0xb1, 0x4a, 0xea, 0xa4, 0x71, 0xd0, 0x3d, 0x5f,
	0xcf, 0xad, 0xca, 0x2b, 0xe3, 0x69, 0xc7, 0xde, 0x35, 0x6c, 0xb7, 0xbc, 0x41, 0xd7, 0x8a, 0x18,
	0x15, 0x7d, 0x5f, 0x5e, 0xdd, 0x4b, 0xc2, 0xea, 0xbf, 0x3a, 0x69, 0x14, 0xdc, 0xa2, 0x8c, 0xfd,
	0x50, 0x16, 0x01, 0x84, 0x91, 0x2c, 0xfe, 0xab, 0x42, 0xc6, 0x7e, 0x68, 0xf4, 0xf4, 0x72, 0x32,
	0x40, 0xc1, 0x06, 0x22, 0x61, 0x22, 0xf2, 0x38, 0xc6, 0xd5, 0x42, 0x9d, 0x34, 0x4a, 0xdd, 0xda,
	0x7a, 0x6e, 0x9d, 0xaa, 0x83, 0x77, 0x04, 0xdb, 0x3d, 0xdc, 0x22, 0xb7, 0x18, 0x77, 0x8e, 0xdf,
	0xa6, 0x96, 0xf6, 0x31, 0xb5, 0xb4, 0xef, 0xcf, 0xe6, 0x9e, 0x7c, 0x50, 0xbf, 0xeb, 0x7e, 0x2d,
	0x4d, 0x32, 0x5b, 0x9a, 0x64, 0xb1, 0x34, 0xc9, 0xfb, 0xca, 0xd4, 0x66, 0x2b, 0x53, 0xfb, 0x59,
	0x99, 0xda, 0xc3, 0x55, 0x9c, 0x88, 0xa7, 0xb1, 0xef, 0x04, 0xc0, 0x69, 0x3e, 0xf7, 0x66, 0xca,
	0x7c, 0xdc, 0x04, 0x3a, 0x69, 0xb7, 0xe9, 0xcb, 0xdf, 0xcd, 0x65, 0x1b, 0xf3, 0x8b, 0xd9, 0xf4,
	0x2e, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x02, 0x9f, 0xd8, 0xf6, 0xde, 0x01, 0x00, 0x00,
}

func (m *CosmWasmPool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CosmWasmPool) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CosmWasmPool) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.InstantiateMsg) > 0 {
		i -= len(m.InstantiateMsg)
		copy(dAtA[i:], m.InstantiateMsg)
		i = encodeVarintPool(dAtA, i, uint64(len(m.InstantiateMsg)))
		i--
		dAtA[i] = 0x22
	}
	if m.CodeId != 0 {
		i = encodeVarintPool(dAtA, i, uint64(m.CodeId))
		i--
		dAtA[i] = 0x18
	}
	if m.PoolId != 0 {
		i = encodeVarintPool(dAtA, i, uint64(m.PoolId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintPool(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintPool(dAtA []byte, offset int, v uint64) int {
	offset -= sovPool(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CosmWasmPool) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovPool(uint64(l))
	}
	if m.PoolId != 0 {
		n += 1 + sovPool(uint64(m.PoolId))
	}
	if m.CodeId != 0 {
		n += 1 + sovPool(uint64(m.CodeId))
	}
	l = len(m.InstantiateMsg)
	if l > 0 {
		n += 1 + l + sovPool(uint64(l))
	}
	return n
}

func sovPool(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPool(x uint64) (n int) {
	return sovPool(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CosmWasmPool) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPool
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
			return fmt.Errorf("proto: CosmWasmPool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CosmWasmPool: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PoolId", wireType)
			}
			m.PoolId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PoolId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CodeId", wireType)
			}
			m.CodeId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CodeId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InstantiateMsg", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.InstantiateMsg = append(m.InstantiateMsg[:0], dAtA[iNdEx:postIndex]...)
			if m.InstantiateMsg == nil {
				m.InstantiateMsg = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPool(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPool
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
func skipPool(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPool
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
					return 0, ErrIntOverflowPool
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
					return 0, ErrIntOverflowPool
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
				return 0, ErrInvalidLengthPool
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPool
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPool
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPool        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPool          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPool = fmt.Errorf("proto: unexpected end of group")
)

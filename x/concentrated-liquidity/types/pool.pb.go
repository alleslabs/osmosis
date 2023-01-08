// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/concentrated-liquidity/types/pool.proto

// This is a legacy package that requires additional migration logic
// in order to use the correct package. Decision made to use legacy package path
// until clear steps for migration logic and the unknowns for state breaking are
// investigated for changing proto package.

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type PoolIncentivizedLiquidityRecord struct {
	ID                                    uint64                                 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	MinimumFreezeDuration                 time.Duration                          `protobuf:"bytes,2,opt,name=minimum_freeze_duration,json=minimumFreezeDuration,proto3,stdduration" json:"minimum_freeze_duration" yaml:"minimum_freeze_duration"`
	SecondsPerIncentivizedLiquidityGlobal github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=seconds_per_incentivized_liquidity_global,json=secondsPerIncentivizedLiquidityGlobal,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"seconds_per_incentivized_liquidity_global" yaml:"seconds_per_incentivized_liquidity_global"`
}

func (m *PoolIncentivizedLiquidityRecord) Reset()         { *m = PoolIncentivizedLiquidityRecord{} }
func (m *PoolIncentivizedLiquidityRecord) String() string { return proto.CompactTextString(m) }
func (*PoolIncentivizedLiquidityRecord) ProtoMessage()    {}
func (*PoolIncentivizedLiquidityRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_8cb52e0aea558b6c, []int{0}
}
func (m *PoolIncentivizedLiquidityRecord) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PoolIncentivizedLiquidityRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PoolIncentivizedLiquidityRecord.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PoolIncentivizedLiquidityRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoolIncentivizedLiquidityRecord.Merge(m, src)
}
func (m *PoolIncentivizedLiquidityRecord) XXX_Size() int {
	return m.Size()
}
func (m *PoolIncentivizedLiquidityRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_PoolIncentivizedLiquidityRecord.DiscardUnknown(m)
}

var xxx_messageInfo_PoolIncentivizedLiquidityRecord proto.InternalMessageInfo

func (m *PoolIncentivizedLiquidityRecord) GetID() uint64 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *PoolIncentivizedLiquidityRecord) GetMinimumFreezeDuration() time.Duration {
	if m != nil {
		return m.MinimumFreezeDuration
	}
	return 0
}

func init() {
	proto.RegisterType((*PoolIncentivizedLiquidityRecord)(nil), "osmosis.concentratedliquidity.v1beta1.PoolIncentivizedLiquidityRecord")
}

func init() {
	proto.RegisterFile("osmosis/concentrated-liquidity/types/pool.proto", fileDescriptor_8cb52e0aea558b6c)
}

var fileDescriptor_8cb52e0aea558b6c = []byte{
	// 402 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x4d, 0xab, 0xd3, 0x40,
	0x14, 0xcd, 0xc4, 0x87, 0x60, 0x04, 0x17, 0x41, 0xb1, 0xef, 0x2d, 0x26, 0x25, 0xf0, 0xa4, 0x0a,
	0x9d, 0xb1, 0xbe, 0x5d, 0x97, 0x21, 0x28, 0x05, 0x17, 0x25, 0x4b, 0x11, 0x42, 0x3e, 0xa6, 0x71,
	0x70, 0x92, 0x1b, 0x33, 0x93, 0x62, 0x0b, 0xfe, 0x07, 0x97, 0x2e, 0xfd, 0x21, 0xfe, 0x80, 0x2e,
	0xbb, 0x14, 0x17, 0x51, 0xda, 0x8d, 0xeb, 0x82, 0x7b, 0x69, 0x32, 0x29, 0x5d, 0xa8, 0xb8, 0x4a,
	0xee, 0xbd, 0xe7, 0x9e, 0x73, 0x38, 0x73, 0x2d, 0x0a, 0x32, 0x07, 0xc9, 0x25, 0x4d, 0xa0, 0x48,
	0x58, 0xa1, 0xaa, 0x48, 0xb1, 0x74, 0x2c, 0xf8, 0xbb, 0x9a, 0xa7, 0x5c, 0xad, 0xa8, 0x5a, 0x95,
	0x4c, 0xd2, 0x12, 0x40, 0x90, 0xb2, 0x02, 0x05, 0xf6, 0xb5, 0x5e, 0x20, 0xe7, 0x0b, 0x27, 0x3c,
	0x59, 0x4e, 0x62, 0xa6, 0xa2, 0xc9, 0x15, 0xce, 0x00, 0x32, 0xc1, 0x68, 0xbb, 0x14, 0xd7, 0x0b,
	0x9a, 0xd6, 0x55, 0xa4, 0x38, 0x14, 0x1d, 0xcd, 0xd5, 0x65, 0xd2, 0xf2, 0x84, 0x6d, 0x45, 0xbb,
	0x42, 0x8f, 0xee, 0x67, 0x90, 0x41, 0xd7, 0x3f, 0xfe, 0x75, 0x5d, 0xf7, 0x97, 0x69, 0x39, 0x73,
	0x00, 0x31, 0x6b, 0x55, 0xf9, 0x92, 0xaf, 0x59, 0xfa, 0xb2, 0x97, 0x0d, 0x58, 0x02, 0x55, 0x6a,
	0xdf, 0xb3, 0xcc, 0x99, 0x3f, 0x40, 0x43, 0x34, 0xba, 0x08, 0xcc, 0x99, 0x6f, 0x7f, 0xb0, 0x1e,
	0xe6, 0xbc, 0xe0, 0x79, 0x9d, 0x87, 0x8b, 0x8a, 0xb1, 0x35, 0x0b, 0x7b, 0x17, 0x03, 0x73, 0x88,
	0x46, 0x77, 0x9f, 0x5d, 0x92, 0xce, 0x26, 0xe9, 0x6d, 0x12, 0x5f, 0x03, 0xbc, 0x27, 0x9b, 0xc6,
	0x31, 0x0e, 0x8d, 0x83, 0x57, 0x51, 0x2e, 0xa6, 0xee, 0x5f, 0x78, 0xdc, 0x4f, 0xdf, 0x1d, 0x14,
	0x3c, 0xd0, 0xd3, 0xe7, 0xed, 0xb0, 0xa7, 0xb0, 0xbf, 0x20, 0xeb, 0xb1, 0x64, 0x09, 0x14, 0xa9,
	0x0c, 0x4b, 0x56, 0x85, 0xfc, 0xcc, 0x7a, 0x78, 0x8a, 0x2c, 0xcc, 0x04, 0xc4, 0x91, 0x18, 0xdc,
	0x1a, 0xa2, 0xd1, 0x1d, 0x2f, 0x3e, 0xca, 0x7e, 0x6b, 0x9c, 0x47, 0x19, 0x57, 0x6f, 0xea, 0x98,
	0x24, 0x90, 0xeb, 0x74, 0xf4, 0x67, 0x2c, 0xd3, 0xb7, 0xdd, 0xab, 0x10, 0x9f, 0x25, 0x87, 0xc6,
	0x79, 0xda, 0x19, 0xfc, 0x6f, 0x21, 0x37, 0xb8, 0xd6, 0xd8, 0x39, 0xab, 0xfe, 0x98, 0xe6, 0x8b,
	0x16, 0x37, 0xbd, 0xf8, 0xf9, 0xd9, 0x41, 0xde, 0xeb, 0xcd, 0x0e, 0xa3, 0xed, 0x0e, 0xa3, 0x1f,
	0x3b, 0x8c, 0x3e, 0xee, 0xb1, 0xb1, 0xdd, 0x63, 0xe3, 0xeb, 0x1e, 0x1b, 0xaf, 0xbc, 0x33, 0x8b,
	0xfa, 0x28, 0xc6, 0x22, 0x8a, 0xe5, 0xe9, 0xa4, 0x96, 0x93, 0x1b, 0xfa, 0xfe, 0x9f, 0x87, 0x15,
	0xdf, 0x6e, 0x83, 0xbf, 0xf9, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xd5, 0x3b, 0x57, 0xba, 0x87, 0x02,
	0x00, 0x00,
}

func (this *PoolIncentivizedLiquidityRecord) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*PoolIncentivizedLiquidityRecord)
	if !ok {
		that2, ok := that.(PoolIncentivizedLiquidityRecord)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.ID != that1.ID {
		return false
	}
	if this.MinimumFreezeDuration != that1.MinimumFreezeDuration {
		return false
	}
	if !this.SecondsPerIncentivizedLiquidityGlobal.Equal(that1.SecondsPerIncentivizedLiquidityGlobal) {
		return false
	}
	return true
}
func (m *PoolIncentivizedLiquidityRecord) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PoolIncentivizedLiquidityRecord) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PoolIncentivizedLiquidityRecord) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.SecondsPerIncentivizedLiquidityGlobal.Size()
		i -= size
		if _, err := m.SecondsPerIncentivizedLiquidityGlobal.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintPool(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	n1, err1 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.MinimumFreezeDuration, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.MinimumFreezeDuration):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintPool(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x12
	if m.ID != 0 {
		i = encodeVarintPool(dAtA, i, uint64(m.ID))
		i--
		dAtA[i] = 0x8
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
func (m *PoolIncentivizedLiquidityRecord) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ID != 0 {
		n += 1 + sovPool(uint64(m.ID))
	}
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.MinimumFreezeDuration)
	n += 1 + l + sovPool(uint64(l))
	l = m.SecondsPerIncentivizedLiquidityGlobal.Size()
	n += 1 + l + sovPool(uint64(l))
	return n
}

func sovPool(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPool(x uint64) (n int) {
	return sovPool(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PoolIncentivizedLiquidityRecord) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: PoolIncentivizedLiquidityRecord: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PoolIncentivizedLiquidityRecord: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ID", wireType)
			}
			m.ID = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ID |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinimumFreezeDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPool
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
				return ErrInvalidLengthPool
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPool
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.MinimumFreezeDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SecondsPerIncentivizedLiquidityGlobal", wireType)
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
			if err := m.SecondsPerIncentivizedLiquidityGlobal.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
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

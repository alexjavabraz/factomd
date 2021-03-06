// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

type Buffer struct {
	bytes.Buffer
}

func (b *Buffer) DeepCopyBytes() []byte {
	return b.Next(b.Len())
}

func NewBuffer(buf []byte) *Buffer {
	tmp := new(Buffer)
	tmp.Buffer = *bytes.NewBuffer(buf)
	return tmp
}

func (b *Buffer) PushBinaryMarshallable(bm interfaces.BinaryMarshallable) error {
	bin, err := bm.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = b.Write(bin)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PushString(s string) error {
	return b.PushBytes([]byte(s))
}

func (b *Buffer) PushBytes(h []byte) error {
	l := uint64(len(h))
	err := EncodeVarInt(b, l)
	if err != nil {
		return err
	}

	_, err = b.Write(h)
	if err != nil {
		return err
	}

	return nil
}

func (b *Buffer) Push(h []byte) error {
	_, err := b.Write(h)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PushUInt32(i uint32) error {
	return binary.Write(b, binary.BigEndian, &i)
}

func (b *Buffer) PushUInt64(i uint64) error {
	return binary.Write(b, binary.BigEndian, &i)
}

func (b *Buffer) PushBool(boo bool) error {
	var err error
	if boo {
		_, err = b.Write([]byte{0x01})
	} else {
		_, err = b.Write([]byte{0x00})
	}
	return err
}

func (b *Buffer) PushVarInt(vi uint64) error {
	return EncodeVarInt(b, vi)
}

func (b *Buffer) PushByte(h byte) error {
	return b.WriteByte(h)
}

func (b *Buffer) PushInt64(i int64) error {
	return b.PushUInt64(uint64(i))
}

func (b *Buffer) PushInt(i int) error {
	return b.PushInt64(int64(i))
}

func (b *Buffer) PopInt() (int, error) {
	i, err := b.PopInt64()
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func (b *Buffer) PopInt64() (int64, error) {
	i, err := b.PopUInt64()
	if err != nil {
		return 0, err
	}
	return int64(i), nil
}

func (b *Buffer) PopByte() (byte, error) {
	return b.ReadByte()
}

func (b *Buffer) PopVarInt() (uint64, error) {
	h := b.DeepCopyBytes()
	l, rest := DecodeVarInt(h)
	b.Reset()
	_, err := b.Write(rest)
	if err != nil {
		return 0, err
	}
	return l, nil
}

func (b *Buffer) PopUInt32() (uint32, error) {
	var i uint32
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *Buffer) PopUInt64() (uint64, error) {
	var i uint64
	err := binary.Read(b, binary.BigEndian, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (b *Buffer) PopBool() (bool, error) {
	boo, err := b.ReadByte()
	if err != nil {
		return false, err
	}
	return boo > 0, nil
}

func (b *Buffer) PopString() (string, error) {
	h, err := b.PopBytes()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", h), nil
}

func (b *Buffer) PopBytes() ([]byte, error) {
	h := b.DeepCopyBytes()
	l, rest := DecodeVarInt(h)

	answer := make([]byte, int(l))
	copy(answer, rest)
	remainder := rest[int(l):]

	b.Reset()
	_, err := b.Write(remainder)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (b *Buffer) PopLen(l int) ([]byte, error) {
	answer := make([]byte, l)
	_, err := b.Read(answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func (b *Buffer) Pop(h []byte) error {
	_, err := b.Read(h)
	if err != nil {
		return err
	}
	return nil
}

func (b *Buffer) PopBinaryMarshallable(dst interfaces.BinaryMarshallable) error {
	h := b.DeepCopyBytes()
	rest, err := dst.UnmarshalBinaryData(h)
	if err != nil {
		return err
	}

	b.Reset()
	_, err = b.Write(rest)
	if err != nil {
		return err
	}
	return nil
}

// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//A placeholder structure for messages
type DirectoryBlockSignature struct {
	MessageBase
	Timestamp interfaces.Timestamp
	DBHeight  uint32
	//DirectoryBlockKeyMR   interfaces.IHash
	DirectoryBlockHeader  interfaces.IDirectoryBlockHeader
	ServerIdentityChainID interfaces.IHash

	Signature interfaces.IFullSignature

	// Signature that goes into the admin block
	// Signature of directory block header
	DBSignature interfaces.IFullSignature

	//Not marshalled
	Processed bool
	hash      interfaces.IHash
}

var _ interfaces.IMsg = (*DirectoryBlockSignature)(nil)
var _ Signable = (*DirectoryBlockSignature)(nil)

func (a *DirectoryBlockSignature) IsSameAs(b *DirectoryBlockSignature) bool {
	if b == nil {
		return false
	}

	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if a.DBHeight != b.DBHeight {
		return false
	}

	if a.DirectoryBlockHeader == nil && b.DirectoryBlockHeader != nil {
		return false
	}
	if a.DirectoryBlockHeader != nil {
		if a.DirectoryBlockHeader.GetPrevFullHash().IsSameAs(b.DirectoryBlockHeader.GetPrevFullHash()) == false {
			return false
		}
	}

	if a.ServerIdentityChainID == nil && b.ServerIdentityChainID != nil {
		return false
	}
	if a.ServerIdentityChainID != nil {
		if a.ServerIdentityChainID.IsSameAs(b.ServerIdentityChainID) == false {
			return false
		}
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (e *DirectoryBlockSignature) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessDBSig(dbheight, e)
}

func (m *DirectoryBlockSignature) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DirectoryBlockSignature) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *DirectoryBlockSignature) GetMsgHash() interfaces.IHash {
	data, _ := m.MarshalForSignature()
	if data == nil {
		return nil
	}
	m.MsgHash = primitives.Sha(data)

	return m.MsgHash
}

func (m *DirectoryBlockSignature) GetTimestamp() interfaces.Timestamp {
	if m.Timestamp == nil {
		m.Timestamp = new(primitives.Timestamp)
	}
	return m.Timestamp
}

func (m *DirectoryBlockSignature) Type() byte {
	return constants.DIRECTORY_BLOCK_SIGNATURE_MSG
}

func (m *DirectoryBlockSignature) Int() int {
	return -1
}

func (m *DirectoryBlockSignature) Bytes() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *DirectoryBlockSignature) Validate(state interfaces.IState) int {
	found, _ := state.GetVirtualServers(m.DBHeight, 9, m.ServerIdentityChainID)

	if found == false {
		return 0
	}

	if m.IsLocal() {
		return 1
	}

	// *********************************  NEEDS FIXED **************
	// Need to check the signature for real. TODO:

	isVer, err := m.VerifySignature()
	if err != nil || !isVer {
		// if there is an error during signature verification
		// or if the signature is invalid
		// the message is considered invalid
		return -1
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *DirectoryBlockSignature) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *DirectoryBlockSignature) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *DirectoryBlockSignature) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (m *DirectoryBlockSignature) Sign(key interfaces.Signer) error {
	// Signature that goes into admin block
	err := m.SignHeader(key)
	if err != nil {
		return err
	}

	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *DirectoryBlockSignature) SignHeader(key interfaces.Signer) error {
	header, err := m.DirectoryBlockHeader.MarshalBinary()
	if err != nil {
		return err
	}
	m.DBSignature = key.Sign(header)
	return nil
}

func (m *DirectoryBlockSignature) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *DirectoryBlockSignature) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *DirectoryBlockSignature) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	// TimeStamp
	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.VMIndex, newData = int(newData[0]), newData[1:]

	header := directoryBlock.NewDBlockHeader()
	newData, err = header.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.DirectoryBlockHeader = header

	hash := new(primitives.Hash)
	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.ServerIdentityChainID = hash

	//if len(newData) > 0 {
	sig := new(primitives.Signature)
	newData, err = sig.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.DBSignature = sig
	//}

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	return nil, nil
}

func (m *DirectoryBlockSignature) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *DirectoryBlockSignature) MarshalForSignature() ([]byte, error) {
	if m.DirectoryBlockHeader == nil {
		m.DirectoryBlockHeader = directoryBlock.NewDBlockHeader()
	}

	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, m.DBHeight)
	binary.Write(&buf, binary.BigEndian, byte(m.VMIndex))

	header, err := m.DirectoryBlockHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(header)

	hash, err := m.ServerIdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(hash)

	if m.DBSignature != nil {
		dbSig, err := m.DBSignature.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(dbSig)
	} else {
		blankSig := make([]byte, constants.SIGNATURE_LENGTH)
		buf.Write(blankSig)
		buf.Write(blankSig[:32])
	}

	return buf.DeepCopyBytes(), nil
}

func (m *DirectoryBlockSignature) MarshalBinary() (data []byte, err error) {

	var sig interfaces.IFullSignature
	resp, err := m.MarshalForSignature()
	if err == nil {
		sig = m.GetSignature()
	}

	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return resp, nil
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
}

func (m *DirectoryBlockSignature) String() string {
	return fmt.Sprintf("%6s-VM%3d:          DBHt:%5d -- Signer[:3]=%x PrevDBKeyMR[:3]=%x hash[:3]=%x",
		"DBSig",
		m.VMIndex,
		m.DBHeight,
		m.ServerIdentityChainID.Bytes()[:3],
		m.DirectoryBlockHeader.GetPrevKeyMR().Bytes()[:3],
		m.GetHash().Bytes()[:3])

}

func (e *DirectoryBlockSignature) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DirectoryBlockSignature) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DirectoryBlockSignature) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
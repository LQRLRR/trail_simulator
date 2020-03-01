package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"trail_simulator/simulator/src/types"
)

// TXO cpntains the balance ,its owner, hash of parent block  of the block includes this TXO, and leaf index of TXOTree.
type TXO struct {
	Index           types.Uint256 // index of leaf nodes assigned this TXO.
	ParentBlockHash [32]byte      // hash of parent block  of the block includes this TXO.
	OwnerAddress    uint32
	Balance         uint64
}

// NewTXOWithoutIndex provide new TXO instance without Index and ParentBlockHash.
func NewTXOWithoutIndex(blockHash [32]byte, address uint32, balance uint64) *TXO {
	return &TXO{ParentBlockHash: blockHash, OwnerAddress: address, Balance: balance}
}

// SetIndex sets Index into TXO
func (u *TXO) SetIndex(index types.Uint256) {
	u.Index = index
}

// Hash returns SHA256 hash of TXO.
// If TXO is unused, the hash value is SHA256(byte(TXO)).
// If TXO is used, the hash value is SHA256(byte(TXO) + byte(TXO)).
func (u TXO) Hash(isUsed bool) [32]byte {
	var binBuf bytes.Buffer
	binary.Write(&binBuf, binary.BigEndian, u)
	utxoBytes := binBuf.Bytes()
	if isUsed {
		utxoBytes = append(utxoBytes, utxoBytes...)
	}
	return sha256.Sum256(utxoBytes)
}

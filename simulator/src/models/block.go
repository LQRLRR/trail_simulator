package models

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"trail_simulator/simulator/src/types"
)

// Block contains block information.
type Block struct {
	Parent         [32]byte      // SHA256 hash of parent Block.
	Height         uint64        // the length of the block chain from the genesis Block to this Block.
	Root           [32]byte      // root of TXO tree
	RightmostIndex types.Uint256 // latest (rightmost) leaf node index.
	RightmostHash  [32]byte      // hash value of latest (rightmost) leaf node
	RightmostProof [255][32]byte // merkle proof of latest (rightmost) leaf node.
}

// NewBlock provides new block instance.
func NewBlock(parentHash [32]byte, height uint64, root [32]byte, rightmostIndex types.Uint256, rightmostHash [32]byte, rightmostProof [255][32]byte) *Block {
	return &Block{parentHash, height, root, rightmostIndex, rightmostHash, rightmostProof}
}

// Hash returns SHA256 hash of Block.
func (b *Block) Hash() [32]byte {
	var binBuf bytes.Buffer
	binary.Write(&binBuf, binary.BigEndian, b)
	blockBytes := binBuf.Bytes()
	return sha256.Sum256(blockBytes)
}

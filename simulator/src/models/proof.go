package models

import (
	"crypto/sha256"
	"trail_simulator/simulator/src/types"
)

// Proof contains an txo and its merkle proof.
type Proof struct {
	TXO    *TXO
	Proofs [255][32]byte
}

// NewProof provide new proof instance.
func NewProof(txo *TXO, proofs [255][32]byte) *Proof {
	return &Proof{txo, proofs}
}

// Root returns the root calucurated from a leaf and its merkle proof.
func (p Proof) Root(isUsed bool) [32]byte {
	hash := p.TXO.Hash(isUsed)
	index := p.TXO.Index
	for h := uint8(0); h < uint8(255); h++ {
		if index[0]%2 == 0 {
			s := append(hash[:], p.Proofs[h][:]...)
			hash = sha256.Sum256(s)
		} else {
			s := append(p.Proofs[h][:], hash[:]...)
			hash = sha256.Sum256(s)
		}
		index = index.Divide2()
	}
	return hash
}

func getProofBranchIDs(leafIndex types.Uint256) [255]BranchID {
	var branchIDs [255]BranchID
	index := leafIndex
	for h := uint8(0); h < uint8(255); h++ {
		if index[0]%2 == 0 {
			index[0] += uint8(1)
		} else {
			index[0] -= uint8(1)
		}
		branchID := BuildBranchID(h, index)
		branchIDs[h] = branchID
		index = index.Divide2()
	}
	return branchIDs
}

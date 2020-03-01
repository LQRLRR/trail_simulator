package models

// Branch records blockHash which this branch hash is updated and the updated branch hash value.
type Branch struct {
	ID  BranchID
	Log map[[32]byte][32]byte // key is blockHash which this branch hash is updated and value is the updated branch hash value.
}

// BranchID identifies a node in the TXO tree.
// The first byte of branchId represents the height at which this branch was stored,
// and the suffix represent the index in that height.
type BranchID [33]byte

// BuildBranchID builds branchId from branch height and index in height.
func BuildBranchID(height byte, index [32]byte) BranchID {
	var branchID BranchID
	branchID[0] = uint8(height)
	copy(branchID[1:], index[:])
	return branchID
}

// NewBranch provides new branch instance.
func NewBranch(id BranchID, blockHash [32]byte, branchHash [32]byte) *Branch {
	return &Branch{id, map[[32]byte][32]byte{blockHash: branchHash}}
}

// AddUpdate add branch update to Log.
func (b *Branch) AddUpdate(blockHash [32]byte, branchHash [32]byte) {
	b.Log[blockHash] = branchHash
}

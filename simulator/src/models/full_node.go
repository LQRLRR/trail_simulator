package models

// Branches is all of update history.
var Branches map[BranchID]*Branch

// Blocks is all of generated blocks
var Blocks map[[32]byte]*Block

func init() {
	Branches = map[BranchID]*Branch{}
	Blocks = map[[32]byte]*Block{}
}

func downloadLatestUpdates(from [32]byte, to [32]byte, branchIDs map[BranchID]bool) map[BranchID][32]byte {
	updateds := map[BranchID][32]byte{}

	for branchID := range branchIDs {
		branch, exists := Branches[branchID]
		if !exists {
			panic("getUpdateBranches: invalid branchId")
		}
		for blockHash := to; blockHash != from; blockHash = Blocks[blockHash].Parent {
			if _, exists := branch.Log[blockHash]; exists {
				if _, exists := updateds[branchID]; !exists {
					updateds[branchID] = blockHash
				}
			}
		}
	}
	return updateds
}

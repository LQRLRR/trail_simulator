package models

import (
	"crypto/sha256"
	"trail_simulator/simulator/src/setting"
	"trail_simulator/simulator/src/types"
)

var NullHash [255][32]byte

func init() {
	hash := sha256.Sum256([]byte(""))
	NullHash[0] = hash
	for h := 1; h < 255; h++ {
		s := append(hash[:], hash[:]...)
		hash = sha256.Sum256(s)
		NullHash[h] = hash
	}
}

// Node generate blocks.
type Node struct {
	ID     uint32
	Client *Client
}

// NewNode provide new node instance.
func NewNode(id uint32, client *Client) *Node {
	return &Node{ID: id, Client: client}
}

func (n *Node) validateTransactions(txs []*Transaction, parentHash [32]byte, parent Block) ([]*Proof, []*TXO, uint64) {
	var validProofs []*Proof
	var validOutputs []*TXO

	totalFee := uint64(0)
	for _, tx := range txs {
		if tx.BlockHash != parentHash {
			continue
		}
		totalInputBalance := uint64(0)

		isInvalid := false
		for _, proof := range tx.Inputs {
			if proof.Root(false) != parent.Root {
				isInvalid = true
				break
			}
			totalInputBalance += proof.TXO.Balance
		}

		if isInvalid {
			continue
		}

		totalOutputBalance := uint64(0)
		for _, txo := range tx.Outputs {
			totalOutputBalance += txo.Balance
		}

		if totalInputBalance < totalOutputBalance+uint64(len(tx.Inputs)*setting.FeePerTXO) {
			continue
		}
		validProofs = append(validProofs, tx.Inputs...)
		validOutputs = append(validOutputs, tx.Outputs...)
		totalFee += totalInputBalance - totalOutputBalance

		if len(validProofs) >= setting.InputsPerBlock {
			break
		}
	}
	return validProofs, validOutputs, totalFee
}

func (n *Node) fillTreeWithProofs(
	branches map[BranchID][32]byte,
	indexes [255]map[types.Uint256]bool,
	proofs []*Proof) (map[BranchID][32]byte, [255]map[types.Uint256]bool) {
	for _, proof := range proofs {
		index := proof.TXO.Index
		for h := uint8(0); h < uint8(255); h++ {
			if index[0]%2 == 0 {
				index[0] += uint8(1)
			} else {
				index[0] -= uint8(1)
			}
			hash := proof.Proofs[h]
			branchID := BuildBranchID(h, index)
			branches[branchID] = hash

			if index[0]%2 == 0 {
				indexes[h][index] = true
			}
			index = index.Divide2()
		}
	}
	return branches, indexes
}

func (n *Node) fillTreeWithParentBlock(
	branches map[BranchID][32]byte,
	indexes [255]map[types.Uint256]bool,
	parent *Block) (map[BranchID][32]byte, [255]map[types.Uint256]bool) {
	index := parent.RightmostIndex
	branchID := BuildBranchID(0, index)
	branches[branchID] = parent.RightmostHash
	if index[0]%2 == 0 {
		indexes[0][index] = true
	}

	for h := uint8(0); h < uint8(255); h++ {
		if index[0]%2 == 0 {
			index[0] += uint8(1)
		} else {
			index[0] -= uint8(1)
		}
		branchID = BuildBranchID(h, index)
		branches[branchID] = parent.RightmostProof[h]
		if index[0]%2 == 0 {
			indexes[h][index] = true
		}
		index = index.Divide2()
	}
	return branches, indexes
}

func (n *Node) fillTreeWithUsedTXOs(
	branches map[BranchID][32]byte,
	indexes [255]map[types.Uint256]bool,
	proofs []*Proof) (map[BranchID][32]byte, [255]map[types.Uint256]bool, []*TXO) {
	var usedTXOs []*TXO
	for _, proof := range proofs {
		index := proof.TXO.Index
		branchID := BuildBranchID(0, index)
		branches[branchID] = proof.TXO.Hash(true)
		if index[0]%2 == 0 {
			indexes[0][index] = true
		}
		usedTXOs = append(usedTXOs, proof.TXO)
	}
	return branches, indexes, usedTXOs
}

func (n *Node) fillTreeWithNewTXOs(
	branches map[BranchID][32]byte,
	indexes [255]map[types.Uint256]bool,
	outputs []*TXO,
	parentHash [32]byte,
	rightmostIndex types.Uint256) (map[BranchID][32]byte, [255]map[types.Uint256]bool, []*TXO, [32]byte) {
	var newTXOs []*TXO
	for _, txo := range outputs {
		rightmostIndex = rightmostIndex.AddUint8(1)
		txo.SetIndex(rightmostIndex)

		branchID := BuildBranchID(0, rightmostIndex)
		branches[branchID] = txo.Hash(false)
		if rightmostIndex[0]%2 == 0 {
			indexes[0][rightmostIndex] = true
		}
		newTXOs = append(newTXOs, txo)
	}
	return branches, indexes, newTXOs, rightmostIndex
}

func (n *Node) getParentBranchHash(height uint8, leftIndex types.Uint256, branches map[BranchID][32]byte) ([32]byte, map[BranchID][32]byte) {
	leftBranchID := BuildBranchID(height, leftIndex)
	rightIndex := leftIndex.AddUint8(1)
	rightBranchID := BuildBranchID(height, rightIndex)

	leftHash := branches[leftBranchID]
	rightHash, exists := branches[rightBranchID]
	if !exists {
		rightHash = NullHash[height]
		branches[rightBranchID] = rightHash
	}
	s := append(leftHash[:], rightHash[:]...)
	return sha256.Sum256(s), branches
}

func (n *Node) calcTreeRoot(
	branches map[BranchID][32]byte,
	indexes [255]map[types.Uint256]bool) ([32]byte, map[BranchID][32]byte) {
	var treeRoot [32]byte
	for h := uint8(0); h < uint8(255); h++ {
		heightindexes := indexes[h]
		for index := range heightindexes {
			parentBranchHash, branches := n.getParentBranchHash(h, index, branches)

			if h+1 == uint8(255) {
				treeRoot = parentBranchHash
			} else {
				parentIndex := index.Divide2()
				parentBranchID := BuildBranchID(h+1, parentIndex)
				branches[parentBranchID] = parentBranchHash
				if parentIndex[0]%2 == 0 {
					indexes[h+1][parentIndex] = true
				}
			}
		}
	}
	return treeRoot, branches
}

func (n *Node) getRightmostProof(branches map[BranchID][32]byte, rightmostIndex [32]byte) [255][32]byte {
	rightmostProof := [255][32]byte{}
	proofIDs := getProofBranchIDs(rightmostIndex)
	for h, proofID := range proofIDs {
		rightmostProof[h] = branches[proofID]
	}
	return rightmostProof
}

// BuildGenesis generate genesis block.
// BuildGenesis returns branch hashes, newTXOs, usedTXOs, genesisBlock.
func (n *Node) BuildGenesis(parentHash [32]byte, txos []*TXO) (map[BranchID][32]byte, []*TXO, []*TXO, *Block) {
	if len(txos) == 0 {
		panic("BuildGenesis: cant build genesis without txos")
	}

	branches := map[BranchID][32]byte{}
	var filledIndexes [255]map[types.Uint256]bool
	for h := uint8(0); h < uint8(255); h++ {
		filledIndexes[h] = map[types.Uint256]bool{}
	}

	branches, filledIndexes, newTXOs, rightmostIndex := n.fillTreeWithNewTXOs(branches, filledIndexes, txos, parentHash, types.Uint256{}.Max())

	treeRoot, branches := n.calcTreeRoot(branches, filledIndexes)

	rightmostHash := branches[BuildBranchID(0, rightmostIndex)]

	rightmostProof := n.getRightmostProof(branches, rightmostIndex)

	return branches, newTXOs, []*TXO{}, NewBlock(parentHash, 0, treeRoot, rightmostIndex, rightmostHash, rightmostProof)
}

// BuildBlock generate new block.
// BuildBlock returns branch hashes, newTXOs, usedTXOs, newBlock.
func (n *Node) BuildBlock(txs []*Transaction) (map[BranchID][32]byte, []*TXO, []*TXO, *Block) {
	if len(txs) == 0 {
		panic("BuildBlock: cant build block without transactions")
	}

	parent := Blocks[n.Client.HeadBlock]
	parentHash := n.Client.HeadBlock

	validProofs, validOutputs, totalFee := n.validateTransactions(txs, parentHash, *parent)
	rewardTXO := NewTXOWithoutIndex(parentHash, n.Client.Address, totalFee)
	validOutputs = append(validOutputs, rewardTXO)

	branches := map[BranchID][32]byte{}
	var filledIndexes [255]map[types.Uint256]bool
	for h := uint8(0); h < uint8(255); h++ {
		filledIndexes[h] = map[types.Uint256]bool{}
	}

	branches, filledIndexes = n.fillTreeWithProofs(branches, filledIndexes, validProofs)
	branches, filledIndexes = n.fillTreeWithParentBlock(branches, filledIndexes, parent)
	branches, filledIndexes, usedTXOs := n.fillTreeWithUsedTXOs(branches, filledIndexes, validProofs)
	branches, filledIndexes, newTXOs, rightmostIndex := n.fillTreeWithNewTXOs(branches, filledIndexes, validOutputs, parentHash, parent.RightmostIndex)

	treeRoot, branches := n.calcTreeRoot(branches, filledIndexes)

	rightmostHash := branches[BuildBranchID(0, rightmostIndex)]

	rightmostProof := n.getRightmostProof(branches, rightmostIndex)

	return branches, newTXOs, usedTXOs, NewBlock(parentHash, parent.Height+1, treeRoot, rightmostIndex, rightmostHash, rightmostProof)
}

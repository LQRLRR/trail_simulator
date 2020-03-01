package models

import (
	"errors"
	"trail_simulator/simulator/src/setting"
	"trail_simulator/simulator/src/types"
)

// Client is a account issues transactions.
type Client struct {
	Address   uint32
	HeadBlock [32]byte                            // hash value of the block client consider as head block.
	Blocks    map[[32]byte]bool                   // hash values of the blocks client recieved.
	TXOs      []*TXO                              // list of own TXOs.
	Unused    map[[32]byte]map[types.Uint256]*TXO // list of unused TXOs at head blocks in each forks.
	Used      map[[32]byte]map[types.Uint256]*TXO // list of used TXOs. the first keys are hash value of the block used TXO.
	Memory    map[BranchID]map[[32]byte]bool      // update history of Merkle proof on device.
	Archive   map[BranchID]map[[32]byte]bool      // update history of Merkle proof archived.
}

// NewClient provide new client.
func NewClient(address uint32) *Client {
	return &Client{Address: address,
		Blocks:  map[[32]byte]bool{},
		TXOs:    []*TXO{},
		Unused:  map[[32]byte]map[types.Uint256]*TXO{},
		Used:    map[[32]byte]map[types.Uint256]*TXO{},
		Memory:  map[BranchID]map[[32]byte]bool{},
		Archive: map[BranchID]map[[32]byte]bool{}}
}

// UnusedSize is number of unused TXOs.
func (c Client) UnusedSize() int {
	size := 0
	for _, b := range c.Unused {
		size += len(b)
	}
	return size
}

// UsedSize is number of used TXOs.
func (c Client) UsedSize() int {
	size := 0
	for _, b := range c.Used {
		size += len(b)
	}
	return size
}

// MemorySize is number of hash values of Merkle proof on device.
func (c Client) MemorySize() int {
	size := 0
	for _, b := range c.Memory {
		size += len(b)
	}
	return size
}

// ArchiveSize is number of hash values of Merkle proof archived.
func (c Client) ArchiveSize() int {
	size := 0
	for _, b := range c.Archive {
		size += len(b)
	}
	return size
}

// BuildProof returns proof of TXO at the block which client consider as head block.
func (c Client) BuildProof(txo *TXO) (*Proof, error) {
	proofs := [255][32]byte{}
	proofIDs := getProofBranchIDs(txo.Index)
	for h, proofID := range proofIDs {
		memory, exists := c.Memory[proofID]
		if !exists {
			return nil, errors.New("BuildProof: dont have this branch")
		}
		if len(memory) == 0 {
			return nil, errors.New("BuildProof: memory has no blockhash")
		}
		latestBlockHash := c.HeadBlock

		for _, exists = memory[latestBlockHash]; !exists; _, exists = memory[latestBlockHash] {
			block, exists := Blocks[latestBlockHash]
			if !exists {
				return nil, errors.New("BuildProof: invalid latest block hash in memory")
			}
			latestBlockHash = block.Parent
		}

		branch, exists := Branches[proofID]
		if !exists {
			return nil, errors.New("BuildProof: no branch data")
		}

		var branchHash [32]byte
		for branchHash, exists = branch.Log[latestBlockHash]; !exists; branchHash, exists = branch.Log[latestBlockHash] {
			block, exists := Blocks[latestBlockHash]
			latestBlockHash = block.Parent
			if !exists {
				return nil, errors.New("BuildProof: invalid latest block hash in branch")
			}
		}
		proofs[h] = branchHash
	}
	return NewProof(txo, proofs), nil
}

// Balance returns client's total balance.
func (c Client) Balance(blockHash [32]byte) uint64 {
	balance := uint64(0)
	for _, txo := range c.Unused[blockHash] {
		balance += txo.Balance
	}
	return balance
}

func (c *Client) downloadBranchUpdates(from [32]byte, newBlockHash [32]byte) {
	var branchIDs map[BranchID]bool
	for _, txo := range c.Unused[newBlockHash] {
		proofIDs := getProofBranchIDs(txo.Index)
		for _, proofID := range proofIDs {
			branchIDs[proofID] = true
		}
	}
	branchUpdates := downloadLatestUpdates(from, newBlockHash, branchIDs)
	for branchID, blockHash := range branchUpdates {
		if _, exists := c.Memory[branchID]; exists {
			c.Memory[branchID][blockHash] = true
		} else {
			c.Memory[branchID] = map[[32]byte]bool{blockHash: true}
		}
	}
}

func (c *Client) markAsUsed(usedTXOs []*TXO, usedBlockHash [32]byte) {
	for _, txo := range usedTXOs {
		if txo.OwnerAddress == c.Address {
			delete(c.Unused[usedBlockHash], txo.Index)
			if _, exists := c.Used[usedBlockHash]; exists {
				c.Used[usedBlockHash][txo.Index] = txo
			} else {
				c.Used[usedBlockHash] = map[types.Uint256]*TXO{txo.Index: txo}
			}
		}
	}
}

func (c *Client) addUnuseds(newTXOs []*TXO, newBlockHash [32]byte) {
	for _, txo := range newTXOs {
		if txo.OwnerAddress == c.Address {
			c.Unused[newBlockHash][txo.Index] = txo
			c.TXOs = append(c.TXOs, txo)
		}
	}
}

func (c *Client) addBranchUpdate(
	branchID BranchID,
	updateBlockHash [32]byte,
	branchIDs map[BranchID]bool) map[[32]byte]bool {
	blockHashs, exists := c.Memory[branchID]
	if exists {
		if _, exists := branchIDs[branchID]; exists {
			if _, exists := Branches[branchID].Log[updateBlockHash]; exists {
				blockHashs[updateBlockHash] = true
			}
		}
	} else if _, exists := branchIDs[branchID]; exists {
		blockHashs, exists = c.Archive[branchID]
		if exists {
			blockHashs[updateBlockHash] = true
		} else {
			blockHashs = map[[32]byte]bool{updateBlockHash: true}
		}
	} else {
		blockHashs = c.Archive[branchID]
	}
	return blockHashs
}

func (c *Client) archiveOldBranchUpdate(branchID BranchID, updateBlockHashes map[[32]byte]bool, threshold uint64) map[[32]byte]bool {
	filteredBlockHashes := updateBlockHashes
	for blockHash := range updateBlockHashes {
		if Blocks[blockHash].Height < threshold {
			delete(filteredBlockHashes, blockHash)
			if _, exists := c.Archive[branchID]; exists {
				c.Archive[branchID][blockHash] = true
			} else {
				c.Archive[branchID] = map[[32]byte]bool{blockHash: true}
			}
		}
	}
	return filteredBlockHashes
}

func (c *Client) archiveNotReferenceBranchUpdate(newMemory map[BranchID]map[[32]byte]bool) {
	for branchID, blockHashs := range c.Memory {
		if _, exists := newMemory[branchID]; !exists {
			if _, exists := c.Archive[branchID]; exists {
				for blockHash := range blockHashs {
					c.Archive[branchID][blockHash] = true
				}
			} else {
				c.Archive[branchID] = blockHashs
			}
		}
	}
}

func (c *Client) downloadBlocksUntilSameHeightBlockAsCurrentHeadBlock(newHeadBlockHash [32]byte) *Block {
	block := Blocks[newHeadBlockHash]
	for ; block.Height > Blocks[c.HeadBlock].Height; block = Blocks[block.Parent] {
		if _, exists := c.Blocks[block.Parent]; exists {
			panic("clinet update: client selected less height block")
		} else {
			c.Blocks[block.Parent] = true
		}
	}
	return block
}

func (c *Client) downloadParentBlocksNotHave(from *Block, unuseds map[types.Uint256]*TXO) (*Block, *Block, map[types.Uint256]*TXO) {
	clientBlock := Blocks[c.HeadBlock]
	block := from
	for block.Parent != clientBlock.Parent {
		if _, exists := c.Blocks[block.Parent]; exists {
			break
		}
		txos, exists := c.Used[clientBlock.Parent]
		if exists {
			for _, txo := range txos {
				unuseds[txo.Index] = txo
			}
		}
		block = Blocks[block.Parent]
		clientBlock = Blocks[clientBlock.Parent]
	}
	return block, clientBlock, unuseds
}

func (c *Client) updateUnuseds(newBlockHash [32]byte, blockHash [32]byte, clientBlockHash [32]byte, unuseds map[types.Uint256]*TXO) {
	if blockHash != clientBlockHash {
		c.Unused[newBlockHash] = c.Unused[blockHash]
		delete(c.Unused, blockHash)
	} else {
		c.Unused[newBlockHash] = map[types.Uint256]*TXO{}
		forkPoint := Blocks[blockHash]
		for _, txo := range unuseds {
			if !txo.Index.Larger(forkPoint.RightmostIndex) {
				c.Unused[newBlockHash][txo.Index] = txo
			}
		}
	}
}

// Update client's data.
func (c *Client) Update(branchIDs map[BranchID]bool, newTXOs []*TXO, usedTXOs []*TXO, newBlockHash [32]byte) {
	newBlock := Blocks[newBlockHash]
	if newBlock.Height != 0 && newBlock.Height <= Blocks[c.HeadBlock].Height {
		return
	}
	if _, exists := c.Blocks[newBlockHash]; exists {
		return
	}

	if newBlock.Height != 0 && newBlock.Parent != c.HeadBlock {
		block := c.downloadBlocksUntilSameHeightBlockAsCurrentHeadBlock(newBlockHash)
		unuseds := c.Unused[c.HeadBlock]

		// fork occurs
		if block.Hash() != c.HeadBlock {
			txos, exists := c.Used[c.HeadBlock]
			if exists {
				for _, txo := range txos {
					unuseds[txo.Index] = txo
				}
			}
			block, clientBlock, unuseds := c.downloadParentBlocksNotHave(block, unuseds)

			c.updateUnuseds(newBlockHash, block.Parent, clientBlock.Parent, unuseds)
			c.downloadBranchUpdates(block.Parent, newBlockHash)
		} else {
			// Client doesn't have blocks that are ancestors of newBLock, and are descendants of c.HeadBlock.
			c.Unused[newBlockHash] = c.Unused[c.HeadBlock]
			delete(c.Unused, c.HeadBlock)
			c.downloadBranchUpdates(c.HeadBlock, newBlockHash)
		}
	} else { // recieves genesis block or child block of c.HeadBlock
		if newBlock.Height == 0 {
			c.Unused[newBlockHash] = map[types.Uint256]*TXO{}
		} else {
			c.Unused[newBlockHash] = c.Unused[c.HeadBlock]
			delete(c.Unused, newBlock.Parent)
		}
	}
	c.HeadBlock = newBlockHash

	c.markAsUsed(usedTXOs, newBlockHash)

	c.addUnuseds(newTXOs, newBlockHash)

	newMemory := map[BranchID]map[[32]byte]bool{}
	for _, txo := range c.Unused[newBlockHash] {
		proofIDs := getProofBranchIDs(txo.Index)
		for _, proofID := range proofIDs {
			updateBlockHashes := c.addBranchUpdate(proofID, newBlockHash, branchIDs)
			if len(updateBlockHashes) > 1 && newBlock.Height > setting.ArchiveHeight {
				threshold := newBlock.Height - setting.ArchiveHeight
				updateBlockHashes = c.archiveOldBranchUpdate(proofID, updateBlockHashes, threshold)
			}
			newMemory[proofID] = updateBlockHashes
		}
	}

	c.archiveNotReferenceBranchUpdate(newMemory)

	c.Memory = newMemory
	c.Blocks[newBlockHash] = true
}

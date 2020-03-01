package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"time"
	"trail_simulator/simulator/src/helpers"
	"trail_simulator/simulator/src/models"
	"trail_simulator/simulator/src/setting"
	"trail_simulator/simulator/src/types"
)

var filePath string

func init() {
	filePath = "/go/src/trail_simulator/simulator/output/output_" + fmt.Sprint(time.Now().Unix()) + ".json"
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString("{")
	file.WriteString(
		"\"setting\":" +
			"{\"number_of_node\":" + fmt.Sprint(setting.NumberOfNode) +
			",\"number_of_client\":" + fmt.Sprint(setting.NumberOfClient) +
			",\"end_block_height\":" + fmt.Sprint(setting.EndBlockHeight) +
			",\"archive_height\":" + fmt.Sprint(setting.ArchiveHeight) +
			",\"inputs_per_block\":" + fmt.Sprint(setting.InputsPerBlock) +
			"},\"blocks\":[")
}

func main() {
	if setting.NumberOfNode == 0 {
		panic("NumberOfNode must be larger than 0")
	}
	if setting.NumberOfNode > setting.NumberOfClient {
		panic("NumberOfClient must be larger than NumberOfNode")
	}
	if setting.NumberOfClient < setting.InputsPerBlock {
		panic("NumberOfClient must be larger than InputsPerBlock")
	}
	if setting.TotalBalance/setting.NumberOfClient < setting.FeePerTXO {
		panic("initial balance must be larger than fee")
	}

	tb := helpers.CreateTimeBomb()
	timer := helpers.CreateTimer()
	timer.Start("simulation")

	var clients []*models.Client
	var nodes []*models.Node

	var genesisTXOs []*models.TXO
	addresses := types.List{}
	parentHash := models.NullHash[0]
	for id := 0; id < setting.NumberOfClient; id++ {
		addresses = append(addresses, id)
		clients = append(clients, models.NewClient(uint32(id)))
		genesisTXOs = append(genesisTXOs, models.NewTXOWithoutIndex(parentHash, uint32(id), setting.TotalBalance/setting.NumberOfClient))
	}

	for id := 0; id < setting.NumberOfNode; id++ {
		nodes = append(nodes, models.NewNode(uint32(id), clients[id]))
	}

	branches, newTXOs, usedTXOs, block := nodes[0].BuildGenesis(parentHash, genesisTXOs)

	blockHash := block.Hash()
	models.Blocks[blockHash] = block
	branchIDs := map[models.BranchID]bool{}
	for branchID, hash := range branches {
		models.Branches[branchID] = models.NewBranch(branchID, blockHash, hash)
		branchIDs[branchID] = true
	}

	for i := 0; i < setting.NumberOfClient; i++ {
		clients[i].Update(branchIDs, newTXOs, usedTXOs, blockHash)
	}

	for block.Height < setting.EndBlockHeight {
		tb.Start(5, "build tx")
		rand.Seed(time.Now().UnixNano())
		var txs []*models.Transaction
		tmpAddresses := addresses
		for i := 0; i < setting.InputsPerBlock; i += 2 {
			r := rand.Intn(len(tmpAddresses))
			index1 := tmpAddresses[r].(int)
			tmpAddresses = tmpAddresses.Remove(r)
			r = rand.Intn(len(tmpAddresses))
			index2 := tmpAddresses[r].(int)
			tmpAddresses = tmpAddresses.Remove(r)

			client1 := clients[index1]
			client2 := clients[index2]
			tx, err := models.BuildTransaction(client1, client2)
			if tx != nil && err == nil {
				txs = append(txs, tx)
			}

			if err != nil {
				panic(err)
			}
		}
		tb.Clear()

		tb.Start(5, "build block")
		nodeID := rand.Intn(setting.NumberOfNode)
		branches, newTXOs, usedTXOs, block = nodes[nodeID].BuildBlock(txs)
		tb.Clear()

		tb.Start(5, "update branches")
		blockHash := block.Hash()
		models.Blocks[blockHash] = block
		branchIDs = map[models.BranchID]bool{}
		for branchID, hash := range branches {
			if _, exists := models.Branches[branchID]; exists {
				if preHash, exists := models.Branches[branchID].Log[block.Parent]; !exists || preHash != hash {
					models.Branches[branchID].AddUpdate(blockHash, hash)
				}
			} else {
				models.Branches[branchID] = models.NewBranch(branchID, blockHash, hash)
			}
			branchIDs[branchID] = true
		}
		tb.Clear()

		tb.Start(10, "update client")
		for i := 0; i < setting.NumberOfClient; i++ {
			clients[i].Update(branchIDs, newTXOs, usedTXOs, blockHash)
		}
		tb.Clear()

		tb.Start(5, "validation")
		validation(clients, newTXOs)
		tb.Clear()

		outputBlockData(clients, *block, branchIDs, newTXOs, usedTXOs)
	}
	timer.RecordLap()
}

func outputBlockData(clients []*models.Client, block models.Block, branchIDs map[models.BranchID]bool, newTXOs []*models.TXO, usedTXOs []*models.TXO) {
	maxUnused := 0
	maxUsed := 0
	maxMemory := 0
	maxArchive := 0
	for _, client := range clients {
		unusedSize := client.UnusedSize()
		usedSize := client.UsedSize()
		memorySize := client.MemorySize()
		archiveSize := client.ArchiveSize()
		if unusedSize > maxUnused {
			maxUnused = unusedSize
		}
		if usedSize > maxUsed {
			maxUsed = usedSize
		}
		if memorySize > maxMemory {
			maxMemory = memorySize
		}
		if archiveSize > maxArchive {
			maxArchive = archiveSize
		}
	}
	blockHash := block.Hash()
	blockHashStr := hex.EncodeToString(blockHash[:8])
	fmt.Println(block.Height, " ", blockHashStr)
	fmt.Printf("unused %d, used %d, memory %d, storage %d\n", maxUnused, maxUsed, maxMemory, maxArchive)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	record := fmt.Sprintf(
		"{\"height\":%d,"+
			"\"block_hash\":%q,"+
			"\"number_of_updated_branchs\":%d,"+
			"\"number_of_new_utxo\":%d,"+
			"\"number_of_used_utxo\":%d,"+
			"\"max_unused\":%d,"+
			"\"max_used\":%d,"+
			"\"max_memory\":%d,"+
			"\"max_archiive\":%d}",
		block.Height,
		blockHashStr,
		len(branchIDs),
		len(newTXOs),
		len(usedTXOs),
		maxUnused,
		maxUsed,
		maxMemory,
		maxArchive)
	file.WriteString(record)
	if block.Height == setting.EndBlockHeight {
		file.WriteString("]}")
	} else {
		file.WriteString(",")
	}
}

func validation(clients []*models.Client, newTXOs []*models.TXO) {
	for _, txo := range newTXOs {
		for _, client := range clients {
			if client.Address == txo.OwnerAddress {
				_, err := client.BuildProof(txo)
				if err != nil {
					panic("validation:" + err.Error())
				}
			}
		}
	}
}

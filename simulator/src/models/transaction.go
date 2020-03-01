package models

import (
	"errors"
	"trail_simulator/simulator/src/setting"
)

// Transaction represents transfer of balance.
type Transaction struct {
	BlockHash [32]byte
	Inputs    []*Proof // input TXOs and its Merkle proof.
	Outputs   []*TXO   // ouput TXOs.
}

// BuildTransaction returns a transaction between two clients.
// In this implementation, the input is simply all proofs of the TXOs of the client,
// and the output is half the total balance of the input minus fees.
func BuildTransaction(a *Client, b *Client) (*Transaction, error) {
	if a.HeadBlock != b.HeadBlock {
		return nil, errors.New("BuildTransaction: clients not follow same block")
	}
	totalBalance := uint64(0)
	var inputs []*Proof
	for _, txo := range a.Unused[a.HeadBlock] {
		totalBalance += txo.Balance
		proof, err := a.BuildProof(txo)
		if err != nil {
			panic(err)
		}
		inputs = append(inputs, proof)
	}

	for _, txo := range b.Unused[b.HeadBlock] {
		totalBalance += txo.Balance
		proof, err := b.BuildProof(txo)
		if err != nil {
			panic(err)
		}
		inputs = append(inputs, proof)
	}

	if totalBalance < uint64(len(inputs)*setting.FeePerTXO) {
		return nil, errors.New("BuildTransaction: cant pay transaction fee")
	}

	outputBalance := totalBalance - uint64(len(inputs)*setting.FeePerTXO)
	output1 := NewTXOWithoutIndex(a.HeadBlock, a.Address, outputBalance/2)
	output2 := NewTXOWithoutIndex(b.HeadBlock, b.Address, outputBalance-outputBalance/2)

	return &Transaction{
		BlockHash: a.HeadBlock,
		Inputs:    inputs,
		Outputs:   []*TXO{output1, output2}}, nil
}

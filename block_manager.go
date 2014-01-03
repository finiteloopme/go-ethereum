
  // Blocks, blocks will have transactions.
  // Transactions/contracts are updated in goroutines
  // Each contract should send a message on a channel with usage statistics
  // The statics can be used for fee calculation within the block update method
  // Statistics{transaction, /* integers */ normal_ops, store_load, extro_balance, crypto, steps}
  // The block updater will wait for all goroutines to be finished and update the block accordingly
  // in one go and should use minimal IO overhead.
  // The actual block updating will happen within a goroutine as well so normal operation may continue

package main

import (
  "fmt"
)

type BlockManager struct {
  vm *Vm
}

func NewBlockManager() *BlockManager {
  bm := &BlockManager{vm: NewVm()}

  return bm
}

// Process a block.
func (bm *BlockManager) ProcessBlock(block *Block) error {
  // Get the tx count. Used to create enough channels to 'join' the go routines
  txCount  := len(block.transactions)
  // Locking channel. When it has been fully buffered this method will return
  lockChan := make(chan bool, txCount)

  // Process each transaction/contract
  for _, tx := range block.transactions {
    // If there's no recipient, it's a contract
    if tx.recipient == "" {
      go bm.ProcessContract(tx, block, lockChan)
    } else {
      // "finish" tx which isn't a contract
      lockChan <- true
    }
  }

  // Wait for all Tx to finish processing
  for i := 0; i < txCount; i++ {
    <- lockChan
  }

  return nil
}

func (bm *BlockManager) ProcessContract(tx *Transaction, block *Block, lockChan chan bool) {
  // Recovering function in case the VM had any errors
  defer func() {
    if r := recover(); r != nil {
      fmt.Println("Recovered from VM execution with err =", r)
      // Let the channel know where done even though it failed (so the execution may resume normally)
      lockChan <- true
    }
  }()

  // Process contract
  bm.vm.ProcContract(tx, block, func(opType OpType) bool {
    // TODO turn on once big ints are in place
    //if !block.PayFee(tx.Hash(), StepFee.Uint64()) {
    //  return false
    //}

    return true // Continue
  })

  // Broadcast we're done
  lockChan <- true
}
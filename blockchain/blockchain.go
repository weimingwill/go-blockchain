package blockchain

// Blockchain keeps sequence of blocks
type Blockchain struct {
	Blocks []*Block
}

// AddBlock adds a new block to blockchain
func (b *Blockchain) AddBlock(data string) {
	prevBlock := b.Blocks[len(b.Blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	b.Blocks = append(b.Blocks, newBlock)
}

// NewBlockchain creates and returns a blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}

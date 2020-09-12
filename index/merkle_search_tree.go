package index

type MerkleSearchNode struct {
	level    uint
	hash     string
	children []string
}

type MerkleSearchTree struct {
}

func NewMerkleSearchTree() *MerkleSearchTree {
	return new(MerkleSearchTree)
}

func (t *MerkleSearchTree) Put(k Key, v Value) {
}

func (t *MerkleSearchTree) Get(k Key) Value {
	return nil
}

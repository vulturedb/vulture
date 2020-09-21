package mst

import (
	"fmt"
	"sort"
)

type Child struct {
	key   Key
	value Value
	high  []byte
}

func (c Child) Key() Key {
	return c.key
}

func (c Child) Value() Value {
	return c.value
}

func (c Child) High() []byte {
	return c.high
}

// Represents a node in a Merkle Search Tree
// It should always be true that len(children) > 0, i.e. empty nodes should not exist.
type Node struct {
	level    uint32
	low      []byte
	children []Child
}

func (n *Node) Level() uint32 {
	return n.level
}

func (n *Node) Low() []byte {
	return n.low
}

func (n *Node) Children() []Child {
	children := make([]Child, len(n.children))
	copy(children, n.children)
	return children
}

func (n *Node) find(key Key) uint {
	i := sort.Search(len(n.children), func(i int) bool {
		return key.Less(n.children[i].key)
	})
	if i < 0 {
		panic(fmt.Sprintf("i cannot be < 0, got %d", i))
	}
	return uint(i)
}

func (n *Node) childAt(i uint) []byte {
	if i == 0 {
		return n.low
	} else {
		return n.children[i-1].high
	}
}

func (n *Node) findChild(key Key) ([]byte, uint) {
	i := n.find(key)
	if i > 0 && keysEqual(key, n.children[i-1].key) {
		panic(fmt.Errorf("Trying to get childHash but key matches. Key: %v, Level: %d", key, n.level))
	}
	return n.childAt(i), i
}

func (n *Node) withHashAt(hash []byte, at uint) *Node {
	if at == 0 {
		return &Node{
			level:    n.level,
			low:      hash,
			children: n.children,
		}
	} else {
		newChildren := make([]Child, len(n.children))
		copy(newChildren, n.children)
		newChildren[at-1].high = hash
		return &Node{
			level:    n.level,
			low:      n.low,
			children: newChildren,
		}
	}
}

func (n *Node) withMergedValueAt(val Value, at uint) *Node {
	newChildren := make([]Child, len(n.children))
	copy(newChildren, n.children)
	newChildren[at].value = n.children[at].value.Merge(val)
	return &Node{
		level:    n.level,
		low:      n.low,
		children: newChildren,
	}
}

func (n *Node) withChildInsertedAt(
	key Key,
	val Value,
	node []byte,
	at uint,
) *Node {
	newChildren := make([]Child, len(n.children)+1)
	copy(newChildren[:at], n.children[:at])
	copy(newChildren[at+1:], n.children[at:])
	newChildren[at] = Child{key, val, node}
	return &Node{
		level:    n.level,
		low:      n.low,
		children: newChildren,
	}
}

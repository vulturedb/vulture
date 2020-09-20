package mst

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type Child struct {
	key   Key
	value Value
	node  []byte
}

type Node struct {
	level    uint32
	low      []byte
	children []Child
}

func (n *Node) PutBytes(w io.Writer) error {
	// This can probably be improved...
	// I still don't even know if there are any cases where this easily breaks
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, n.level)
	_, err := w.Write(buf)
	if err != nil {
		return err
	}
	// Write key/values first
	if n.children != nil {
		for _, child := range n.children {
			err = child.key.PutBytes(w)
			if err != nil {
				return err
			}
			err = child.value.PutBytes(w)
			if err != nil {
				return err
			}
		}
	}
	// Links second
	if n.low != nil {
		_, err = w.Write(n.low)
		if err != nil {
			return err
		}
	}
	if n.children != nil {
		for _, child := range n.children {
			if child.node != nil {
				_, err = w.Write(child.node)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
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
		return n.children[i-1].node
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
		newChildren[at-1].node = hash
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

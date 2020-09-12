package index

import "crypto"

func hash(obj Hashable, h crypto.Hash) []byte {
	return h.New().Sum(obj.Bytes())
}

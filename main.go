package main

import (
	"bufio"
	"context"
	"crypto"
	_ "crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	mh "github.com/multiformats/go-multihash"

	"github.com/vulturedb/vulture/mst"
	"github.com/vulturedb/vulture/server"
)

const usage string = `
Available commands:
put <key> <value>
get <key>
root-hash
`

func printReplUsage() {
	fmt.Printf(strings.TrimLeft(usage, "\n"))
}

type UInt32KeyReader struct{}

func (kr UInt32KeyReader) FromBytes(b []byte) (mst.Key, error) {
	return mst.UInt32(binary.LittleEndian.Uint32(b)), nil
}

type UInt32ValueReader struct{}

func (kr UInt32ValueReader) FromBytes(b []byte) (mst.Value, error) {
	return mst.UInt32(binary.LittleEndian.Uint32(b)), nil
}

func repl() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs, err := server.SpawnDefault(ctx)
	if err != nil {
		panic(err)
	}

	store := server.NewIPFSMSTNodeStore(
		ctx,
		ipfs.Dag(),
		mh.SHA2_256,
		UInt32KeyReader{},
		UInt32ValueReader{},
	)
	ind := mst.NewMST(mst.Base2, crypto.SHA256, store)
	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Goodbye!")
				return
			}
			panic(err)
		}
		text = strings.TrimRight(text, "\n")
		tokens := strings.Split(text, " ")
		if len(tokens) == 0 {
			printReplUsage()
			continue
		}
		switch tokens[0] {
		case "put":
			if len(tokens) != 3 {
				printReplUsage()
				continue
			}
			rawKey, err := strconv.ParseUint(tokens[1], 10, 32)
			if err != nil {
				printReplUsage()
				continue
			}
			rawVal, err := strconv.ParseUint(tokens[2], 10, 32)
			if err != nil {
				printReplUsage()
				continue
			}
			ind.Put(mst.UInt32(rawKey), mst.UInt32(rawVal))
		case "get":
			if len(tokens) != 2 {
				printReplUsage()
				continue
			}
			rawKey, err := strconv.ParseUint(tokens[1], 10, 32)
			if err != nil {
				printReplUsage()
				continue
			}
			fmt.Printf("%d\n", ind.Get(mst.UInt32(rawKey)))
		case "root-hash":
			if len(tokens) != 1 {
				printReplUsage()
				continue
			}
			fmt.Println(hex.EncodeToString(ind.RootHash()))
		case "print":
			if len(tokens) != 1 {
				printReplUsage()
				continue
			}
			ind.PrintInOrder()
		default:
			printReplUsage()
		}
	}
}

func main() {
	repl()
}

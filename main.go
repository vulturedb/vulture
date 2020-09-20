package main

import (
	"bufio"
	"crypto"
	_ "crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/vulturedb/vulture/mst"
)

type UInt32 uint32

func (f UInt32) Less(than mst.Key) bool {
	return f < than.(UInt32)
}

func (f UInt32) PutBytes(w io.Writer) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(f))
	_, err := w.Write(buf)
	return err
}

func (f UInt32) Merge(with mst.Value) mst.Value {
	if f > with.(UInt32) {
		return f
	} else {
		return with
	}
}

const usage string = `
Available commands:
put <key> <value>
get <key>
root-hash
`

func printReplUsage() {
	fmt.Printf(strings.TrimLeft(usage, "\n"))
}

func repl() {
	ind := mst.NewLocalMST(mst.Base2, crypto.SHA256)
	reader := bufio.NewReader(os.Stdin)
	for true {
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
			ind.Put(UInt32(rawKey), UInt32(rawVal))
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
			fmt.Printf("%d\n", ind.Get(UInt32(rawKey)))
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

package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/vulturedb/vulture/service/rpc"
	"google.golang.org/grpc"
)

const usage string = `
Available commands:
put <key> <value>
get <key>
loop
`

func printReplUsage() {
	fmt.Printf(strings.TrimLeft(usage, "\n"))
}

var host = flag.String("host", "localhost", "host to connect to")
var port = flag.Int("port", 6667, "port to connect to")

func main() {
	flag.Parse()

	// Connect to a node
	address := fmt.Sprintf("%s:%d", *host, *port)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	fmt.Printf("Connected to %s\n", address)
	client := rpc.NewMSTServiceClient(conn)
	streamClient := rpc.NewMSTManagerServiceClient(conn)

	// Repl loop
	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Printf("> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("bye bye")
				return
			}
			log.Fatalf("Unexpected error from read string: %v", err)
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

			_, err = client.Put(context.Background(), &rpc.MSTPutRequest{
				Key:   uint32(rawKey),
				Value: uint32(rawVal),
			})
			if err != nil {
				log.Fatalf("Error when putting: %v", err)
			}
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
			resp, err := client.Get(context.Background(), &rpc.MSTGetRequest{Key: uint32(rawKey)})
			if err != nil {
				log.Fatalf("Error when getting: %v", err)
			}
			fmt.Printf("%d\n", resp.GetValue())
		case "loop":
			stream, err := streamClient.Manage(context.Background())
			if err != nil {
				log.Fatalf("Error when starting stream: %v", err)
			}
			for i := 0; i < 5; i++ {
				err = stream.Send(&rpc.MSTManageCommand{Val: fmt.Sprintf("Hi %d", i)})
				if err != nil {
					log.Fatalf("Error when sending in stream: %v", err)
				}
				out, err := stream.Recv()
				if err != nil {
					log.Fatalf("Error when receiving from stream: %v", err)
				}
				fmt.Printf("%s\n", out.GetVal())
			}
		default:
			printReplUsage()
		}
	}
}

package main

import (
	"context"
	"fmt"
	"path/filepath"

	cid "github.com/ipfs/go-cid"
	config "github.com/ipfs/go-ipfs-config"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	cbor "github.com/ipfs/go-ipld-cbor"
	icore "github.com/ipfs/interface-go-ipfs-core"
	path "github.com/ipfs/interface-go-ipfs-core/path"
	mh "github.com/multiformats/go-multihash"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"

	// This package is needed so that all the preloaded plugins are loaded automatically
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

/// ------ Setting up the IPFS Repo

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

/// ------ Spawning the node

// Creates an IPFS node and returns its coreAPI
func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, error) {
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	// Construct the node

	nodeOptions := &core.BuildCfg{
		Online: true,
		// This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		Routing: libp2p.DHTOption,
		// This option sets the node to be a client DHT node (only fetching records)
		// Routing: libp2p.DHTClientOption,
		Repo: repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

// Spawns a node on the default repo location, if the repo exists
func spawnDefault(ctx context.Context) (icore.CoreAPI, error) {
	defaultPath, err := config.PathRoot()
	if err != nil {
		// shouldn't be possible
		return nil, err
	}

	if err := setupPlugins(defaultPath); err != nil {
		return nil, err

	}

	return createNode(ctx, defaultPath)
}

type MyStruct struct {
	Items map[string]MyStruct
	Foo   string
	Bar   []byte
	Baz   []int
}

func testStruct() MyStruct {
	return MyStruct{
		Items: map[string]MyStruct{
			"Foo": {
				Foo: "Foo",
				Bar: []byte("Bar"),
				Baz: []int{1, 2, 3, 4},
			},
			"Bar": {
				Bar: []byte("Bar"),
				Baz: []int{1, 2, 3, 4},
			},
		},
		Baz: []int{5, 1, 2},
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ipfs, err := spawnDefault(ctx)
	if err != nil {
		fmt.Println("No IPFS repo available on the default path")
	}
	fmt.Println("IPFS node is running")

	exampleCIDStr := "QmUaoioqU7bxezBQZkUcgcSyokatMY71sxsALxQmRRrHrj"

	fmt.Printf("Fetching a file from the network with CID %s\n", exampleCIDStr)
	testCIDPath := path.New(exampleCIDStr)

	cont, err := ipfs.Unixfs().Get(ctx, testCIDPath)
	if err != nil {
		panic(fmt.Errorf("Could not get file with CID: %s", err))
	}
	s, err := cont.Size()
	if err != nil {
		panic(fmt.Errorf("Could not get size: %s", err))
	}
	fmt.Printf("Size of file: %d\n", s)

	testCID, err := cid.Decode(exampleCIDStr)
	if err != nil {
		panic(fmt.Errorf("Could not decode cid(%s): %s", testCID, err))
	}

	cbor.RegisterCborType(MyStruct{})
	nd, err := cbor.WrapObject(testStruct(), mh.SHA2_256, -1)
	if err != nil {
		panic(fmt.Errorf("Could not wrap object: %s", err))
	}

	dagService := ipfs.Dag()
	err = dagService.Add(ctx, nd)
	if err != nil {
		panic(fmt.Errorf("Could not add node to dag: %s", err))
	}

	fmt.Printf("Added node %s\n", nd.Cid().String())

	nd2, err := dagService.Get(ctx, nd.Cid())
	if err != nil {
		panic(fmt.Errorf("Could not get node from dag: %s", err))
	}

	fmt.Printf("Got node %s:\n", nd2.Cid().String())
	for _, ln := range nd2.Tree("", -1) {
		fmt.Println(ln)
	}
}

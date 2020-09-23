package ipfs

import (
	"context"

	"github.com/mitchellh/mapstructure"

	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"

	"github.com/vulturedb/vulture/core"
)

func RegisterTypes() {
	cbor.RegisterCborType(core.FieldSpec{})
	cbor.RegisterCborType(core.Schema{})

	cbor.RegisterCborType(iPFSMSTChild{})
	cbor.RegisterCborType(iPFSMSTNode{})

	cbor.RegisterCborType(format.Link{})
}

func unmarshal(v interface{}, m interface{}) error {
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   v,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	decoder.Decode(m)
	return nil
}

func PutSchema(c context.Context, a ipld.NodeAdder, s core.Schema) (cid.Cid, error) {
	nd, err := cbor.WrapObject(s, mh.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	err = a.Add(c, nd)
	if err != nil {
		return cid.Undef, err
	}
	return nd.Cid(), nil
}

func GetSchema(c context.Context, a ipld.NodeGetter, cid cid.Cid) (core.Schema, error) {
	nd, err := a.Get(c, cid)
	if err != nil {
		return core.GenesisSchema(), err
	}
	raw, _, err := nd.Resolve([]string{})
	if err != nil {
		return core.GenesisSchema(), err
	}
	s := &core.Schema{}
	if err = unmarshal(s, raw); err != nil {
		return core.GenesisSchema(), err
	}
	return *s, nil
}

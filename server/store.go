package server

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/mitchellh/mapstructure"
	mh "github.com/multiformats/go-multihash"

	"github.com/wojtechnology/cado/core"
)

func RegisterTypes() {
	cbor.RegisterCborType(core.FieldSpec{})
	cbor.RegisterCborType(core.Schema{})
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
		panic(fmt.Errorf("Couldn't unmarshal: %s", err))
	}
	return *s, nil
}

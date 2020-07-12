package util

import (
	"bytes"
	"context"
	"fmt"
	"github.com/filecoin-project/lotus/api"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// require for AMT and HAMT access
// TODO extract this to a common location in lotus and reuse the code
type ApiIpldStore struct {
	ctx context.Context
	api api.FullNode
}

func NewApiIpldStore(ctx context.Context, api api.FullNode) *ApiIpldStore {
	return &ApiIpldStore{
		ctx: ctx,
		api: api,
	}
}

func (ht *ApiIpldStore) Context() context.Context {
	return ht.ctx
}

func (ht *ApiIpldStore) Get(ctx context.Context, c cid.Cid, out interface{}) error {
	raw, err := ht.api.ChainReadObj(ctx, c)
	if err != nil {
		return err
	}

	cu, ok := out.(cbg.CBORUnmarshaler)
	if ok {
		if err := cu.UnmarshalCBOR(bytes.NewReader(raw)); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("Object does not implement CBORUnmarshaler: %T", out)
}

func (ht *ApiIpldStore) Put(ctx context.Context, v interface{}) (cid.Cid, error) {
	return cid.Undef, fmt.Errorf("Put is not implemented on ApiIpldStore")
}

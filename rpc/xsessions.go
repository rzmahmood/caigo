package rpc

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/dontpanicdao/caigo"
	"github.com/dontpanicdao/caigo/rpc/types"
)

type Policy struct {
	ContractAddress string `json:"contractAddress"`
	Selector        string `json:"selector"`
}

type XSession struct {
	Key      string   `json:"key"`
	Expires  uint64   `json:"expires"`
	Policies []Policy `json:"policies"`
}

var _ AccountPlugin = &XSessionsPlugin{}

// H(Policy(contractAddress:felt,selector:selector))
var POLICY_TYPE_HASH, _ = big.NewInt(0).SetString("0x2f0026e78543f036f33e26a8f5891b88c58dc1e20cbbfaf0bb53274da6fa568", 0)

type XSessionsPlugin struct {
	accountAddress types.Hash
	classHash      *big.Int
	mt             caigo.FixedSizeMerkleTree
	private        *big.Int
	xsession       XSession
}

func WithXSessionsPlugin(pluginClassHash string, xsession XSession) AccountOptionFunc {
	return func(private, address string) (AccountOption, error) {
		plugin, ok := big.NewInt(0).SetString(pluginClassHash, 0)
		if !ok {
			return AccountOption{}, errors.New("could not convert plugin class hash")
		}
		leaves := []*big.Int{}
		for _, policy := range xsession.Policies {
			contract, ok := big.NewInt(0).SetString(policy.ContractAddress, 0)
			if !ok {
				return AccountOption{}, errors.New("could not convert contract address")
			}
			leaf, _ := caigo.Curve.HashElements(append(
				[]*big.Int{},
				POLICY_TYPE_HASH,
				contract,
				caigo.GetSelectorFromName(policy.Selector),
			))
			leaves = append(leaves, leaf)
		}
		mt, err := caigo.NewFixedSizeMerkleTree(leaves...)
		if err != nil {
			return AccountOption{}, fmt.Errorf("could not create merkle tree, error: %v", err)
		}
		pk, ok := big.NewInt(0).SetString(private, 0)
		if !ok {
			return AccountOption{}, fmt.Errorf("could not get private key")
		}
		return AccountOption{
			AccountPlugin: &XSessionsPlugin{
				accountAddress: types.HexToHash(address),
				classHash:      plugin,
				mt:             *mt,
				private:        pk,
				xsession:       xsession,
			},
		}, nil
	}
}

func (xsessions *XSessionsPlugin) PluginCall(calls []types.FunctionCall) (types.FunctionCall, error) {
	data := []string{
		fmt.Sprintf("0x%s", xsessions.classHash.Text(16)),
		fmt.Sprintf("0x%s", big.NewInt(int64(xsessions.xsession.Expires))),
		fmt.Sprintf("0x%s", xsessions.mt.Root.Text(16)),
	}
	proofs := []*big.Int{}
	proofSize := int64(0)
	for _, call := range calls {
		leaf, _ := caigo.Curve.HashElements(append(
			[]*big.Int{},
			POLICY_TYPE_HASH,
			call.ContractAddress.Big(),
			caigo.GetSelectorFromName(call.EntryPointSelector),
		))
		p, err := xsessions.mt.GetProof(leaf, 0, []*big.Int{})
		if err != nil {
			return types.FunctionCall{}, err
		}
		if proofSize == 0 {
			proofSize = int64(len(p))
		}
		if proofSize != int64(len(p)) {
			return types.FunctionCall{}, errors.New("proof does not match proofsize")
		}
		proofs = append(proofs, p...)
	}
	data = append(data, fmt.Sprintf("0x%s", big.NewInt(proofSize).Text(16)))
	for _, proof := range proofs {
		data = append(data, fmt.Sprintf("0x%s", proof.Text(16)))
	}
	return types.FunctionCall{
		ContractAddress:    xsessions.accountAddress,
		EntryPointSelector: "use_plugin",
		CallData:           data,
	}, nil
}

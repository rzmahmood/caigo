package rpc

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/dontpanicdao/caigo/rpc/types"
)

// TestCall tests Call
func TestCall(t *testing.T) {
	testConfig := beforeEach(t)

	type testSetType struct {
		FunctionCall          types.FunctionCall
		BlockID               types.BlockID
		ExpectedPatternResult string
	}
	testSet := map[string][]testSetType{
		"devnet": {
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash("0x035a55a64238b776664d7723de1f6b50350116a1ab1ca1fe154320a0eba53d3a"),
					EntryPointSelector: "get_count",
					CallData:           []string{},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x01$",
			},
			{
				FunctionCall: types.FunctionCall{
					// ContractAddress of devnet ETH
					ContractAddress:    types.HexToHash("0x62230ea046a9a5fbc261ac77d03c8d41e5d442db2284587570ab46455fd2488"),
					EntryPointSelector: "balanceOf",
					CallData:           []string{DevNetAccount032Address},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x[0-9a-f]+$",
			},
		},
		"mock": {
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash("0xdeadbeef"),
					EntryPointSelector: "decimals",
					CallData:           []string{},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x12$",
			},
		},
		"testnet": {
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash("0x029260ce936efafa6d0042bc59757a653e3f992b97960c1c4f8ccd63b7a90136"),
					EntryPointSelector: "decimals",
					CallData:           []string{},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x12$",
			},
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash(TestNetETHAddress),
					EntryPointSelector: "balanceOf",
					CallData:           []string{"0x0207aCC15dc241e7d167E67e30E769719A727d3E0fa47f9E187707289885Dfde"},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x[0-9a-f]+$",
			},
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash(TestNetAccount032Address),
					EntryPointSelector: "get_nonce",
					CallData:           []string{},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x[0-9a-f]+$",
			},
		},
		"mainnet": {
			{
				FunctionCall: types.FunctionCall{
					ContractAddress:    types.HexToHash("0x06a09ccb1caaecf3d9683efe335a667b2169a409d19c589ba1eb771cd210af75"),
					EntryPointSelector: "decimals",
					CallData:           []string{},
				},
				BlockID:               WithBlockTag("latest"),
				ExpectedPatternResult: "^0x12$",
			},
		},
	}[testEnv]

	for _, test := range testSet {
		function := test.FunctionCall
		spy := NewSpy(testConfig.provider.c)
		testConfig.provider.c = spy
		output, err := testConfig.provider.Call(context.Background(), function, test.BlockID)
		if err != nil {
			t.Fatal(err)
		}
		if diff, err := spy.Compare(output, false); err != nil || diff != "FullMatch" {
			spy.Compare(output, true)
			t.Fatal("expecting to match", err)
		}
		if len(output) == 0 {
			t.Fatal("should return an output")
		}
		match, err := regexp.Match(test.ExpectedPatternResult, []byte(output[0]))
		if err != nil || !match {
			t.Fatalf("checking output(%v) expecting %s, got: %v", err, test.ExpectedPatternResult, output[0])
		}
		fmt.Println("output[0]", output[0])
	}
}

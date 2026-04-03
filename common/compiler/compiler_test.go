// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package compiler

import (
	"encoding/json"
	"testing"
)

// sampleSolcOutput is representative of the output produced by solc < 0.8.0
// with --combined-output --abi --bin --bin-runtime --userdoc --devdoc --hashes.
const sampleSolcOutput = `{
  "contracts": {
    "foo.sol:MyContract": {
      "abi": "[{\"constant\":false,\"inputs\":[],\"name\":\"hello\",\"outputs\":[],\"type\":\"function\"}]",
      "bin": "6001",
      "bin-runtime": "6002",
      "srcmap": "0:10:0:-",
      "srcmap-runtime": "0:10:0:-",
      "userdoc": "{\"methods\":{}}",
      "devdoc": "{\"methods\":{}}",
      "metadata": "{\"compiler\":{\"version\":\"0.7.6\"}}",
      "hashes": {"hello()": "19ff1d21"}
    }
  },
  "version": "0.7.6+commit.7338295f"
}`

// sampleSolcOutputV8 is representative of the output produced by solc >= 0.8.0,
// where ABI, devdoc and userdoc are JSON objects rather than strings.
const sampleSolcOutputV8 = `{
  "contracts": {
    "bar.sol:MyContractV8": {
      "bin": "6003",
      "bin-runtime": "6004",
      "srcmap": "0:5:0:-:0",
      "srcmap-runtime": "0:5:0:-:0",
      "abi": [{"inputs":[],"name":"greet","outputs":[],"stateMutability":"nonpayable","type":"function"}],
      "devdoc": {"kind":"dev","methods":{},"version":1},
      "userdoc": {"kind":"user","methods":{},"version":1},
      "metadata": "{\"compiler\":{\"version\":\"0.8.0\"}}",
      "hashes": {"greet()": "cfae3217"}
    }
  },
  "version": "0.8.0+commit.c7dfd78e"
}`

func TestParseCombinedJSON(t *testing.T) {
	contracts, err := ParseCombinedJSON(
		[]byte(sampleSolcOutput),
		"pragma solidity ^0.7.0;",
		"0.7.6",
		"0.7.6+commit.7338295f",
		"--optimize",
	)
	if err != nil {
		t.Fatalf("ParseCombinedJSON error: %v", err)
	}
	if len(contracts) != 1 {
		t.Fatalf("got %d contracts, want 1", len(contracts))
	}

	c, ok := contracts["foo.sol:MyContract"]
	if !ok {
		t.Fatal("contract 'foo.sol:MyContract' not found in output")
	}

	if c.Code != "0x6001" {
		t.Errorf("Code = %q, want %q", c.Code, "0x6001")
	}
	if c.RuntimeCode != "0x6002" {
		t.Errorf("RuntimeCode = %q, want %q", c.RuntimeCode, "0x6002")
	}
	if c.Info.Source != "pragma solidity ^0.7.0;" {
		t.Errorf("Info.Source = %q, want %q", c.Info.Source, "pragma solidity ^0.7.0;")
	}
	if c.Info.Language != "Solidity" {
		t.Errorf("Info.Language = %q, want %q", c.Info.Language, "Solidity")
	}
	if c.Info.LanguageVersion != "0.7.6" {
		t.Errorf("Info.LanguageVersion = %q, want %q", c.Info.LanguageVersion, "0.7.6")
	}
	if c.Info.CompilerVersion != "0.7.6+commit.7338295f" {
		t.Errorf("Info.CompilerVersion = %q, want %q", c.Info.CompilerVersion, "0.7.6+commit.7338295f")
	}
	if c.Info.CompilerOptions != "--optimize" {
		t.Errorf("Info.CompilerOptions = %q, want %q", c.Info.CompilerOptions, "--optimize")
	}
	if c.Info.SrcMap != "0:10:0:-" {
		t.Errorf("Info.SrcMap = %q, want %q", c.Info.SrcMap, "0:10:0:-")
	}
	if c.Info.SrcMapRuntime != "0:10:0:-" {
		t.Errorf("Info.SrcMapRuntime = %q, want %q", c.Info.SrcMapRuntime, "0:10:0:-")
	}
	if c.Hashes["hello()"] != "19ff1d21" {
		t.Errorf("Hashes[hello()] = %q, want %q", c.Hashes["hello()"], "19ff1d21")
	}
	if c.Info.AbiDefinition == nil {
		t.Error("AbiDefinition is nil")
	}
	if c.Info.UserDoc == nil {
		t.Error("UserDoc is nil")
	}
	if c.Info.DeveloperDoc == nil {
		t.Error("DeveloperDoc is nil")
	}
}

func TestParseCombinedJSONV8(t *testing.T) {
	contracts, err := ParseCombinedJSON(
		[]byte(sampleSolcOutputV8),
		"pragma solidity ^0.8.0;",
		"0.8.0",
		"0.8.0+commit.c7dfd78e",
		"",
	)
	if err != nil {
		t.Fatalf("ParseCombinedJSON (v8) error: %v", err)
	}
	if len(contracts) != 1 {
		t.Fatalf("got %d contracts, want 1", len(contracts))
	}

	c, ok := contracts["bar.sol:MyContractV8"]
	if !ok {
		t.Fatal("contract 'bar.sol:MyContractV8' not found in output")
	}

	if c.Code != "0x6003" {
		t.Errorf("Code = %q, want %q", c.Code, "0x6003")
	}
	if c.RuntimeCode != "0x6004" {
		t.Errorf("RuntimeCode = %q, want %q", c.RuntimeCode, "0x6004")
	}
	if c.Info.Language != "Solidity" {
		t.Errorf("Info.Language = %q, want %q", c.Info.Language, "Solidity")
	}
	// ABI should be a non-nil slice (parsed as interface{})
	if c.Info.AbiDefinition == nil {
		t.Error("AbiDefinition is nil for v8 output")
	}
	// Verify the ABI is a slice (array)
	if _, ok := c.Info.AbiDefinition.([]interface{}); !ok {
		t.Errorf("AbiDefinition type = %T, want []interface{}", c.Info.AbiDefinition)
	}
	if c.Hashes["greet()"] != "cfae3217" {
		t.Errorf("Hashes[greet()] = %q, want %q", c.Hashes["greet()"], "cfae3217")
	}
}

func TestParseCombinedJSONInvalid(t *testing.T) {
	_, err := ParseCombinedJSON([]byte("not valid json"), "", "", "", "")
	if err == nil {
		t.Error("ParseCombinedJSON with invalid JSON: want error, got nil")
	}
}

func TestParseCombinedJSONEmpty(t *testing.T) {
	// Valid JSON but with no contracts
	input := `{"contracts": {}, "version": "0.7.6"}`
	contracts, err := ParseCombinedJSON([]byte(input), "", "", "", "")
	if err != nil {
		t.Fatalf("ParseCombinedJSON empty contracts error: %v", err)
	}
	if len(contracts) != 0 {
		t.Errorf("got %d contracts, want 0", len(contracts))
	}
}

func TestParseCombinedJSONInvalidABI(t *testing.T) {
	// Valid outer JSON but ABI field is not valid JSON
	input := `{
		"contracts": {
			"foo.sol:Bad": {
				"abi": "not-valid-json",
				"bin": "",
				"bin-runtime": "",
				"srcmap": "",
				"srcmap-runtime": "",
				"userdoc": "{}",
				"devdoc": "{}",
				"metadata": ""
			}
		},
		"version": "0.7.6"
	}`
	_, err := ParseCombinedJSON([]byte(input), "", "", "", "")
	if err == nil {
		t.Error("ParseCombinedJSON with invalid ABI JSON: want error, got nil")
	}
}

func TestContractCodePrefix(t *testing.T) {
	// Verify that Code and RuntimeCode are always prefixed with "0x"
	input := `{
		"contracts": {
			"test.sol:C": {
				"abi": "[]",
				"bin": "deadbeef",
				"bin-runtime": "cafebabe",
				"srcmap": "",
				"srcmap-runtime": "",
				"userdoc": "{}",
				"devdoc": "{}",
				"metadata": ""
			}
		},
		"version": "0.7.6"
	}`
	contracts, err := ParseCombinedJSON([]byte(input), "", "", "", "")
	if err != nil {
		t.Fatalf("ParseCombinedJSON error: %v", err)
	}
	c := contracts["test.sol:C"]
	if c.Code != "0xdeadbeef" {
		t.Errorf("Code = %q, want %q", c.Code, "0xdeadbeef")
	}
	if c.RuntimeCode != "0xcafebabe" {
		t.Errorf("RuntimeCode = %q, want %q", c.RuntimeCode, "0xcafebabe")
	}
}

func TestContractInfoFields(t *testing.T) {
	// Verify ContractInfo JSON serialisation round-trip
	info := ContractInfo{
		Source:          "src",
		Language:        "Solidity",
		LanguageVersion: "0.7.6",
		CompilerVersion: "0.7.6+commit",
		CompilerOptions: "--optimize",
		SrcMap:          "0:10:0:-",
		SrcMapRuntime:   "0:10:0:-:0",
		AbiDefinition:   []interface{}{},
		UserDoc:         map[string]interface{}{"methods": map[string]interface{}{}},
		DeveloperDoc:    map[string]interface{}{"methods": map[string]interface{}{}},
		Metadata:        `{"compiler":{"version":"0.7.6"}}`,
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal(ContractInfo) error: %v", err)
	}

	var decoded ContractInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(ContractInfo) error: %v", err)
	}

	if decoded.Source != info.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, info.Source)
	}
	if decoded.Language != info.Language {
		t.Errorf("Language = %q, want %q", decoded.Language, info.Language)
	}
	if decoded.CompilerVersion != info.CompilerVersion {
		t.Errorf("CompilerVersion = %q, want %q", decoded.CompilerVersion, info.CompilerVersion)
	}
}

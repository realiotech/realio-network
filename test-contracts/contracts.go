// Package contracts embeds compiled CosmWasm contract binaries used as
// fixtures in integration tests.
package contracts

import _ "embed"

// HackatomWasm is the "hackatom" example contract from CosmWasm/wasmd, used
// to exercise the full store/instantiate/execute/query lifecycle in tests.
// It is instantiated with a verifier and a beneficiary address: only the
// verifier can execute "release", which sends the contract's balance to the
// beneficiary.
//
//go:embed hackatom.wasm
var HackatomWasm []byte

package chainfee

import (
	"fmt"

	"github.com/ltcsuite/ltcd/blockchain"
	"github.com/ltcsuite/ltcd/ltcutil"
)

const (
	// FeePerKwFloor is the lowest fee rate in sat/kw that we should use for
	// estimating transaction fees before signing.
	FeePerKwFloor SatPerKWeight = 253

	// AbsoluteFeePerKwFloor is the lowest fee rate in sat/kw of a
	// transaction that we should ever _create_. This is the the equivalent
	// of 1 sat/byte in sat/kw.
	AbsoluteFeePerKwFloor SatPerKWeight = 250
)

// SatPerKVByte represents a fee rate in sat/kb.
type SatPerKVByte ltcutil.Amount

// FeeForVSize calculates the fee resulting from this fee rate and the given
// vsize in vbytes.
func (s SatPerKVByte) FeeForVSize(vbytes int64) ltcutil.Amount {
	return ltcutil.Amount(s) * ltcutil.Amount(vbytes) / 1000
}

// FeePerKWeight converts the current fee rate from sat/kb to sat/kw.
func (s SatPerKVByte) FeePerKWeight() SatPerKWeight {
	return SatPerKWeight(s / blockchain.WitnessScaleFactor)
}

// String returns a human-readable string of the fee rate.
func (s SatPerKVByte) String() string {
	return fmt.Sprintf("%v sat/kb", int64(s))
}

// SatPerKWeight represents a fee rate in sat/kw.
type SatPerKWeight ltcutil.Amount

// FeeForWeight calculates the fee resulting from this fee rate and the given
// weight in weight units (wu).
func (s SatPerKWeight) FeeForWeight(wu int64) ltcutil.Amount {
	// The resulting fee is rounded down, as specified in BOLT#03.
	return ltcutil.Amount(s) * ltcutil.Amount(wu) / 1000
}

// FeePerKVByte converts the current fee rate from sat/kw to sat/kb.
func (s SatPerKWeight) FeePerKVByte() SatPerKVByte {
	return SatPerKVByte(s * blockchain.WitnessScaleFactor)
}

// String returns a human-readable string of the fee rate.
func (s SatPerKWeight) String() string {
	return fmt.Sprintf("%v sat/kw", int64(s))
}

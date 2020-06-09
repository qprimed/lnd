package input

import (
	"github.com/ltcsuite/ltcd/txscript"
	"github.com/ltcsuite/ltcd/wire"
)

// Input represents an abstract UTXO which is to be spent using a sweeping
// transaction. The method provided give the caller all information needed to
// construct a valid input within a sweeping transaction to sweep this
// lingering UTXO.
type Input interface {
	// Outpoint returns the reference to the output being spent, used to
	// construct the corresponding transaction input.
	OutPoint() *wire.OutPoint

	// WitnessType returns an enum specifying the type of witness that must
	// be generated in order to spend this output.
	WitnessType() WitnessType

	// SignDesc returns a reference to a spendable output's sign
	// descriptor, which is used during signing to compute a valid witness
	// that spends this output.
	SignDesc() *SignDescriptor

	// CraftInputScript returns a valid set of input scripts allowing this
	// output to be spent. The returns input scripts should target the
	// input at location txIndex within the passed transaction. The input
	// scripts generated by this method support spending p2wkh, p2wsh, and
	// also nested p2sh outputs.
	CraftInputScript(signer Signer, txn *wire.MsgTx,
		hashCache *txscript.TxSigHashes,
		txinIdx int) (*Script, error)

	// BlocksToMaturity returns the relative timelock, as a number of
	// blocks, that must be built on top of the confirmation height before
	// the output can be spent. For non-CSV locked inputs this is always
	// zero.
	BlocksToMaturity() uint32

	// HeightHint returns the minimum height at which a confirmed spending
	// tx can occur.
	HeightHint() uint32
}

type inputKit struct {
	outpoint        wire.OutPoint
	witnessType     WitnessType
	signDesc        SignDescriptor
	heightHint      uint32
	blockToMaturity uint32
}

// OutPoint returns the breached output's identifier that is to be included as
// a transaction input.
func (i *inputKit) OutPoint() *wire.OutPoint {
	return &i.outpoint
}

// WitnessType returns the type of witness that must be generated to spend the
// breached output.
func (i *inputKit) WitnessType() WitnessType {
	return i.witnessType
}

// SignDesc returns the breached output's SignDescriptor, which is used during
// signing to compute the witness.
func (i *inputKit) SignDesc() *SignDescriptor {
	return &i.signDesc
}

// HeightHint returns the minimum height at which a confirmed spending
// tx can occur.
func (i *inputKit) HeightHint() uint32 {
	return i.heightHint
}

// BlocksToMaturity returns the relative timelock, as a number of blocks, that
// must be built on top of the confirmation height before the output can be
// spent. For non-CSV locked inputs this is always zero.
func (i *inputKit) BlocksToMaturity() uint32 {
	return i.blockToMaturity
}

// BaseInput contains all the information needed to sweep a basic output
// (CSV/CLTV/no time lock)
type BaseInput struct {
	inputKit
}

// MakeBaseInput assembles a new BaseInput that can be used to construct a
// sweep transaction.
func MakeBaseInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32) BaseInput {

	return BaseInput{
		inputKit{
			outpoint:    *outpoint,
			witnessType: witnessType,
			signDesc:    *signDescriptor,
			heightHint:  heightHint,
		},
	}
}

// NewBaseInput allocates and assembles a new *BaseInput that can be used to
// construct a sweep transaction.
func NewBaseInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32) *BaseInput {

	input := MakeBaseInput(
		outpoint, witnessType, signDescriptor, heightHint,
	)

	return &input
}

// NewCsvInput assembles a new csv-locked input that can be used to
// construct a sweep transaction.
func NewCsvInput(outpoint *wire.OutPoint, witnessType WitnessType,
	signDescriptor *SignDescriptor, heightHint uint32,
	blockToMaturity uint32) *BaseInput {

	return &BaseInput{
		inputKit{
			outpoint:        *outpoint,
			witnessType:     witnessType,
			signDesc:        *signDescriptor,
			heightHint:      heightHint,
			blockToMaturity: blockToMaturity,
		},
	}
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returned input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (bi *BaseInput) CraftInputScript(signer Signer, txn *wire.MsgTx,
	hashCache *txscript.TxSigHashes, txinIdx int) (*Script, error) {

	witnessFunc := bi.witnessType.WitnessGenerator(signer, bi.SignDesc())

	return witnessFunc(txn, hashCache, txinIdx)
}

// HtlcSucceedInput constitutes a sweep input that needs a pre-image. The input
// is expected to reside on the commitment tx of the remote party and should
// not be a second level tx output.
type HtlcSucceedInput struct {
	inputKit

	preimage []byte
}

// MakeHtlcSucceedInput assembles a new redeem input that can be used to
// construct a sweep transaction.
func MakeHtlcSucceedInput(outpoint *wire.OutPoint,
	signDescriptor *SignDescriptor, preimage []byte, heightHint,
	blocksToMaturity uint32) HtlcSucceedInput {

	return HtlcSucceedInput{
		inputKit: inputKit{
			outpoint:        *outpoint,
			witnessType:     HtlcAcceptedRemoteSuccess,
			signDesc:        *signDescriptor,
			heightHint:      heightHint,
			blockToMaturity: blocksToMaturity,
		},
		preimage: preimage,
	}
}

// CraftInputScript returns a valid set of input scripts allowing this output
// to be spent. The returns input scripts should target the input at location
// txIndex within the passed transaction. The input scripts generated by this
// method support spending p2wkh, p2wsh, and also nested p2sh outputs.
func (h *HtlcSucceedInput) CraftInputScript(signer Signer, txn *wire.MsgTx,
	hashCache *txscript.TxSigHashes, txinIdx int) (*Script, error) {

	desc := h.signDesc
	desc.SigHashes = hashCache
	desc.InputIndex = txinIdx

	witness, err := SenderHtlcSpendRedeem(
		signer, &desc, txn, h.preimage,
	)
	if err != nil {
		return nil, err
	}

	return &Script{
		Witness: witness,
	}, nil
}

// Compile-time constraints to ensure each input struct implement the Input
// interface.
var _ Input = (*BaseInput)(nil)
var _ Input = (*HtlcSucceedInput)(nil)

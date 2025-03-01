package netann

import (
	"image/color"
	"net"
	"time"

	"github.com/ltcsuite/lnd/keychain"
	"github.com/ltcsuite/lnd/lnwallet"
	"github.com/ltcsuite/lnd/lnwire"
)

// NodeAnnModifier is a closure that makes in-place modifications to an
// lnwire.NodeAnnouncement.
type NodeAnnModifier func(*lnwire.NodeAnnouncement)

// NodeAnnSetAlias is a functional option that sets the alias of the
// given node announcement.
func NodeAnnSetAlias(alias lnwire.NodeAlias) func(*lnwire.NodeAnnouncement) {
	return func(nodeAnn *lnwire.NodeAnnouncement) {
		nodeAnn.Alias = alias
	}
}

// NodeAnnSetAddrs is a functional option that allows updating the addresses of
// the given node announcement.
func NodeAnnSetAddrs(addrs []net.Addr) func(*lnwire.NodeAnnouncement) {
	return func(nodeAnn *lnwire.NodeAnnouncement) {
		nodeAnn.Addresses = addrs
	}
}

// NodeAnnSetColor is a functional option that sets the color of the
// given node announcement.
func NodeAnnSetColor(newColor color.RGBA) func(*lnwire.NodeAnnouncement) {
	return func(nodeAnn *lnwire.NodeAnnouncement) {
		nodeAnn.RGBColor = newColor
	}
}

// NodeAnnSetFeatures is a functional option that allows updating the features of
// the given node announcement.
func NodeAnnSetFeatures(features *lnwire.RawFeatureVector) func(*lnwire.NodeAnnouncement) {
	return func(nodeAnn *lnwire.NodeAnnouncement) {
		nodeAnn.Features = features
	}
}

// NodeAnnSetTimestamp is a functional option that sets the timestamp of the
// announcement to the current time, or increments it if the timestamp is
// already in the future.
func NodeAnnSetTimestamp(nodeAnn *lnwire.NodeAnnouncement) {
	newTimestamp := uint32(time.Now().Unix())
	if newTimestamp <= nodeAnn.Timestamp {
		// Increment the prior value to  ensure the timestamp
		// monotonically increases, otherwise the announcement won't
		// propagate.
		newTimestamp = nodeAnn.Timestamp + 1
	}
	nodeAnn.Timestamp = newTimestamp
}

// SignNodeAnnouncement signs the lnwire.NodeAnnouncement provided, which
// should be the most recent, valid update, otherwise the timestamp may not
// monotonically increase from the prior.
func SignNodeAnnouncement(signer lnwallet.MessageSigner,
	keyLoc keychain.KeyLocator, nodeAnn *lnwire.NodeAnnouncement) error {

	// Create the DER-encoded ECDSA signature over the message digest.
	sig, err := SignAnnouncement(signer, keyLoc, nodeAnn)
	if err != nil {
		return err
	}

	// Parse the DER-encoded signature into a fixed-size 64-byte array.
	nodeAnn.Signature, err = lnwire.NewSigFromSignature(sig)
	return err
}

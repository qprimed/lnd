//go:build !bitcoind && !neutrino
// +build !bitcoind,!neutrino

package lntest

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ltcsuite/lnd/lntest/node"
	"github.com/ltcsuite/ltcd/btcjson"
	"github.com/ltcsuite/ltcd/chaincfg"
	"github.com/ltcsuite/ltcd/integration/rpctest"
	"github.com/ltcsuite/ltcd/rpcclient"
)

// logDirPattern is the pattern of the name of the temporary log directory.
const logDirPattern = "%s/.backendlogs"

// BtcdBackendConfig is an implementation of the BackendConfig interface
// backed by a ltcd node.
type BtcdBackendConfig struct {
	// rpcConfig houses the connection config to the backing ltcd instance.
	rpcConfig rpcclient.ConnConfig

	// harness is the backing ltcd instance.
	harness *rpctest.Harness

	// minerAddr is the p2p address of the miner to connect to.
	minerAddr string
}

// A compile time assertion to ensure BtcdBackendConfig meets the BackendConfig
// interface.
var _ node.BackendConfig = (*BtcdBackendConfig)(nil)

// GenArgs returns the arguments needed to be passed to LND at startup for
// using this node as a chain backend.
func (b BtcdBackendConfig) GenArgs() []string {
	var args []string
	encodedCert := hex.EncodeToString(b.rpcConfig.Certificates)
	args = append(args, "--litecoin.node=ltcd")
	args = append(args, fmt.Sprintf("--ltcd.rpchost=%v", b.rpcConfig.Host))
	args = append(args, fmt.Sprintf("--ltcd.rpcuser=%v", b.rpcConfig.User))
	args = append(args, fmt.Sprintf("--ltcd.rpcpass=%v", b.rpcConfig.Pass))
	args = append(args, fmt.Sprintf("--ltcd.rawrpccert=%v", encodedCert))

	return args
}

// ConnectMiner is called to establish a connection to the test miner.
func (b BtcdBackendConfig) ConnectMiner() error {
	return b.harness.Client.Node(btcjson.NConnect, b.minerAddr, &temp)
}

// DisconnectMiner is called to disconnect the miner.
func (b BtcdBackendConfig) DisconnectMiner() error {
	return b.harness.Client.Node(btcjson.NDisconnect, b.minerAddr, &temp)
}

// Credentials returns the rpc username, password and host for the backend.
func (b BtcdBackendConfig) Credentials() (string, string, string, error) {
	return b.rpcConfig.User, b.rpcConfig.Pass, b.rpcConfig.Host, nil
}

// Name returns the name of the backend type.
func (b BtcdBackendConfig) Name() string {
	return "ltcd"
}

// NewBackend starts a new rpctest.Harness and returns a BtcdBackendConfig for
// that node. miner should be set to the P2P address of the miner to connect
// to.
func NewBackend(miner string, netParams *chaincfg.Params) (
	*BtcdBackendConfig, func() error, error) {

	baseLogDir := fmt.Sprintf(logDirPattern, node.GetLogDir())
	args := []string{
		"--rejectnonstd",
		"--txindex",
		"--trickleinterval=100ms",
		"--debuglevel=debug",
		"--logdir=" + baseLogDir,
		"--nowinservice",
		// The miner will get banned and disconnected from the node if
		// its requested data are not found. We add a nobanning flag to
		// make sure they stay connected if it happens.
		"--nobanning",
		// Don't disconnect if a reply takes too long.
		"--nostalldetect",
	}
	chainBackend, err := rpctest.New(
		netParams, nil, args, node.GetBtcdBinary(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create ltcd node: %w",
			err)
	}

	// We want to overwrite some of the connection settings to make the
	// tests more robust. We might need to restart the backend while there
	// are already blocks present, which will take a bit longer than the
	// 1 second the default settings amount to. Doubling both values will
	// give us retries up to 4 seconds.
	const (
		maxConnRetries   = rpctest.DefaultMaxConnectionRetries * 2
		connRetryTimeout = rpctest.DefaultConnectionRetryTimeout * 2
	)

	chainBackend.MaxConnRetries = maxConnRetries
	chainBackend.ConnectionRetryTimeout = connRetryTimeout

	if err := chainBackend.SetUp(false, 0); err != nil {
		return nil, nil, fmt.Errorf("unable to set up ltcd backend: %w",
			err)
	}

	bd := &BtcdBackendConfig{
		rpcConfig: chainBackend.RPCConfig(),
		harness:   chainBackend,
		minerAddr: miner,
	}

	cleanUp := func() error {
		var errStr string
		if err := chainBackend.TearDown(); err != nil {
			errStr += err.Error() + "\n"
		}

		// After shutting down the chain backend, we'll make a copy of
		// the log files, including any compressed log files from
		// logrorate, before deleting the temporary log dir.
		logDir := fmt.Sprintf("%s/%s", baseLogDir, netParams.Name)
		files, err := ioutil.ReadDir(logDir)
		if err != nil {
			errStr += fmt.Sprintf(
				"unable to read log directory: %v\n", err,
			)
		}

		for _, file := range files {
			logFile := fmt.Sprintf("%s/%s", logDir, file.Name())
			newFilename := strings.Replace(
				file.Name(), "ltcd.log",
				"output_ltcd_chainbackend.log", 1,
			)
			logDestination := fmt.Sprintf(
				"%s/%s", node.GetLogDir(), newFilename,
			)
			err := node.CopyFile(logDestination, logFile)
			if err != nil {
				errStr += fmt.Sprintf("unable to copy file: "+
					"%v\n", err)
			}
		}

		if err = os.RemoveAll(baseLogDir); err != nil {
			errStr += fmt.Sprintf(
				"cannot remove dir %s: %v\n", baseLogDir, err,
			)
		}
		if errStr != "" {
			return errors.New(errStr)
		}
		return nil
	}

	return bd, cleanUp, nil
}

package kvdb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ltcsuite/ltcwallet/walletdb"
	"github.com/stretchr/testify/require"
)

type boltFixture struct {
	t       *testing.T
	tempDir string
}

func NewBoltFixture(t *testing.T) *boltFixture {
	tempDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)

	return &boltFixture{
		t:       t,
		tempDir: tempDir,
	}
}

func (b *boltFixture) Cleanup() {
	os.RemoveAll(b.tempDir)
}

func (b *boltFixture) NewBackend() walletdb.DB {
	dbPath := filepath.Join(b.tempDir)

	db, err := GetBoltBackend(&BoltBackendConfig{
		DBPath:         dbPath,
		DBFileName:     "test.db",
		NoFreelistSync: true,
		DBTimeout:      DefaultDBTimeout,
	})
	require.NoError(b.t, err)

	return db
}

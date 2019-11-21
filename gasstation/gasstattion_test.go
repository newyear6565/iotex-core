// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package gasstation

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/action/protocol/account"
	"github.com/iotexproject/iotex-core/action/protocol/execution"
	"github.com/iotexproject/iotex-core/action/protocol/rolldpos"
	"github.com/iotexproject/iotex-core/blockchain"
	"github.com/iotexproject/iotex-core/blockchain/blockdao"
	"github.com/iotexproject/iotex-core/blockindex"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/pkg/unit"
	"github.com/iotexproject/iotex-core/pkg/version"
	"github.com/iotexproject/iotex-core/state/factory"
	"github.com/iotexproject/iotex-core/test/identityset"
	"github.com/iotexproject/iotex-core/testutil"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"
)

func TestNewGasStation(t *testing.T) {
	require := require.New(t)
	require.NotNil(NewGasStation(nil, config.Default.API))
}

func TestSuggestGasPriceForUserAction(t *testing.T) {
	ctx := context.Background()
	cfg := newConfig()
	bc := newBlockchain(cfg, t)
	require.NoError(t, bc.Start(ctx))
	defer func() {
		require.NoError(t, bc.Stop(ctx))
	}()

	for i := 0; i < 30; i++ {
		tsf, err := action.NewTransfer(
			uint64(i)+1,
			big.NewInt(100),
			identityset.Address(27).String(),
			[]byte{}, uint64(100000),
			big.NewInt(1).Mul(big.NewInt(int64(i)+10), big.NewInt(unit.Qev)),
		)
		require.NoError(t, err)

		bd := &action.EnvelopeBuilder{}
		elp1 := bd.SetAction(tsf).
			SetNonce(uint64(i) + 1).
			SetGasLimit(100000).
			SetGasPrice(big.NewInt(1).Mul(big.NewInt(int64(i)+10), big.NewInt(unit.Qev))).Build()
		selp1, err := action.Sign(elp1, identityset.PrivateKey(0))
		require.NoError(t, err)

		actionMap := make(map[string][]action.SealedEnvelope)
		actionMap[identityset.Address(0).String()] = []action.SealedEnvelope{selp1}

		blk, err := bc.MintNewBlock(
			actionMap,
			testutil.TimestampNow(),
		)
		require.NoError(t, err)
		require.Equal(t, 2, len(blk.Actions))
		require.Equal(t, 1, len(blk.Receipts))
		var gasConsumed uint64
		for _, receipt := range blk.Receipts {
			gasConsumed += receipt.GasConsumed
		}
		require.True(t, gasConsumed <= cfg.Genesis.BlockGasLimit)
		err = bc.ValidateBlock(blk)
		require.NoError(t, err)
		err = bc.CommitBlock(blk)
		require.NoError(t, err)
	}
	height := bc.TipHeight()
	fmt.Printf("Open blockchain pass, height = %d\n", height)

	gs := NewGasStation(bc, cfg.API)
	require.NotNil(t, gs)

	gp, err := gs.SuggestGasPrice()
	require.NoError(t, err)
	// i from 10 to 29,gasprice for 20 to 39,60%*20+20=31
	require.Equal(t, big.NewInt(1).Mul(big.NewInt(int64(31)), big.NewInt(unit.Qev)).Uint64(), gp)
}

func TestSuggestGasPriceForSystemAction(t *testing.T) {
	ctx := context.Background()
	cfg := newConfig()
	cfg.Genesis.BlockGasLimit = uint64(100000)
	bc := newBlockchain(cfg, t)
	require.NoError(t, bc.Start(ctx))
	defer func() {
		require.NoError(t, bc.Stop(ctx))
	}()

	for i := 0; i < 30; i++ {
		actionMap := make(map[string][]action.SealedEnvelope)

		blk, err := bc.MintNewBlock(
			actionMap,
			testutil.TimestampNow(),
		)
		require.NoError(t, err)
		require.Equal(t, 1, len(blk.Actions))
		require.Equal(t, 0, len(blk.Receipts))
		var gasConsumed uint64
		for _, receipt := range blk.Receipts {
			gasConsumed += receipt.GasConsumed
		}
		require.True(t, gasConsumed <= cfg.Genesis.BlockGasLimit)
		err = bc.ValidateBlock(blk)
		require.NoError(t, err)
		err = bc.CommitBlock(blk)
		require.NoError(t, err)
	}
	height := bc.TipHeight()
	fmt.Printf("Open blockchain pass, height = %d\n", height)

	gs := NewGasStation(bc, cfg.API)
	require.NotNil(t, gs)

	gp, err := gs.SuggestGasPrice()
	fmt.Println(gp)
	require.NoError(t, err)
	// i from 10 to 29,gasprice for 20 to 39,60%*20+20=31
	require.Equal(t, gs.cfg.GasStation.DefaultGas, gp)
}

func TestEstimateGasForAction(t *testing.T) {
	require := require.New(t)
	act := getAction()
	require.NotNil(act)
	cfg := newConfig()
	bc := newBlockchain(cfg, t)
	require.NoError(bc.Start(context.Background()))
	require.NotNil(bc)
	gs := NewGasStation(bc, config.Default.API)
	require.NotNil(gs)
	ret, err := gs.EstimateGasForAction(act)
	require.NoError(err)
	// base intrinsic gas 10000
	require.Equal(uint64(10000), ret)

	// test for payload
	act = getActionWithPayload()
	require.NotNil(act)
	require.NoError(bc.Start(context.Background()))
	require.NotNil(bc)
	ret, err = gs.EstimateGasForAction(act)
	require.NoError(err)
	// base intrinsic gas 10000,plus data size*ExecutionDataGas
	require.Equal(uint64(10000)+10*action.ExecutionDataGas, ret)
}

func getAction() (act *iotextypes.Action) {
	pubKey1 := identityset.PrivateKey(28).PublicKey()
	addr2 := identityset.Address(29).String()

	act = &iotextypes.Action{
		Core: &iotextypes.ActionCore{
			Action: &iotextypes.ActionCore_Transfer{
				Transfer: &iotextypes.Transfer{Recipient: addr2},
			},
			Version: version.ProtocolVersion,
			Nonce:   101,
		},
		SenderPubKey: pubKey1.Bytes(),
	}
	return
}

func getActionWithPayload() (act *iotextypes.Action) {
	pubKey1 := identityset.PrivateKey(28).PublicKey()
	addr2 := identityset.Address(29).String()

	act = &iotextypes.Action{
		Core: &iotextypes.ActionCore{
			Action: &iotextypes.ActionCore_Transfer{
				Transfer: &iotextypes.Transfer{Recipient: addr2, Payload: []byte("1234567890")},
			},
			Version: version.ProtocolVersion,
			Nonce:   101,
		},
		SenderPubKey: pubKey1.Bytes(),
	}
	return
}

func newConfig() config.Config {
	cfg := config.Default

	testTrieFile, _ := ioutil.TempFile(os.TempDir(), "trie")
	testTriePath := testTrieFile.Name()
	testDBFile, _ := ioutil.TempFile(os.TempDir(), "db")
	testDBPath := testDBFile.Name()
	testIndexFile, _ := ioutil.TempFile(os.TempDir(), "index")
	testIndexPath := testIndexFile.Name()

	cfg.Plugins[config.GatewayPlugin] = true
	cfg.Chain.TrieDBPath = testTriePath
	cfg.Chain.ChainDBPath = testDBPath
	cfg.Chain.IndexDBPath = testIndexPath
	cfg.Chain.EnableAsyncIndexWrite = false
	cfg.Genesis.EnableGravityChainVoting = true
	cfg.ActPool.MinGasPriceStr = "0"
	cfg.API.RangeQueryLimit = 100

	return cfg
}

func newBlockchain(cfg config.Config, t *testing.T) blockchain.Blockchain {
	registry := protocol.Registry{}
	acc := account.NewProtocol()
	require.NoError(t, registry.Register(account.ProtocolID, acc))
	rp := rolldpos.NewProtocol(cfg.Genesis.NumCandidateDelegates, cfg.Genesis.NumDelegates, cfg.Genesis.NumSubEpochs)
	require.NoError(t, registry.Register(rolldpos.ProtocolID, rp))
	dbConfig := cfg.DB
	sf, err := factory.NewFactory(cfg, factory.DefaultTrieOption())
	require.NoError(t, err)
	// create indexer
	dbConfig.DbPath = cfg.Chain.IndexDBPath
	indexer, err := blockindex.NewIndexer(db.NewBoltDB(dbConfig), cfg.Genesis.Hash())
	require.NoError(t, err)
	// create BlockDAO
	dbConfig.DbPath = cfg.Chain.ChainDBPath
	dao := blockdao.NewBlockDAO(db.NewBoltDB(dbConfig), indexer, cfg.Chain.CompressBlock, dbConfig)
	require.NotNil(t, dao)
	bc := blockchain.NewBlockchain(
		cfg,
		dao,
		blockchain.PrecreatedStateFactoryOption(sf),
		blockchain.RegistryOption(&registry),
	)
	require.NotNil(t, bc)
	bc.Validator().AddActionEnvelopeValidators(protocol.NewGenericValidator(bc.Factory().Nonce))
	exec := execution.NewProtocol(bc.BlockDAO().GetBlockHash)
	require.NoError(t, registry.Register(execution.ProtocolID, exec))
	bc.Validator().AddActionValidators(acc, exec)
	return bc
}

package indexer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"time"

	"github.com/crypto4people/zoro/eth"
	"github.com/crypto4people/zoro/rate"
	"github.com/crypto4people/zoro/rate/contract"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/itsabgr/retry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var errSleep = errors.New("sleep")

type IndexerInfo struct {
	Uptime    time.Duration
	Version   string
	Ratefeed  common.Address
	Chain     string
	Height    int64
	BlockTime time.Duration
}

type Indexer struct {
	client    *ethclient.Client
	db        *mongo.Database
	chain     string
	ratefeed  common.Address
	start     time.Time
	blockTime time.Duration
}

type transfersDoc struct {
	Block int64          `bson:"block"`
	To    common.Address `bson:"to"`
	Tx    []byte         `bson:"tx"`
	Hash  common.Hash    `bson:"hash"`
	Chain string         `bson:"chain"`
}

type rateDoc struct {
	U        []byte         `bson:"u"`
	Block    int64          `bson:"block"`
	Ratefeed common.Address `bson:"ratefeed"`
	Chain    string         `bson:"chain"`
	Sub      []byte         `bson:"sub"`
	Div      []byte         `bson:"div"`
	Mul      []byte         `bson:"mul"`
	Add      []byte         `bson:"add"`
}

func (indexer *Indexer) Info(ctx context.Context) (*IndexerInfo, error) {
	transfer, err := indexer.getLastTransfer(ctx)
	if err != nil {
		return nil, err
	}

	return &IndexerInfo{
		Uptime:    time.Since(indexer.start),
		Version:   "v1",
		Ratefeed:  indexer.ratefeed,
		Chain:     indexer.chain,
		Height:    transfer.Block,
		BlockTime: indexer.blockTime,
	}, nil
}

func (indexer *Indexer) RateAt(ctx context.Context, chain string, block int64, ratefeed common.Address) (*rate.Rate, error) {
	var record rateDoc

	filter := bson.D{
		{Key: "block", Value: block},
		{Key: "chain", Value: chain},
		{Key: "ratefeed", Value: ratefeed},
	}
	opts := options.FindOne().SetSort(bson.D{{Key: "block", Value: -1}})

	err := indexer.db.Collection("rates").FindOne(ctx, filter, opts).Decode(&record)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	res := &rate.Rate{}
	res.Sub.SetBytes(record.Sub)
	res.Div.SetBytes(record.Div)
	res.Mul.SetBytes(record.Mul)
	res.Add.SetBytes(record.Add)

	return res, nil
}

func (indexer *Indexer) GetTx(ctx context.Context, hash common.Hash) ([]byte, error) {
	var transfer transfersDoc

	filter := bson.D{
		{Key: "hash", Value: hash},
	}

	err := indexer.db.Collection("transfers").FindOne(ctx, filter).Decode(&transfer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return transfer.Tx, nil
}

func (indexer *Indexer) Receives(ctx context.Context, chain string, block int64, accounts map[common.Address]struct{}) ([][]byte, error) {
	addrs := make([]common.Address, 0, len(accounts))
	for addr := range accounts {
		addrs = append(addrs, addr)
	}

	filter := bson.D{
		{Key: "chain", Value: chain},
		{Key: "block", Value: block},
		{Key: "to", Value: bson.D{
			{Key: "$in", Value: addrs},
		}},
	}

	cur, err := indexer.db.Collection("transfers").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var txs [][]byte
	for cur.Next(ctx) {
		var doc transfersDoc
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		txs = append(txs, doc.Tx)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return txs, nil
}

func (indexer *Indexer) Routine(ctx context.Context) error {
	height, err := indexer.getLocalHeight(ctx)
	if err != nil {
		return err
	}

	for {
		err = indexer.foreachSafeBlock(ctx, height, func(ctx context.Context, number int64) error {
			blockData, err := indexer.getBlockData(ctx, number)
			if err != nil {
				return err
			}

			for _, tx := range blockData.Txs {
				if err := indexer.appendTransfer(ctx, number, tx); err != nil {
					// todo log
					return err
				}
			}

			for _, log := range blockData.Logs {
				rateUpdatedEvent, err := contract.ABI().UnpackRateUpdatedEvent(&log)
				if err != nil {
					return err
				}

				rate := &rate.Rate{}
				rate.Sub.SetFromBig(rateUpdatedEvent.Sub)
				rate.Div.SetFromBig(rateUpdatedEvent.Div)
				rate.Mul.SetFromBig(rateUpdatedEvent.Mul)
				rate.Add.SetFromBig(rateUpdatedEvent.Add)

				if err := indexer.appendRate(ctx, number, rate); err != nil {
					// todo log
					return err
				}
			}

			return nil
		})
		if err != nil {
			if !errors.Is(err, errSleep) {
				return err
			}
			retry.Sleep(ctx, time.Second)
		}
	}
}

type getBlockDataResult struct {
	Txs  []*types.Transaction
	Logs []types.Log
}

func (indexer *Indexer) getBlockData(ctx context.Context, number int64) (*getBlockDataResult, error) {
	block, err := indexer.client.BlockByNumber(ctx, big.NewInt(number))
	if err != nil {
		return nil, err
	}

	hash := block.Hash()
	filter := ethereum.FilterQuery{
		BlockHash: &hash,
		Addresses: []common.Address{indexer.ratefeed},
		Topics:    [][]common.Hash{{contract.RateChangedID()}},
	}

	logs, err := indexer.client.FilterLogs(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &getBlockDataResult{
		Txs:  block.Transactions(),
		Logs: logs,
	}, nil
}

func (indexer *Indexer) foreachSafeBlock(ctx context.Context, start int64, fn func(ctx context.Context, number int64) error) error {
	finalHdr, err := indexer.client.HeaderByNumber(ctx, big.NewInt(eth.FinalBlock))
	if err != nil {
		return err
	}

	if !finalHdr.Number.IsInt64() {
		return errors.New("block number should be an int64")
	}

	target := finalHdr.Number.Int64()

	if target == start {
		return errSleep
	}

	if target < start {
		return errors.New("safe block number cannot be more than start")
	}

	for range target - start {
		err := fn(ctx, start)
		if err != nil {
			return err
		}
		start++
	}

	return nil
}

func (indexer *Indexer) getLastTransfer(ctx context.Context) (*transfersDoc, error) {
	var transfer transfersDoc

	filter := bson.D{
		{Key: "chain", Value: indexer.chain},
	}
	opts := options.FindOne().SetSort(bson.D{
		{Key: "block", Value: -1},
	})

	err := indexer.db.Collection("transfers").FindOne(ctx, filter, opts).Decode(&transfer)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &transfer, nil
}

func (indexer *Indexer) getLocalHeight(ctx context.Context) (int64, error) {
	transfer, err := indexer.getLastTransfer(ctx)
	if err != nil {
		return -1, err
	}

	if transfer == nil {
		return 0, nil
	}

	start := transfer.Block - 1
	if start < 0 {
		start = 0
	}

	return start, nil
}

func (indexer *Indexer) appendTransfer(ctx context.Context, block int64, tx *types.Transaction) error {
	rawBytes, err := tx.MarshalBinary()
	if err != nil {
		return err
	}

	to := tx.To()
	if to == nil {
		return nil
	}

	newTransfer := &transfersDoc{
		Block: block,
		To:    *to,
		Tx:    rawBytes,
		Hash:  tx.Hash(),
		Chain: indexer.chain,
	}

	filter := bson.D{
		{Key: "hash", Value: tx.Hash()},
	}
	opts := options.Replace().SetUpsert(true)

	_, err = indexer.db.Collection("transfers").ReplaceOne(ctx, filter, newTransfer, opts)
	return err
}

func (indexer *Indexer) appendRate(ctx context.Context, block int64, rate *rate.Rate) error {
	hash := rateUniqueHash(block, indexer.chain, indexer.ratefeed, rate)

	newRate := &rateDoc{
		U:        hash,
		Block:    block,
		Ratefeed: indexer.ratefeed,
		Chain:    indexer.chain,
		Sub:      rate.Sub.Bytes(),
		Div:      rate.Div.Bytes(),
		Mul:      rate.Mul.Bytes(),
		Add:      rate.Add.Bytes(),
	}

	filter := bson.D{
		{Key: "u", Value: hash},
	}
	opts := options.Replace().SetUpsert(true)

	_, err := indexer.db.Collection("rates").ReplaceOne(ctx, filter, newRate, opts)
	return err
}

func rateUniqueHash(block int64, chain string, ratefeed common.Address, rate *rate.Rate) []byte {
	hash := sha256.New()

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(block))
	hash.Write(buf[:])

	hash.Write([]byte(chain))
	hash.Write([]byte{0})

	hash.Write(ratefeed.Bytes())
	hash.Write([]byte{0})

	hash.Write(rate.Sub.Bytes())
	hash.Write([]byte{0})

	hash.Write(rate.Div.Bytes())
	hash.Write([]byte{0})

	hash.Write(rate.Mul.Bytes())
	hash.Write([]byte{0})

	hash.Write(rate.Add.Bytes())
	hash.Write([]byte{0})

	return hash.Sum(nil)
}

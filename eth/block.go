package eth

import "github.com/ethereum/go-ethereum/rpc"

const FinalBlock int64 = int64(rpc.FinalizedBlockNumber)
const LastBlock int64 = int64(rpc.LatestBlockNumber)
const PendBlock int64 = int64(rpc.PendingBlockNumber)

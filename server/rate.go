package server

import (
	"math"
	"math/big"
	"time"
)

type Rate struct {
	TTL   time.Duration
	Fee   *big.Int
	Price *big.Int
}

func (rate *Rate) Convert(value *big.Int) (int, time.Duration) {
	bigGas := new(big.Int).Sub(value, rate.Fee)
	bigGas = bigGas.Div(bigGas, rate.Price)
	if bigGas.Sign() <= 0 || bigGas.Cmp(big.NewInt(math.MaxInt32)) > 0 {
		return 0, 0
	}
	gas := int(bigGas.Int64())
	ttl := time.Duration(gas) * rate.TTL
	if ttl > math.MaxInt64/time.Duration(gas) {
		return 0, 0
	}
	return gas, ttl
}

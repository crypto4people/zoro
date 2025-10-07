package server

import (
	"math/big"
	"time"
)

type Rate struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
}

type convertResult struct {
	Dur time.Duration
	TTL time.Duration
}

var seconds = big.NewInt(int64(time.Second))

func (rate *Rate) convert(value *big.Int) convertResult {

	//dur := (value - sub) / div

	//ttl := (durSeconds * mul) + add

	dur := new(big.Int).Div(new(big.Int).Sub(value, rate.Sub), rate.Div)
	ttl := new(big.Int).Add(new(big.Int).Mul(dur, rate.Mul), rate.Add)

	dur = new(big.Int).Mul(dur, seconds)
	ttl = new(big.Int).Mul(ttl, seconds)

	if dur.Sign() <= 0 || ttl.Sign() <= 0 || !dur.IsInt64() || !ttl.IsInt64() {
		return convertResult{}
	}

	return convertResult{
		Dur: time.Duration(dur.Int64()),
		TTL: time.Duration(ttl.Int64()),
	}
}

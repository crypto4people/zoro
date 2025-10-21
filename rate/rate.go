package rate

import (
	"math/big"
	"time"
	"zoro/contract"
)

type Rate struct {
	Sub *big.Int
	Div *big.Int
	Mul *big.Int
	Add *big.Int
}

func FromGetRateOutput(out contract.GetRateOutput) Rate {
	return Rate{
		Sub: out.Sub,
		Div: out.Div,
		Mul: out.Mul,
		Add: out.Sub,
	}
}

func (rate *Rate) Apply(value *big.Int) (dur time.Duration, ttl time.Duration) {

	// dur := (value - sub) / div
	// ttl := (durSeconds * mul) + add

	durBig := new(big.Int).Div(new(big.Int).Sub(value, rate.Sub), rate.Div)
	ttlBig := new(big.Int).Add(new(big.Int).Mul(durBig, rate.Mul), rate.Add)

	durBig = new(big.Int).Mul(durBig, big.NewInt(int64(time.Second)))
	ttlBig = new(big.Int).Mul(ttlBig, big.NewInt(int64(time.Second)))

	if durBig.Sign() <= 0 || ttlBig.Sign() <= 0 || !durBig.IsInt64() || !ttlBig.IsInt64() {
		return 0, 0
	}

	return time.Duration(durBig.Int64()), time.Duration(ttlBig.Int64())

}

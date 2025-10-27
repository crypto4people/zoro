package rate

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

func (rate *Rate) Apply(value *big.Int) (dur time.Duration, ttl time.Duration) {

	// dur := (value - sub) / div
	// ttl := (dur * mul) + add

	durBig := new(big.Int).Div(new(big.Int).Sub(value, rate.Sub), rate.Div)
	ttlBig := new(big.Int).Add(new(big.Int).Mul(durBig, rate.Mul), rate.Add)

	durBig = new(big.Int).Mul(durBig, big.NewInt(int64(time.Second)))
	ttlBig = new(big.Int).Mul(ttlBig, big.NewInt(int64(time.Second)))

	if durBig.Sign() <= 0 || ttlBig.Sign() <= 0 || !durBig.IsInt64() || !ttlBig.IsInt64() {
		return 0, 0
	}

	return time.Duration(durBig.Int64()), time.Duration(ttlBig.Int64())

}

func (rate *Rate) Unapply(dur time.Duration, ttl time.Duration) *big.Int {
	ttlSec := new(big.Int).Div(big.NewInt(int64(ttl)), big.NewInt(int64(time.Second)))

	tmp := new(big.Int).Sub(ttlSec, rate.Add)
	if rate.Mul.Sign() == 0 {
		return big.NewInt(0)
	}
	durCalc := new(big.Int).Div(tmp, rate.Mul)

	value := new(big.Int).Add(new(big.Int).Mul(durCalc, rate.Div), rate.Sub)

	return value
}

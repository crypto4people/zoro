package rate

import (
	"math/big"
	"time"

	"github.com/holiman/uint256"
)

type Rate struct {
	Sub uint256.Int
	Div uint256.Int
	Mul uint256.Int
	Add uint256.Int
}

func (rate *Rate) Apply(value *big.Int) (dur time.Duration, ttl time.Duration) {

	// dur := (val - sub) / div
	// ttl := (dur * mul) + add

	durBig := new(big.Int).Div(new(big.Int).Sub(value, rate.Sub.ToBig()), rate.Div.ToBig())
	ttlBig := new(big.Int).Add(new(big.Int).Mul(durBig, rate.Mul.ToBig()), rate.Add.ToBig())

	durBig = new(big.Int).Mul(durBig, big.NewInt(int64(time.Second)))
	ttlBig = new(big.Int).Mul(ttlBig, big.NewInt(int64(time.Second)))

	if durBig.Sign() <= 0 || ttlBig.Sign() <= 0 || !durBig.IsInt64() || !ttlBig.IsInt64() {
		return 0, 0
	}

	return time.Duration(durBig.Int64()), time.Duration(ttlBig.Int64())

}

func (rate *Rate) Unapply(dur time.Duration, ttl time.Duration) *big.Int {
	ttlSec := new(big.Int).Div(big.NewInt(int64(ttl)), big.NewInt(int64(time.Second)))

	tmp := new(big.Int).Sub(ttlSec, rate.Add.ToBig())
	if rate.Mul.Sign() == 0 {
		return big.NewInt(0)
	}
	durCalc := new(big.Int).Div(tmp, rate.Mul.ToBig())

	value := new(big.Int).Add(new(big.Int).Mul(durCalc, rate.Div.ToBig()), rate.Sub.ToBig())

	return value
}

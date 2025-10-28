package misc

import (
	"bufio"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func LoadPrivateKey() (*ecdsa.PrivateKey, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, _ := io.ReadAll(os.Stdin)
		return crypto.HexToECDSA(strings.TrimSpace(string(data)))
	}

	return nil, errors.New("failed to read private key from stdin")
}

func SignTx(priv *ecdsa.PrivateKey, nonce uint64, to common.Address, signer types.Signer, amount, gasPrice *big.Int) (*types.Transaction, error) {
	if gasPrice == nil || gasPrice.Sign() <= 0 {
		return nil, fmt.Errorf("invalid gas price")
	}

	if amount == nil || amount.Sign() < 0 {
		return nil, fmt.Errorf("invalid amount")
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    new(big.Int).Set(amount),
		Gas:      21_000,
		GasPrice: new(big.Int).Set(gasPrice),
	})

	signed, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	return signed, nil
}

func Confirm(msg string) (bool, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return false, fmt.Errorf("no TTY available for confirmation: %w", err)
	}
	defer tty.Close()

	fmt.Fprintf(os.Stdout, "%s [y/N]: ", msg)
	reader := bufio.NewReader(tty)
	line, _ := reader.ReadString('\n')
	resp := strings.TrimSpace(strings.ToLower(line))
	return resp == "y" || resp == "yes", nil
}

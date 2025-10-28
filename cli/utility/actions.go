package utility

import (
	"context"
	"fmt"

	"github.com/crypto4people/zoro/cli/misc"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v3"
)

func convertAction(_ context.Context, cmd *cli.Command) error {
	priv, err := misc.LoadPrivateKey()
	if err != nil {
		return err
	}

	fmt.Println("Address: " + crypto.PubkeyToAddress(priv.PublicKey).Hex())
	return nil
}

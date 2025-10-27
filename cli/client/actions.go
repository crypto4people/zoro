package client

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"zoro/misc"
	"zoro/proxy"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v3"
)

func infoAction(ctx context.Context, cmd *cli.Command) error {
	server := cmd.String("server")
	account := cmd.String("account")

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))

	info, err := client.GetInfo(ctx)
	if err != nil {
		return err
	}

	printServerInfo(info)
	return nil
}

func subsAction(ctx context.Context, cmd *cli.Command) error {
	server := cmd.String("server")
	account := cmd.String("account")

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))

	subs, err := client.GetSubs(ctx)
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		fmt.Println("No active subscription found.")
	} else {
		printSubs(subs)
	}
	return nil
}

func connectAction(ctx context.Context, cmd *cli.Command) error {
	priv, err := loadPrivateKey()
	if err != nil {
		return err
	}

	server := cmd.String("server")
	account := crypto.PubkeyToAddress(priv.PublicKey).Hex()

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))

	localSocks := cmd.String("local-socks")

	subs, err := client.GetSubs(ctx)
	if err != nil {
		return err
	}

	if len(subs) == 0 {
		fmt.Println("you got no active sub")
		return nil
	}

	box, err := connect(ctx, client.Endpoint(), client.Account().Hex(), localSocks)
	if err != nil {
		return err
	}

	box.Start()
	<-ctx.Done()
	return ctx.Err()
}

func applyAction(ctx context.Context, cmd *cli.Command) error {
	priv, err := loadPrivateKey()
	if err != nil {
		return err
	}

	server := cmd.String("server")
	account := crypto.PubkeyToAddress(priv.PublicKey)
	client := proxy.NewClient(http.DefaultClient, server, account)

	dur, err := time.ParseDuration(cmd.String("duration"))
	if err != nil {
		return fmt.Errorf("parse duration: %w", err)
	}
	if dur <= 0 {
		return errors.New("duration must be > 0")
	}

	info, err := client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}

	amount := info.Rate.Unapply(dur, dur)
	fmt.Printf("You're about to pay %s ETH for %s of service.\n", misc.WeiToEth(amount), dur)

	confirmed, err := confirm("Do you agree to continue?")
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Println("Transaction cancelled.")
		return nil
	}

	nonce, err := client.GetNonce(ctx)
	if err != nil {
		return fmt.Errorf("get nonce: %w", err)
	}

	signer := types.LatestSignerForChainID(info.Chain.ToBig())

	tx, err := signTx(priv, nonce, info.Account, signer, amount, info.GasPrice.ToBig())
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	if err := client.SubmitTx(ctx, tx); err != nil {
		return fmt.Errorf("submit tx: %w", err)
	}

	renderTable([][]string{
		{"Status", "Submitted"},
		{"Hash", tx.Hash().Hex()},
		{"To", info.Account.Hex()},
		{"Value", fmt.Sprintf("%s ETH", misc.WeiToEth(amount))},
		{"Nonce", fmt.Sprint(nonce)},
		{"GasPrice", fmt.Sprintf("%s wei", info.GasPrice.Dec())},
	})

	return nil
}

func checkAction(ctx context.Context, cmd *cli.Command) error {
	server := cmd.String("server")

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(""))

	hash := cmd.String("hash")
	tx, isPending, err := client.GetReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		return err
	}

	printReceipt(tx, isPending)
	return nil
}

func loadPrivateKey() (*ecdsa.PrivateKey, error) {
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

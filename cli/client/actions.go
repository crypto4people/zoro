package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	cliMisc "github.com/crypto4people/zoro/cli/misc"
	"github.com/crypto4people/zoro/misc"
	"github.com/crypto4people/zoro/proxy"
	"github.com/crypto4people/zoro/singbox"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v3"
)

func infoAction(ctx context.Context, cmd *cli.Command) error {
	server := cmd.String("server")

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress("00"))

	info, err := client.GetInfo(ctx)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(info)
		if err != nil {
			return err
		}
	} else {
		printServerInfo(info)
	}
	return nil
}

func balanceAction(ctx context.Context, cmd *cli.Command) error {
	server := cmd.String("server")
	account := cmd.String("account")

	client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))

	balance, err := client.GetBalance(ctx)
	if err != nil {
		return err
	}

	if cmd.Bool("json") {
		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(struct {
			Balance string `json:"balance"`
		}{balance.Dec()})
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("Balance: %s ETH\n", misc.WeiToEth(balance.ToBig()))
	}
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

	if cmd.Bool("json") {
		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(subs)
		if err != nil {
			return err
		}
	} else {
		printSubs(subs)
	}
	return nil
}

func connectAction(ctx context.Context, cmd *cli.Command) error {
	priv, err := cliMisc.LoadPrivateKey()
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
		return errors.New("no active subscription")
	}

	box, err := singbox.Config(ctx, client.Endpoint(), localSocks, priv, singbox.ModeSocks)
	if err != nil {
		return err
	}

	box.Start()
	<-ctx.Done()
	return ctx.Err()
}

func applyAction(ctx context.Context, cmd *cli.Command) error {
	jsonOut := cmd.Bool("json")
	autoYes := cmd.Bool("yes")
	server := cmd.String("server")

	if strings.TrimSpace(server) == "" {
		return errors.New("server is required (use --server)")
	}

	durStr := cmd.String("duration")
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return fmt.Errorf("invalid --duration %q: %w", durStr, err)
	}

	if dur <= 0 {
		return errors.New("duration must be > 0")
	}

	priv, err := cliMisc.LoadPrivateKey()
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	account := crypto.PubkeyToAddress(priv.PublicKey)

	client := proxy.NewClient(http.DefaultClient, server, account)

	info, err := client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("get info: %w", err)
	}

	amount := info.Rate.Unapply(dur, dur)
	if amount == nil || amount.Sign() <= 0 {
		return fmt.Errorf("calculated amount is not positive for duration %s", dur)
	}

	if !jsonOut && !autoYes {
		fmt.Printf(
			"You're about to pay %s ETH for %s of service (to %s on chain %s).\n",
			misc.WeiToEth(amount), dur, info.Account.Hex(), info.Chain.String(),
		)
		confirmed, err := cliMisc.Confirm("Do you agree to continue?")
		if err != nil {
			return fmt.Errorf("confirm: %w", err)
		}
		if !confirmed {
			return errors.New("transaction cancelled")
		}
	}

	nonce, err := client.GetNonce(ctx)
	if err != nil {
		return fmt.Errorf("get nonce: %w", err)
	}

	signer := types.LatestSignerForChainID(info.Chain.ToBig())
	tx, err := cliMisc.SignTx(priv, nonce, info.Account, signer, amount, info.GasPrice.ToBig())
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	if err := client.SubmitTx(ctx, tx); err != nil {
		return fmt.Errorf("submit tx: %w", err)
	}

	type jsonResponse struct {
		Status   string `json:"status"`
		Hash     string `json:"hash"`
		To       string `json:"to"`
		ValueETH string `json:"value_eth"`
		Nonce    uint64 `json:"nonce"`
		GasPrice string `json:"gas_price_wei"`
		ChainID  string `json:"chain_id"`
		Duration string `json:"duration"`
		Server   string `json:"server"`
		Account  string `json:"account"`
	}

	if jsonOut {
		out := jsonResponse{
			Status:   "submitted",
			Hash:     tx.Hash().Hex(),
			To:       info.Account.Hex(),
			ValueETH: misc.WeiToEth(amount),
			Nonce:    nonce,
			GasPrice: info.GasPrice.Dec(),
			ChainID:  info.Chain.String(),
			Duration: dur.String(),
			Server:   server,
			Account:  account.Hex(),
		}
		enc := json.NewEncoder(os.Stdout)
		if err := enc.Encode(out); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
	} else {
		renderTable([][]string{
			{"Status", "Submitted"},
			{"Hash", tx.Hash().Hex()},
			{"To", info.Account.Hex()},
			{"Value", fmt.Sprintf("%s ETH", misc.WeiToEth(amount))},
			{"Nonce", fmt.Sprint(nonce)},
			{"GasPrice", fmt.Sprintf("%s wei", info.GasPrice.Dec())},
			{"Chain", info.Chain.String()},
			{"Duration", dur.String()},
			{"Server", server},
			{"From (payer)", account.Hex()},
		})
	}

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

	if cmd.Bool("json") {
		receipt := struct {
			Tx        *types.Transaction "json:\"tx\""
			IsPending bool               "json:\"is_pending\""
		}{Tx: tx, IsPending: isPending}

		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode(receipt)
		if err != nil {
			return err
		}
	} else {
		printReceipt(tx, isPending)
	}
	return nil
}

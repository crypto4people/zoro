package client

import (
	"fmt"
	"os"
	"time"

	"github.com/crypto4people/zoro/misc"
	"github.com/crypto4people/zoro/proxy"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/olekukonko/tablewriter"
)

func printServerInfo(info *proxy.Info) {
	renderTable([][]string{
		{"Version", info.Version},
		{"Uptime", (time.Duration(info.Uptime) * time.Second).String()},
		{"Connections", fmt.Sprint(info.Conns)},
		{"Subs", fmt.Sprint(info.Subs)},
		{"Chain", info.Chain.Dec()},
		{"Height", fmt.Sprint(info.Height)},
		{"Contract", info.RateFeed.Hex()},
		{"Account", info.Account.Hex()},
		{"Gas Price", misc.WeiToGwei(info.GasPrice.ToBig()) + " GWEI"},
		{
			"Rate",
			fmt.Sprintf(
				"((value - %s) / %s) * %s + %s",
				info.Rate.Sub.String(),
				info.Rate.Div.String(),
				info.Rate.Mul.String(),
				info.Rate.Add.String(),
			),
		},
	})
}

func printSubs(subs []proxy.Sub) {
	if len(subs) == 0 {
		fmt.Println("No active subscription found.")
		return
	}

	tblContents := [][]string{}

	for idx, sub := range subs {
		tblContents = append(tblContents, []string{"ID", fmt.Sprint(idx)})
		tblContents = append(tblContents, []string{"TX", sub.Tx.Hex()})
		tblContents = append(tblContents, []string{"Remaining", fmt.Sprint(time.Duration(sub.Remain) * time.Second)})
		tblContents = append(tblContents, []string{"Expiration", time.Unix(int64(sub.Expire), 0).String()})
		if idx != len(subs)-1 {
			tblContents = append(tblContents, []string{"", ""})
		}
	}

	renderTable(tblContents)
}

func printReceipt(tx *types.Transaction, isPending bool) {
	if tx == nil {
		fmt.Println("No transaction.")
		return
	}

	status := "Confirmed"
	if isPending {
		status = "Pending"
	}

	renderTable([][]string{
		{"Status", status},
		{"Hash", tx.Hash().Hex()},
		{"To", tx.To().Hex()},
		{"Value", fmt.Sprintf("%s ETH", misc.WeiToEth(tx.Value()))},
		{"Gas Price", fmt.Sprintf("%s GWEI", misc.WeiToGwei(tx.GasPrice()))},
		{"Nonce", fmt.Sprint(tx.Nonce())},
	})
}

func renderTable(data [][]string) {
	tbl := tablewriter.NewWriter(os.Stdout)
	tbl.Bulk(data)
	tbl.Render()
}

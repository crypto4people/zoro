package cli

import (
	"fmt"
	"os"
	"time"
	"zoro/proxy"

	"github.com/olekukonko/tablewriter"
)

func renderTable(name string, header []string, rows [][]string) {
	if name != "" {
		fmt.Println(name)
	}
	t := tablewriter.NewWriter(os.Stdout)
	t.Header(header)
	t.Bulk(rows)
	t.Render()
}

func printServerInfo(info *proxy.Info) {
	renderTable("--- ServerInfo ---", []string{"Version", "Uptime", "Conns", "Subs", "Chain", "Height", "RateFeed"}, [][]string{
		{
			info.Version,
			(time.Duration(info.Uptime) * time.Second).String(),
			fmt.Sprint(info.Conns),
			fmt.Sprint(info.Subs),
			info.Chain,
			fmt.Sprint(info.Height),
			info.RateFeed.Hex(),
		},
	})
	renderTable("", []string{"Account", "Balance", "GasPrice", "Sub", "Div", "Mul", "Add"}, [][]string{
		{
			info.Account.Hex(),
			info.Balance,
			info.GasPrice,
			info.Rate.Sub.String(),
			info.Rate.Div.String(),
			info.Rate.Mul.String(),
			info.Rate.Add.String(),
		},
	})
}

func printSubs(subs []proxy.Sub) {
	if len(subs) == 0 {
		fmt.Println("No subscriptions found.")
		return
	}

	header := []string{"ID", "TX", "Remain", "Expire"}
	data := [][]string{}

	for idx, sub := range subs {
		data = append(data, []string{fmt.Sprint(idx), sub.Tx.Hex(), fmt.Sprint(sub.Remain), fmt.Sprint(sub.Expire)})
	}

	renderTable("--- Subs List ---", header, data)
}

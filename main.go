package main

import (
	"fmt"
	"log"

	"github.com/IsaqueB/ofx-reader/ofx"
)

func main() {
	doc, err := ofx.ParseFile("bb-agosto.ofx")
	if err != nil {
		log.Fatal(err)
	}

	for _, statement := range doc.Statements {
		fmt.Println("Banco:", statement.Account.BankID)
		fmt.Println("Agência:", statement.Account.BranchID)
		fmt.Println("Conta:", statement.Account.AccountID)

		for _, tx := range statement.Transactions {
			fmt.Printf("%s | %s | %s | %s\n",
				tx.PostedAt.Format("2006-01-02"),
				tx.Amount,
				tx.FITID,
				tx.Memo,
			)
		}
	}
}

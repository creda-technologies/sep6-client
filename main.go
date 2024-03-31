package main

import (
	"fmt"

	sep6_client "github.com/creda-technologies/sep6_client/pkg"
	"github.com/gin-gonic/gin"
)

func main() {
	client := defaultClient()

	// info := getInfoWrapper(client)
	webhook_url := "https://e3ba-129-2-192-18.ngrok-free.app/webhook"

	transaction := createDepositWrapper(client, 10000, "LTC", 0.0001, client.Address, &webhook_url, false)
	fmt.Println(transaction.How)
	fmt.Println(transaction.ID)

	// tx := getTransactionWrapper(client, transaction.ID, false)
	// tx := getTransactionWrapper(client, "bb756266155f4bb648cc173b4ef4", false)
	lim := 100
	fmt.Println(len(getTransactionsWrapper(client, "GCZIUNNAH7LIKP7FATNU6DTCI5JB7EPJNRIBO5WYC773AEL5Q53BVGG7", "LTC", &lim, nil, 100, false)))
	app := gin.Default()

	client.RegisterCallbackRoute(app, "/webhook", func(txUpdate sep6_client.Transaction) {
		fmt.Println(txUpdate.Status)
	})

	app.Run(":8000")
}

func getInfoWrapper(client *sep6_client.Sep6Client, print bool) sep6_client.InfoResponse {
	info, err := client.GetInfo()
	if err != nil {
		panic(fmt.Sprintf("GetInfo failed: %v", err))
	}
	if print {
		fmt.Printf("GetInfo: %+v\n", info)
	}
	return *info
}

func getTransactionsWrapper(client *sep6_client.Sep6Client, account string, assetCode string, limit *int, order *string, memo int, print bool) []sep6_client.Transaction {
	transactions, err := client.GetTransactions(account, &memo, assetCode, limit, order)
	if err != nil {
		panic(fmt.Sprintf("GetTransactions failed: %v", err))
	}
	if print {
		fmt.Printf("GetTransactions: %+v\n", transactions)
	}
	return transactions
}

func getTransactionWrapper(client *sep6_client.Sep6Client, transactionID string, print bool) sep6_client.Transaction {
	transaction, err := client.GetTransaction(transactionID)
	if err != nil {
		panic(fmt.Sprintf("GetTransaction failed: %v", err))
	}
	if print {
		fmt.Printf("GetTransaction: %+v\n", transaction)
	}
	return *transaction
}

func createDepositWrapper(client *sep6_client.Sep6Client, memo int, assetCode string, amount float32, account string, onChangeCallback *string, print bool) sep6_client.DepositResponse {
	deposit, err := client.CreateDeposit(assetCode, amount, memo, account, onChangeCallback)
	if err != nil {
		panic(fmt.Sprintf("CreateDeposit failed: %v", err))
	}
	if print {
		fmt.Printf("CreateDeposit: %+v\n", deposit)
	}
	return *deposit
}

func defaultClient() *sep6_client.Sep6Client {
	client, _ := sep6_client.NewSep6Client(
		"SCPCKDXMR3RI74ISS5RKQCN6LJKMKPZOJC6IRAUQWSPXXVXUXD3Y2VBK",
		"https://sep6.whalestack.com/",
		"https://horizon.stellar.org/",
		"e3ba-129-2-192-18.ngrok-free.app",
	)
	return client
}

package sep6-client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/stellar/go/keypair"
)

type Sep6Client struct {
	SecretKey     string
	AnchorUrl     string
	HorizonServer string
	Address       string
	signingKey    string
	HomeDomain    string
}

// Sep6Client struct and error types remain unchanged.

func (c *Sep6Client) GetInfo() (*InfoResponse, error) {
	infoURL := buildURL(c.AnchorUrl, "info", nil)

	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return nil, &FetchError{Field: "URL", Message: "Creating request failed"}
	}

	var infoResp InfoResponse
	if err := sendRequest(req, &infoResp); err != nil {
		return nil, err
	}

	return &infoResp, nil
}
func (c *Sep6Client) GetTransactions(account string, memo *int, assetCode string, limit *int, order *string) ([]Transaction, error) {
	if err := validateNotEmpty(map[string]string{"account": account, "asset_code": assetCode}); err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	if memo != nil {
		queryParams.Add("account", account)
	} else {
		queryParams.Add("account", fmt.Sprintf("%s:%d", account, memo))
	}
	queryParams.Add("asset_code", assetCode)
	if limit != nil {
		queryParams.Add("limit", fmt.Sprintf("%d", *limit))
	} else {
		queryParams.Add("limit", "10")
	}

	if order != nil && (*order == "asc" || *order == "desc") {
		queryParams.Add("order", *order)
	}

	transactionsURL := buildURL(c.AnchorUrl, "transactions", queryParams)
	fmt.Println(transactionsURL)
	req, err := http.NewRequest("GET", transactionsURL, nil)
	if err != nil {
		return nil, &FetchError{Field: "HTTP Request", Message: "Creating request failed"}
	}

	var transactionsResp TransactionsResponse
	if err := sendRequest(req, &transactionsResp); err != nil {
		return nil, err
	}

	return transactionsResp.Transactions, nil
}
func (c *Sep6Client) GetTransaction(transactionID string) (*Transaction, error) {
	if err := validateNotEmpty(map[string]string{"transactionID": transactionID}); err != nil {
		return nil, err
	}

	transactionURL := buildURL(c.AnchorUrl, "transaction", url.Values{"id": []string{transactionID}})

	req, err := http.NewRequest("GET", transactionURL, nil)
	if err != nil {
		return nil, &FetchError{Field: "URL", Message: "Creating request failed"}
	}

	var transactionResp TransactionResponse
	if err := sendRequest(req, &transactionResp); err != nil {
		return nil, err
	}

	return &transactionResp.Transaction, nil
}
func (c *Sep6Client) CreateDeposit(
	assetCode string,
	amount float32,
	memo int,
	account string,
	onChangeCallback *string,
) (*DepositResponse, error) {
	// Validate required parameters using the new helper functions.
	err := validateNotEmpty(map[string]string{
		"assetCode": assetCode,
		"account":   account,
	})
	if err != nil {
		return nil, err
	}

	err = validateAmount(amount)
	if err != nil {
		return nil, err
	}

	// Construct the URL for the /deposit endpoint with query parameters.
	queryParams := url.Values{}
	queryParams.Add("asset_code", assetCode)
	queryParams.Add("amount", fmt.Sprintf("%f", amount))
	queryParams.Add("memo", fmt.Sprintf("%d", memo))
	queryParams.Add("memo_type", "id")
	queryParams.Add("account", account)
	if onChangeCallback != nil {
		queryParams.Add("on_change_callback", *onChangeCallback)
	}

	depositURL := buildURL(c.AnchorUrl, "deposit", queryParams)

	// Create and send the request to the /deposit endpoint.
	req, err := http.NewRequest("GET", depositURL, nil)
	if err != nil {
		return nil, &FetchError{Field: "URL", Message: fmt.Sprintf("Creating request failed: %v", err)}
	}

	var depositResp DepositResponse
	err = sendRequest(req, &depositResp)
	if err != nil {
		return nil, err
	}

	return &depositResp, nil
}

func (c *Sep6Client) RegisterCallbackRoute(server *gin.Engine, path string, handler func(txUpdate Transaction)) error {
	if server == nil {
		return &ValidationError{
			Field:   "Server",
			Message: "no server was provided",
		}
	}

	server.POST(fmt.Sprintf("/%s", path), func(ctx *gin.Context) {
		header := ""
		signatureHeader := ctx.GetHeader("Signature")
		xStellarSignatureHeader := ctx.GetHeader("X-Stellar-Signature")

		if signatureHeader == "" {
			header = xStellarSignatureHeader
		} else {
			header = signatureHeader
		}

		body, bodyErr := io.ReadAll(ctx.Request.Body)
		if bodyErr != nil {
			ctx.JSON(400, gin.H{"error": bodyErr.Error()})
			return
		}

		signatureErr := verifySignatureFromString(
			header,
			string(body),
			c.signingKey,
			c.HomeDomain,
			2,
		)

		if signatureErr != nil {
			ctx.JSON(400, gin.H{"error": signatureErr.Error()})
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))

		var transaction TransactionResponse
		if err := ctx.BindJSON(&transaction); err != nil {
			return
		}

		go handler(transaction.Transaction)

		ctx.JSON(200, gin.H{})
		return
	})
	return nil
}

func NewSep6Client(
	secretKey string,
	anchorUrl string,
	horizonServer string,
	homeDomain string,
) (*Sep6Client, error) {
	account, accParseErr := keypair.ParseFull(secretKey)
	if accParseErr != nil {
		return nil, &ValidationError{
			Field:   "secretKey",
			Message: accParseErr.Error(),
		}
	}

	_, horizonParseErr := url.ParseRequestURI(horizonServer)
	if horizonParseErr != nil {
		return nil, &ValidationError{
			Field:   "horizonServer",
			Message: "invalid Horizon server URL",
		}
	}
	horizonServer = RemoveSlashFromUrl(horizonServer)
	_, anchorParseErr := url.ParseRequestURI(anchorUrl)
	if anchorParseErr != nil {
		return nil, &ValidationError{
			Field:   "anchorUrl",
			Message: "invalid Anchor URL",
		}
	}
	anchorUrl = RemoveSlashFromUrl(anchorUrl)
	signingKey, stellarTomlErr := fetchSigningKey(anchorUrl)
	if stellarTomlErr != nil {
		return nil, &FetchError{
			Field:   "anchorUrl",
			Message: "couldn't fetch Signing Key from stellar.toml",
		}
	}

	return &Sep6Client{
		SecretKey:     secretKey,
		AnchorUrl:     anchorUrl,
		HorizonServer: horizonServer,
		Address:       account.Address(),
		signingKey:    *signingKey,
		HomeDomain:    homeDomain,
	}, nil
}

package sep6_client

type StellarToml struct {
	SigningKey string `toml:"SIGNING_KEY"`
}

type InfoResponse struct {
	Deposit          map[string]AssetDetails    `json:"deposit"`
	Withdraw         map[string]WithdrawDetails `json:"withdraw"`
	Transaction      TransactionDetails         `json:"transaction"`
	Transactions     TransactionsDetails        `json:"transactions"`
	Features         Features                   `json:"features"`
	Fee              FeeDetails                 `json:"fee"`
	DepositExchange  ExchangeDetails            `json:"deposit-exchange"`
	WithdrawExchange ExchangeDetails            `json:"withdraw-exchange"`
	Supply           map[string]SupplyDetails   `json:"supply"`
}

type AssetDetails struct {
	Enabled                bool    `json:"enabled"`
	MinAmount              float64 `json:"min_amount"`
	MaxAmount              float64 `json:"max_amount"`
	FeeFixed               float64 `json:"fee_fixed,omitempty"`
	FeePercent             float64 `json:"fee_percent,omitempty"`
	AuthenticationRequired bool    `json:"authentication_required"`
}

type WithdrawDetails struct {
	AssetDetails
	Types map[string]WithdrawTypeDetails `json:"types"`
}

type WithdrawTypeDetails struct {
	Fields map[string]FieldDetails `json:"fields"`
}

type FieldDetails struct {
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
}

type TransactionDetails struct {
	Enabled                bool `json:"enabled"`
	AuthenticationRequired bool `json:"authentication_required"`
}

// TransactionsDetails mirrors the structure of TransactionDetails since their JSON structure is identical.
type TransactionsDetails TransactionDetails

type Features struct {
	AccountCreation   bool `json:"account_creation"`
	ClaimableBalances bool `json:"claimable_balances"`
}

type FeeDetails struct {
	Enabled                bool `json:"enabled"`
	AuthenticationRequired bool `json:"authentication_required"`
}

type ExchangeDetails struct {
	Enabled                bool `json:"enabled"`
	AuthenticationRequired bool `json:"authentication_required"`
}

type SupplyDetails struct {
	CirculatingSupply           float64          `json:"circulating_supply"`
	CirculatingSupplyComponents SupplyComponents `json:"circulating_supply_components"`
	HotwalletReserves           float64          `json:"hotwallet_reserves"`
	ColdwalletReserves          float64          `json:"coldwallet_reserves"`
	TotalReserves               float64          `json:"total_reserves"`
}

type SupplyComponents struct {
	Amount                  float64 `json:"amount"`
	ClaimableBalancesAmount float64 `json:"claimable_balances_amount"`
	LiquidityPoolsAmount    float64 `json:"liquidity_pools_amount"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type TransactionResponse struct {
	Transaction Transaction `json:"transaction"`
}

type Transaction struct {
	ID           string  `json:"id"`
	Kind         string  `json:"kind"`
	Status       string  `json:"status"`
	MoreInfoURL  string  `json:"more_info_url,omitempty"`
	AmountIn     string  `json:"amount_in,omitempty"`
	AmountOut    string  `json:"amount_out,omitempty"`
	AmountFee    string  `json:"amount_fee,omitempty"`
	StartedAt    string  `json:"started_at,omitempty"`
	CompletedAt  *string `json:"completed_at,omitempty"`
	StellarTxID  *string `json:"stellar_transaction_id,omitempty"`
	ExternalTxID *string `json:"external_transaction_id,omitempty"`
}

type DepositResponse struct {
	How          string                           `json:"how"`
	Instructions map[string]FinancialAccountField `json:"instructions"`
	ID           string                           `json:"id"`
	ETA          int                              `json:"eta"`
	MinAmount    string                           `json:"min_amount"`
	MaxAmount    string                           `json:"max_amount"`
	FeeFixed     string                           `json:"fee_fixed"`
	FeePercent   float64                          `json:"fee_percent"`
	ExtraInfo    map[string]interface{}           `json:"extra_info"`
}

type FinancialAccountField struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

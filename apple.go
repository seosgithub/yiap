package yiap

import (
	"encoding/json"
	"strconv"
	"time"
)

/*
	---------------------------------------------------------------------------
	Definitions
	---------------------------------------------------------------------------
*/

// Each transaction represents one payment from apple.
type AppleTransaction struct {
	Quantity     string `json: "quantity"`
	PurchaseDate string `json:"purchase_date_ms"`
	ExpiredDate  string `json:"expires_date_ms"`
	IsTrial      int    `json:"is_trial"`

	ProductId     string `json:"product_id"`
	TransactionId string `json:"transaction_id"`
}

// Apple will return some information from a receipt
type AppleReceiptResponse struct {
	Status      int    `json:"status"`
	Environment string `json:"environment"`

	LatestReceiptInfo []AppleTransaction `json:"latest_receipt_info"`
	Receipt           struct {
		InApp []AppleTransaction `json:"in_app"`
	}
}

/*
	---------------------------------------------------------------------------
	Constructors
	---------------------------------------------------------------------------
*/

// Returns a set of apple transactions from a response
func NewAppleReceiptResponseFromData(resp []byte) (*AppleReceiptResponse, error) {
	var receipt AppleReceiptResponse

	if err := json.Unmarshal(resp, &receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

/*
	---------------------------------------------------------------------------
	Getters
	---------------------------------------------------------------------------
*/

func (a *AppleReceiptResponse) GetStatus() bool {
	return a.Status == 1
}

func (a *AppleReceiptResponse) EnvironmentIsSandbox() bool {
	return a.Environment == "Sandbox"
}

func (a *AppleReceiptResponse) GetTransactions() []AppleTransaction {
	// Map into unique by transaction ids because `latest_receipt_info` may or
	// may not have duplicates of the `in-app` section.
	resMap := map[string]AppleTransaction{}

	for _, e := range a.LatestReceiptInfo {
		resMap[e.GetTransactionId()] = e
	}

	for _, e := range a.Receipt.InApp {
		resMap[e.GetTransactionId()] = e
	}

	res := []AppleTransaction{}
	for _, e := range resMap {
		res = append(res, e)
	}

	return res
}

func (t *AppleTransaction) GetPurchaseDate() time.Time {
	v, err := strconv.ParseInt(t.PurchaseDate, 10, 64)
	if err != nil {
		return time.Unix(-1, 0)
	}
	return time.Unix(v/100, 0)
}

func (t *AppleTransaction) GetExpiredDate() time.Time {
	v, err := strconv.ParseInt(t.ExpiredDate, 10, 64)
	if err != nil {
		return time.Unix(-1, 0)
	}
	return time.Unix(v/100, 0)
}

func (t *AppleTransaction) GetIsTrial() bool {
	return t.IsTrial == 1
}

func (t *AppleTransaction) GetTransactionId() string {
	return t.TransactionId
}

func (t *AppleTransaction) GetProductId() string {
	return t.ProductId
}

func (t *AppleTransaction) GetQuantity() int64 {
	v, err := strconv.ParseInt(t.Quantity, 10, 64)
	if err != nil {
		return 1
	}
	return v
}

/*
	---------------------------------------------------------------------------
	Setters
	---------------------------------------------------------------------------
*/

/*
	---------------------------------------------------------------------------
	Helpers
	---------------------------------------------------------------------------
*/

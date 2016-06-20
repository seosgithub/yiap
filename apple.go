package yiap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
	PurchaseDate string `json:"original_purchase_date_ms"`
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
func NewAppleReceiptResponseFromData(resp string) (*AppleReceiptResponse, error) {
	var receipt AppleReceiptResponse

	if err := json.Unmarshal([]byte(resp), &receipt); err != nil {
		return nil, err
	}

	return &receipt, nil
}

var specOverrideAppleIAPRequestEndpoint string

// Contacts apples servers and retrieves a receipt. Password is optional
// for non-subscription type receipts
func ProcessAppleIAPRequestPayload(payload string, password string, isProduction bool) (*AppleReceiptResponse, error) {
	payloadStr := strings.TrimSpace(payload) // Payload must not have a newline

	if strings.Contains(payloadStr, "mock_response:") {
		payloadStr = strings.Replace(payloadStr, "mock_response:", "", 1)
		return NewAppleReceiptResponseFromData(payloadStr)
	}

	info := map[string]interface{}{
		"receipt-data": payloadStr,
		"password":     password,
	}

	infoJson, err := json.Marshal(info)
	if err != nil {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload: Tried to marhshal payload '%s'...: %s", payloadStr[0:10], err)
	}

	var url string

	if specOverrideAppleIAPRequestEndpoint == "" {
		if isProduction {
			url = "https://buy.itunes.apple.com/verifyReceipt"
		} else {
			url = "https://sandbox.itunes.apple.com/verifyReceipt"
		}
	} else {
		url = specOverrideAppleIAPRequestEndpoint
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(infoJson))
	if err != nil {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload: Tried to create a request but this failed: %s", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload: Tried to execute request but this failed: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload: Tried to read request buffer but this failed: %s", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload: Tried to read request from apple but got a non-200 status code (got '%d'). Request payload was: '%s', Apple's response payload was: '%s'", payloadStr[:10], resp.StatusCode, body)
	}

	receipt, err := NewAppleReceiptResponseFromData(string(body))
	if err != nil {
		return nil, err
	}

	// Determine if the returned status code is invalid. We don't consider the 21006 case (expired subscription)
	// as the receipt is still processed correctly under those circumstances and it doesn't really make sense.
	iapReceiptCodes := map[int]string{
		21000: "The App Store could not read the JSON object you provided.",
		21002: "The data in the receipt-data property was malformed or missing.",
		21003: "The receipt could not be authenticated.",
		21004: "The shared secret you provided does not match the shared secret on file for your account.",
		21005: "The receipt server is not currently available.",
		21007: "This receipt is from the test environment, but it was sent to the production environment for verification. Send it to the test environment instead.",
		21008: "This receipt is from the production environment, but it was sent to the test environment for verification. Send it to the production environment instead.",
	}
	if codeMsg, ok := iapReceiptCodes[receipt.Status]; ok {
		return nil, fmt.Errorf("ProcessAppleIAPRequestPayload failed to process receipt, it did get a 200 response back from apple but the status code on the receipt was invalid (%d) with request payload: '%s', response payload: '%s'.  Apple's reason for error is: '%s", payloadStr[:10], receipt.Status, body[:100], codeMsg)
	}

	return receipt, nil
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
	return time.Unix(v/1000, 0)
}

func (t *AppleTransaction) GetExpiredDate() time.Time {
	v, err := strconv.ParseInt(t.ExpiredDate, 10, 64)
	if err != nil {
		return time.Unix(-1, 0)
	}
	return time.Unix(v/1000, 0)
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

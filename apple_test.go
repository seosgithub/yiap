package yiap

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestYiap(t *testing.T) {
	Convey("Can decode apple receipt", t, func() {
		responseCode := getFixtureWithPath("apple/receipt0_response.json")

		receipt, err := NewAppleReceiptResponseFromData(responseCode)
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)
		So(len(receipt.GetTransactions()), ShouldEqual, 1)

		txns := receipt.GetTransactions()

		So(txns[0].GetProductId(), ShouldEqual, "com.justalab.cloudplayer.premium")
		So(txns[0].GetTransactionId(), ShouldEqual, "1000000215946012")
		So(txns[0].GetPurchaseDate().Unix(), ShouldEqual, 1465257749)
		So(txns[0].GetExpiredDate().Unix(), ShouldEqual, 1465258049)
		fmt.Printf("%s", txns[0].GetExpiredDate())
		So(txns[0].GetIsTrial(), ShouldEqual, false)
		So(txns[0].GetQuantity(), ShouldEqual, 1)
	})

	Convey("Can decode another apple receipt", t, func() {
		responseCode := getFixtureWithPath("apple/receipt2_response.json")

		receipt, err := NewAppleReceiptResponseFromData(responseCode)
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)
		So(len(receipt.GetTransactions()), ShouldEqual, 1)
	})

	// Sometimes? apple receipts may contain entries in in-app seciton of the receipt which do not
	// co-incide with the subscription results in `latest_receipt_info`.  in this cases, we are
	// going to find all unique transaction items (identified by the transaction_id and then take
	// the union of all the unique transaction items
	Convey("Can decode apple receipt with mixed in-app and updated transactions not from a parent subscrptions", t, func() {
		// This response has been modified to include a in_app transaction where the
		// transaction id is not listed in the latest_receipt_info
		responseCode := getFixtureWithPath("apple/receipt1_response_fake.json")

		receipt, err := NewAppleReceiptResponseFromData(responseCode)
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)

		// 7 Unique transactions (6 in the `latest_receipt_info` and 1 in the `in_app` purchase section.
		So(len(receipt.GetTransactions()), ShouldEqual, 7)
	})

	Convey("Can mock getting receipt from apple by passing mock_response:resp json>", t, func() {
		password := "password"
		respPayload := "mock_response:" + getFixtureWithPath("apple/receipt1_response.json")
		var err error
		receipt, err := ProcessAppleIAPRequestPayload(respPayload, password, false)
		checkErr(err)

		So(len(receipt.GetTransactions()), ShouldEqual, 7)
	})

	Convey("Can request a receipt to be verified & decoded from apple with the latest version", t, func() {
		requestPayload := getFixtureWithPath("apple/receipt2_request")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, getFixtureWithPath("apple/receipt2_response.json"))
		}))
		defer ts.Close()
		url := ts.URL

		// Sends the payload off to apple where it is processed and returned as a
		// JSON receipt
		password := "password"
		specOverrideAppleIAPRequestEndpoint = url
		receipt, err := ProcessAppleIAPRequestPayload(requestPayload, password, false)
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)

		// 1 Unique Transaction as we're returning the response directly
		So(len(receipt.GetTransactions()), ShouldEqual, 1)
		specOverrideAppleIAPRequestEndpoint = ""
	})

	Convey("If apple replies with a non-200x error code, then the error is returned", t, func() {
		requestPayload := getFixtureWithPath("apple/receipt2_request")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// Return a bad request
			http.Error(w, "uh oh", http.StatusBadRequest)
		}))
		defer ts.Close()
		url := ts.URL

		// Send our bogus request
		password := "password"
		specOverrideAppleIAPRequestEndpoint = url
		_, err := ProcessAppleIAPRequestPayload(requestPayload, password, false)
		So(err, ShouldNotEqual, nil)
		So(strings.Contains(fmt.Sprintf("%s", err), "uh oh"), ShouldEqual, true)

		specOverrideAppleIAPRequestEndpoint = ""
	})

	Convey("Does throw an error when contacting apple about an invalid subscription receipt", t, func() {
		requestPayload := getFixtureWithPath("apple/receipt2_request")

		// Forward a bogus password so our request fails
		password := "password"
		_, err := ProcessAppleIAPRequestPayload(requestPayload, password, true)

		// Should fail with a 21007 error (should not be in a production environment)
		So(err, ShouldNotEqual, nil)
		So(strings.Contains(fmt.Sprintf("%s", err), "21007"), ShouldEqual, true)

		_, err = ProcessAppleIAPRequestPayload(requestPayload, password, false)

		// Password is invalid
		So(err, ShouldNotEqual, nil)
		So(strings.Contains(fmt.Sprintf("%s", err), "21004"), ShouldEqual, true)

	})
}

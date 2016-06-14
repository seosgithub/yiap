package yiap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestYiap(t *testing.T) {
	Convey("Can decode apple receipt", t, func() {
		responseCode := getFixtureWithPath("apple/receipt0_response.json")

		receipt, err := NewAppleReceiptResponseFromData([]byte(responseCode))
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)
		So(len(receipt.GetTransactions()), ShouldEqual, 1)

		txns := receipt.GetTransactions()

		So(txns[0].GetProductId(), ShouldEqual, "com.justalab.cloudplayer.premium")
		So(txns[0].GetTransactionId(), ShouldEqual, "1000000215946012")
		So(txns[0].GetPurchaseDate().Unix(), ShouldEqual, 14652577490)
		So(txns[0].GetExpiredDate().Unix(), ShouldEqual, 14652580490)
		So(txns[0].GetIsTrial(), ShouldEqual, false)
	})

	// Sometimes? apple receipts may contain entries in in-app seciton of the receipt which do not
	// co-incide with the subscription results in `latest_receipt_info`.  in this cases, we are
	// going to find all unique transaction items (identified by the transaction_id and then take
	// the union of all the unique transaction items
	Convey("Can decode apple receipt with mixed in-app and updated transactions not from a parent subscrptions", t, func() {
		// This response has been modified to include a in_app transaction where the
		// transaction id is not listed in the latest_receipt_info
		responseCode := getFixtureWithPath("apple/receipt1_response_fake.json")

		receipt, err := NewAppleReceiptResponseFromData([]byte(responseCode))
		checkErr(err)

		So(receipt.GetStatus(), ShouldEqual, false)
		So(receipt.EnvironmentIsSandbox(), ShouldEqual, true)

		// 7 Unique transactions (6 in the `latest_receipt_info` and 1 in the `in_app` purchase section.
		So(len(receipt.GetTransactions()), ShouldEqual, 7)
	})
}

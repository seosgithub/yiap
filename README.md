![YIAP: IAP Helpers](./banner.png) 

[![License](http://img.shields.io/badge/license-MI.T-green.svg?style=flat)](https://github.com/sotownsend/yipyap/blob/master/LICENSE)
[![Build Status](https://circleci.com/gh/sotownsend/yiap.png?circle-token=:circle-token)](https://circleci.com/gh/sotownsend/yiap)

# What is this?

This package provides helper methods for managing IAP with **iOS**® and eventually **Android**®.  Especially designed for auto-renewing subscription systems.

> This has yet to be tested on consumables for iOS, but it should work all the same.

## Apple IAP Specifics


### Overview

In a typical use scenerio, a client (Your **iOS®** app) would have sent your backend a copy of the base64 in-app-purchase receipt.

At this point, it's up to your backend to decode that receipt and verify payments. That's where this package (**yiap**) comes in.

> 🏂 Subscriptions, **including auto-renewal subscriptions**, work in the same fashion as normal store purchases. I.e. you still get back a receipt from the client.  Auto-renewing subscriptions, however, can use the same base64 response to check at a future date whether or not a user auto-renewed or cancelled their subscription.

### How to think about the apple IAP receipt correctly

In a naive implementation, you may try to use the receipt itself as a representation of a transaction.  However, like a real store receipt, a receipt may contain multiple transactions of different product ids (some of which in turn may be auto-renewable, consumable, etc.).  Additionally, there are no guarantees on a receipt having *all* transactions available and you may even receive duplicate transactions on different receipts!  

> 🏂 Apple®'s documentation is not very consistent on what exactly is returned in the receipt if multiple purchase types are made, but from our testing, subscriptions seem to be returned on their own receipts (always).  Therefore, you should be safe using the receipt

So you should consider each receipt as a small snapshot of the total transaction ledger.  E.g. Don't insert receipts into your database, insert transactions into your database and use the transaction-id so that if you get a duplicate, you'll be able to detect this.    **This is especially important when handling restore-purchases, which Apple® often requires per App Store Policy**.

> 🏂 The one corner case where you should store the receipt **request** (base64), is a transaction which represents a subscription.  That way you can verify in a future date whether or not the subscription is valid.

### Example

So here's an example of handling a receipt:

```go
import 'github.com/sotownsend/yiap'

receiptRequest := "<base64 request from client>"
password := "" // Set to your shared-secret if your verifying subscriptions, else blank.

// Also verifies receipt
isProduction := true
receipt, err := yiap.ProcessAppleIAPRequestPayload(receiptRequest, password, isProduction)
checkErr(err)

// Retrieve transactions.  If duplicates of the same transaction appear
// on the receipt, then those duplicates are removed. This includes all
// transactions from `in_app` and `latest_receipt_info` (fetched from
// apple's servers, so it should be up to date)
transactions := receipt.GetTransactions()
for _, tx := range transactions {
  productId := tx.GetProductId()
  transactionId := tx.GetTransactionId()
  
  purchaseDate := tx.GetPurchaseDate()
  expireDate := tx.GetExpiredDate()
  isTrial := tx.GetIsTrial()
  quantity := tx.GetQuantity()

  InsertIntoDBUniq(transactionId, productId, purchaseDate, expiredDate, isTrial, receiptRequest, quantity)
}
```

## Communication
> ♥ This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](http://contributor-covenant.org) code of conduct.

- If you **found a bug**, open an issue.
- If you **have a feature request**, open an issue.
- If you **want to contribute**, submit a pull request.

---

## FAQ

Todo

### Creator

- [Seo Townsend](http://github.com/sotownsend) ([@seotownsend](https://twitter.com/seotownsend))


## License

lzoned is released under the MIT license. See LICENSE for details.

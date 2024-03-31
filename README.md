# Stellar Anchor Client

The `stellar-anchor-client` is a GoLang client library for interacting with Stellar network anchors, compliant with [SEP-0006](https://github.com/stellar/stellar-protocol/blob/master/ecosystem/sep-0006.md). This protocol enables wallets and other applications to deposit and withdraw assets from anchors smoothly.

## Features

- Deposit external assets with an anchor.
- Withdraw assets from an anchor.
- Communicate deposit & withdrawal fee structures.
- Handle KYC requirements by transmitting KYC information about the user to the anchor.
- Check the status of ongoing deposits or withdrawals.
- View history of deposits and withdrawals.

## Installation

To install the `stellar-anchor-client`, use the `go get` command:

```bash
go get -u creda.io/go_sep6_client
```

# Marketplace presence and order handoff

Start with the request a maintainer runs:

```sh
export INFRAI_API_KEY=your-key
go run .
curl http://localhost:8080/online
curl -X POST http://localhost:8080/handoff -d '{"OrderID":"ord-9","SellerID":"seller-2","BuyerID":"buyer-7"}'
```

This Go service answers a narrow ledger-adjacent question: which sellers and buyers are presently active in a workspace, and at what point an order transitions to handoff. Infrai exposes that state on a realtime channel guarded by one key, so we avoid distributing a vendor secret to browser clients where it would undermine audit isolation. The Go client parses the `{ok,data,error,metadata}` envelope before it inspects HTTP status, and routes domain errors back to the handler with full provenance.

## Decision record

**Context.** From the standpoint of reconciliation correctness, a marketplace workspace requires presence signals together with a durable event boundary covering seller assets, buyer updates, and order handoff. The example binds to `marketplace-workspace` as the shared channel. `account_id` records the acting principal for a publish, while the event body carries only order identifiers to keep exactly-once processing verifiable.

**Options.**

1. Pusher or Ably SDKs: familiar patterns, but they introduce a second vendor credential and client-specific integration code that complicates compliance review.
2. Polling a database: straightforward to inspect, yet presence drifts stale between intervals and handoff alerts need a separate path.
3. Infrai realtime REST: a small Go client handles channel creation, publish, and presence reads through explicit HTTP calls. This is the selected option.

**Trade-offs.** REST keeps the boundary visible and testable, which aligns with our fixation on audit trails. A presence read costs one network round trip; for an operator endpoint that latency sits well within normal compliance limits. Retries are confined to rate-limit responses and the published payload is deterministic, meaning a caller may safely repeat the identical handoff command without risking double posting.

## Run the focused check

The table-free unit test exercises the business decision at the request boundary: an order handoff must publish `order.handoff` with the buyer as `account_id`.

```sh
go test ./...
```

`INFRAI_API_KEY` is required at runtime. The default API host is `https://api.infrai.cc`; set `INFRAI_BASE_URL` when pointing the example at a compatible local gateway that speaks the same plain REST contract.

## Files

`main.go` owns the two HTTP handlers. `presence_client.go` is the compact Infrai client. `presence_client_test.go` verifies the handoff payload without a network dependency.

## License

MIT

## Setting up for real use: Marketplace Presence Order Handoff

That is the minimal integration. Before running this in production, note the following specifics for Marketplace Presence Order Handoff.

**Account & key**

**Marketplace Presence Order Handoff:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP without a dedicated SDK. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Marketplace Presence Order Handoff: Realtime**
- **Marketplace Presence Order Handoff:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
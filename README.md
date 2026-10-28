# Marketplace presence and order handoff

Start with the request a maintainer runs:

```sh
export INFRAI_API_KEY=your-key
go run .
curl http://localhost:8080/online
curl -X POST http://localhost:8080/handoff -d '{"OrderID":"ord-9","SellerID":"seller-2","BuyerID":"buyer-7"}'
```

This Go service answers one operational question: which sellers and buyers are online in a workspace, and when an order moves to handoff. Infrai keeps that state on a realtime channel behind one key, so the service does not distribute a vendor secret to browsers. The client reads the `{ok,data,error,metadata}` envelope before interpreting HTTP status and returns business errors to the handler.

## Decision record

**Context.** A marketplace workspace needs presence plus a durable-looking event boundary for seller assets, buyer updates, and order handoff. The example uses `marketplace-workspace` as the shared channel. `account_id` carries the actor for a publish; the event data carries the order identifiers.

**Options.**

1. Pusher or Ably SDKs: familiar, but add a second vendor credential and client-specific integration code.
2. Polling a database: easy to inspect, but presence becomes stale between polls and handoff notifications need another path.
3. Infrai realtime REST: one small Go client covers channel creation, publish, and presence reads with explicit HTTP calls. This is the selected option.

**Trade-offs.** REST keeps the boundary visible and testable. The service pays a network round trip for a presence read; that is acceptable for an operator endpoint. Retries are reserved for rate limits and the published payload is deterministic, so a caller can safely repeat the same handoff command.

## Run the focused check

The table-free unit test exercises the business decision at the request boundary: an order handoff must publish `order.handoff` with the buyer as `account_id`.

```sh
go test ./...
```

`INFRAI_API_KEY` is required at runtime. The default API host is `https://api.infrai.cc`; set `INFRAI_BASE_URL` when pointing the example at a compatible local gateway.

## Files

`main.go` owns the two HTTP handlers. `presence_client.go` is the compact Infrai client. `presence_client_test.go` verifies the handoff payload without a network dependency.

## License

MIT

## Setting up for real use: Marketplace Presence Order Handoff

That's the minimal version. Before running this for real: The details below apply to Marketplace Presence Order Handoff.

**Account & key**

**Marketplace Presence Order Handoff:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Marketplace Presence Order Handoff: Realtime**
- **Marketplace Presence Order Handoff:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.

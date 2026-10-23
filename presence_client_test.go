package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandoffPublishesBuyerIdentity(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{}}`))
	}))
	defer srv.Close()
	c := &PresenceClient{BaseURL: srv.URL, Key: "test", HTTP: srv.Client()}
	err := c.Publish("marketplace-workspace", "order.handoff", "buyer-7", map[string]string{"order_id": "ord-9"})
	if err != nil {
		t.Fatal(err)
	}
	if got["account_id"] != "buyer-7" || got["event"] != "order.handoff" {
		t.Fatalf("unexpected publish: %#v", got)
	}
}

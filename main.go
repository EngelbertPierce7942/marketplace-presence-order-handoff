package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const marketplaceChannel = "marketplace-workspace"

func main() {
	client, err := NewPresenceClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/online", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		data, err := client.Presence(marketplaceChannel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	})
	http.HandleFunc("/handoff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var order struct{ OrderID, SellerID, BuyerID string }
		if json.NewDecoder(r.Body).Decode(&order) != nil || order.OrderID == "" {
			http.Error(w, "order_id required", http.StatusBadRequest)
			return
		}
		err := client.Publish(marketplaceChannel, "order.handoff", order.BuyerID, map[string]string{"order_id": order.OrderID, "seller_id": order.SellerID, "buyer_id": order.BuyerID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

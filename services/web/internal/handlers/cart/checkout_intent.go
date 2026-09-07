package cart

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

const checkoutIntentCookieName = "checkout_intent"

type checkoutIntentStore struct {
	CartID string            `json:"c"`
	Keys   map[string]string `json:"k"`
}

func readCheckoutIntentStore(r *http.Request) checkoutIntentStore {
	store := checkoutIntentStore{Keys: map[string]string{}}
	c, err := r.Cookie(checkoutIntentCookieName)
	if err != nil || c.Value == "" {
		return store
	}
	raw, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		return store
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		return checkoutIntentStore{Keys: map[string]string{}}
	}
	if store.Keys == nil {
		store.Keys = map[string]string{}
	}
	return store
}

func writeCheckoutIntentStore(w http.ResponseWriter, store checkoutIntentStore) {
	if store.CartID == "" || len(store.Keys) == 0 {
		http.SetCookie(w, &http.Cookie{Name: checkoutIntentCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
		return
	}
	raw, err := json.Marshal(store)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     checkoutIntentCookieName,
		Value:    base64.RawURLEncoding.EncodeToString(raw),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func intentForMerchant(store *checkoutIntentStore, cartID, merchantID string) string {
	if store.CartID != cartID {
		store.CartID = cartID
		store.Keys = map[string]string{}
	}
	if key := store.Keys[merchantID]; key != "" {
		return key
	}
	key := uuid.NewString()
	store.Keys[merchantID] = key
	return key
}

func rotateCheckoutIntent(w http.ResponseWriter, r *http.Request, cartID, merchantID string) {
	if cartID == "" || merchantID == "" {
		return
	}
	store := readCheckoutIntentStore(r)
	if store.CartID != cartID {
		store.CartID = cartID
		store.Keys = map[string]string{}
	}
	store.Keys[merchantID] = uuid.NewString()
	writeCheckoutIntentStore(w, store)
}

func clearCheckoutIntent(w http.ResponseWriter, r *http.Request, cartID, merchantID string) {
	if cartID == "" || merchantID == "" {
		return
	}
	store := readCheckoutIntentStore(r)
	if store.CartID != cartID {
		writeCheckoutIntentStore(w, checkoutIntentStore{})
		return
	}
	delete(store.Keys, merchantID)
	writeCheckoutIntentStore(w, store)
}

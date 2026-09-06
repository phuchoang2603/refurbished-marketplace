package cart

import (
	"net/http"
	"strings"

	shared "github.com/phuchoang2603/refurbished-marketplace/services/web/internal/handlers/shared"
)

func DrainPaidProductIDs(w http.ResponseWriter, r *http.Request, cartSvc shared.CartService, productIDs []string) {
	if cartSvc == nil || len(productIDs) == 0 {
		return
	}
	cartID := ""
	if c, err := r.Cookie(cartCookieName); err == nil {
		cartID = strings.TrimSpace(c.Value)
	}
	if cartID == "" {
		return
	}
	updated, err := cartSvc.RemoveCartItems(r.Context(), cartID, productIDs)
	if err != nil {
		return
	}
	if updated != nil && len(updated.GetItems()) == 0 {
		http.SetCookie(w, &http.Cookie{Name: cartCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1, SameSite: http.SameSiteLaxMode})
	}
}

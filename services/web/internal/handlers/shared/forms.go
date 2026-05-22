package shared

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

func RefreshTokenFromForm(r *http.Request) (string, error) {
	if !parseForm(r) {
		return "", ErrInvalidRequestBody
	}
	return r.FormValue("refresh_token"), nil
}

func EmailPasswordFromForm(r *http.Request) (string, string, error) {
	if !parseForm(r) {
		return "", "", ErrInvalidRequestBody
	}
	return r.FormValue("email"), r.FormValue("password"), nil
}

func ProductQuantityFromForm(r *http.Request) (string, int32, error) {
	if !parseForm(r) {
		return "", 0, ErrInvalidRequestBody
	}
	quantity, err := parseInt32FormValue(r, "quantity")
	if err != nil {
		return "", 0, err
	}
	return r.FormValue("product_id"), quantity, nil
}

func ProductQuantityMerchantFromForm(r *http.Request) (string, string, int32, error) {
	if !parseForm(r) {
		return "", "", 0, ErrInvalidRequestBody
	}
	quantity, err := parseInt32FormValue(r, "quantity")
	if err != nil {
		return "", "", 0, err
	}
	return r.FormValue("product_id"), r.FormValue("merchant_id"), quantity, nil
}

func ProductCreateFromForm(r *http.Request) (string, string, int64, int32, error) {
	if !parseForm(r) {
		return "", "", 0, 0, ErrInvalidRequestBody
	}
	priceCents, err := parsePriceDollarsToCents(r.FormValue("price"))
	if err != nil {
		return "", "", 0, 0, ErrInvalidRequestBody
	}
	initialStock, err := parseInt32FormValue(r, "initial_stock")
	if err != nil {
		return "", "", 0, 0, err
	}
	return r.FormValue("name"), r.FormValue("description"), priceCents, initialStock, nil
}

func MerchantIDFromForm(r *http.Request) (string, error) {
	if !parseForm(r) {
		return "", ErrInvalidRequestBody
	}
	merchantID := strings.TrimSpace(r.FormValue("merchant_id"))
	if merchantID == "" {
		return "", ErrInvalidRequestBody
	}
	return merchantID, nil
}

func parseForm(r *http.Request) bool {
	return r.ParseForm() == nil
}

func parseInt32FormValue(r *http.Request, key string) (int32, error) {
	value, err := strconv.ParseInt(r.FormValue(key), 10, 32)
	if err != nil {
		return 0, ErrInvalidRequestBody
	}
	return int32(value), nil
}

func parsePriceDollarsToCents(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, ErrInvalidRequestBody
	}
	parts := strings.Split(raw, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalidRequestBody
	}
	dollars, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || dollars < 0 {
		return 0, ErrInvalidRequestBody
	}
	cents := int64(0)
	if len(parts) == 2 {
		fraction := parts[1]
		if len(fraction) == 0 || len(fraction) > 2 {
			return 0, ErrInvalidRequestBody
		}
		if len(fraction) == 1 {
			fraction += "0"
		}
		parsedCents, err := strconv.ParseInt(fraction, 10, 64)
		if err != nil {
			return 0, ErrInvalidRequestBody
		}
		cents = parsedCents
	}
	return dollars*100 + cents, nil
}

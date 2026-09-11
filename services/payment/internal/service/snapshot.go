package service

import (
	"encoding/json"
	"strings"
)

func usableShippingAddress(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	return strings.TrimSpace(m["line1"]) != "" &&
		strings.TrimSpace(m["city"]) != "" &&
		strings.TrimSpace(m["postal_code"]) != "" &&
		strings.TrimSpace(m["country"]) != ""
}

func partySnapshotJSON(id, email string) (json.RawMessage, error) {
	m := map[string]string{"id": strings.TrimSpace(id)}
	if v := strings.TrimSpace(email); v != "" {
		m["email"] = v
	}
	return json.Marshal(m)
}

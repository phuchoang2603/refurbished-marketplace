package main

import "testing"

func TestNormalizeCallbackURLUpgradesPublicHTTP(t *testing.T) {
	got := normalizeCallbackURL("http://shop.phuchoang.sbs/callbacks/hosted-payment")
	want := "https://shop.phuchoang.sbs/callbacks/hosted-payment"
	if got != want {
		t.Fatalf("normalizeCallbackURL = %q, want %q", got, want)
	}
}

func TestNormalizeCallbackURLKeepsClusterHTTP(t *testing.T) {
	got := normalizeCallbackURL("http://web:8080/callbacks/hosted-payment")
	want := "http://web:8080/callbacks/hosted-payment"
	if got != want {
		t.Fatalf("normalizeCallbackURL = %q, want %q", got, want)
	}
}

func TestNormalizeCallbackURLRewritesLocalhost(t *testing.T) {
	got := normalizeCallbackURL("http://localhost:8080/callbacks/hosted-payment")
	want := "http://host.docker.internal:8080/callbacks/hosted-payment"
	if got != want {
		t.Fatalf("normalizeCallbackURL = %q, want %q", got, want)
	}
}

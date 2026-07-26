/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 * Licensed under the Apache License, Version 2.0.
 */

package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wso2/openfgc/portal/backend/internal/system/config"
)

func TestSplitTokenEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		maxPart int
		want1   string
		want2   string
		wantErr bool
	}{
		{name: "even", token: "abcdef", maxPart: 3, want1: "abc", want2: "def"},
		{name: "odd midpoint", token: "abcde", maxPart: 3, want1: "ab", want2: "cde"},
		{name: "empty", token: "", maxPart: 3, wantErr: true},
		{name: "one byte", token: "a", maxPart: 1, wantErr: true},
		{name: "part too large", token: "abcdefg", maxPart: 3, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			part1, part2, err := splitToken(test.token, test.maxPart)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if part1 != test.want1 || part2 != test.want2 {
				t.Fatalf("unexpected parts: %q %q", part1, part2)
			}
		})
	}
}

func TestReconstructTokenLimitsAndOrder(t *testing.T) {
	cfg := config.AuthConfig{MaxTokenPartBytes: 4, MaxReconstructedTokenBytes: 7}
	if got, err := reconstructToken("abc", "defg", cfg); err != nil || got != "abcdefg" {
		t.Fatalf("unexpected reconstruction: %q, %v", got, err)
	}
	if got, err := reconstructToken("defg", "abc", cfg); err != nil || got != "defgabc" {
		t.Fatalf("reconstruction must preserve supplied order: %q, %v", got, err)
	}
	for _, parts := range [][2]string{{"", "a"}, {"a", ""}, {"abcde", "a"}, {"abcd", "abcd"}} {
		if _, err := reconstructToken(parts[0], parts[1], cfg); err == nil {
			t.Fatalf("expected reconstruction to reject %#v", parts)
		}
	}
}

func TestBearerPartStrictParsing(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    string
		wantErr bool
	}{
		{name: "valid", values: []string{"Bearer abc"}, want: "abc"},
		{name: "case insensitive scheme", values: []string{"bearer abc"}, want: "abc"},
		{name: "missing", wantErr: true},
		{name: "empty", values: []string{""}, wantErr: true},
		{name: "scheme only", values: []string{"Bearer"}, wantErr: true},
		{name: "wrong scheme", values: []string{"Basic abc"}, wantErr: true},
		{name: "extra field", values: []string{"Bearer abc def"}, wantErr: true},
		{name: "combined values", values: []string{"Bearer abc, Bearer def"}, wantErr: true},
		{name: "duplicate headers", values: []string{"Bearer abc", "Bearer def"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			for _, value := range test.values {
				r.Header.Add("Authorization", value)
			}
			got, err := bearerPart(r)
			if (err != nil) != test.wantErr || got != test.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestExactlyOneCookie(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    string
		wantErr bool
	}{
		{name: "valid", header: "other=x; token=part2", want: "part2"},
		{name: "missing", header: "other=x", wantErr: true},
		{name: "empty", header: "token=", wantErr: true},
		{name: "duplicate", header: "token=a; token=b", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("Cookie", test.header)
			got, err := exactlyOneCookie(r, "token")
			if (err != nil) != test.wantErr || got != test.want {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestCookieIssuanceAndExactClearing(t *testing.T) {
	cfg := config.AuthConfig{
		CookieSecure: true, CookieSameSite: "Strict", MaxTokenPartBytes: 32,
		AccessTokenPart1Cookie: "at1", AccessTokenPart2Cookie: "at2",
		RefreshTokenPart1Cookie: "rt1", RefreshTokenPart2Cookie: "rt2",
		IDTokenPart1Cookie: "id1", IDTokenPart2Cookie: "id2",
	}
	manager := &Manager{cfg: cfg}

	issued := httptest.NewRecorder()
	if err := manager.setSplitCookies(issued, "abcdefgh", "at1", "at2", true, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	cookies := issued.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected two cookies, got %d", len(cookies))
	}
	for index, cookie := range cookies {
		if cookie.Path != "/" || cookie.Domain != "" || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode || cookie.MaxAge < 1 {
			t.Fatalf("unexpected issued cookie: %#v", cookie)
		}
		if cookie.HttpOnly != (index == 1) {
			t.Fatalf("unexpected HttpOnly value for %s", cookie.Name)
		}
	}

	cleared := httptest.NewRecorder()
	manager.clearTokenCookies(cleared)
	clearCookies := cleared.Result().Cookies()
	if len(clearCookies) != 6 {
		t.Fatalf("expected six clearing cookies, got %d", len(clearCookies))
	}
	for _, cookie := range clearCookies {
		if cookie.Value != "" || cookie.Path != "/" || cookie.Domain != "" || cookie.MaxAge != -1 || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
			t.Fatalf("unexpected clearing cookie: %#v", cookie)
		}
		wantHTTPOnly := cookie.Name == "at2" || cookie.Name == "rt2"
		if cookie.HttpOnly != wantHTTPOnly {
			t.Fatalf("unexpected clearing HttpOnly value for %s", cookie.Name)
		}
	}
}

func TestParseSameSite(t *testing.T) {
	tests := map[string]http.SameSite{
		"Lax": http.SameSiteLaxMode, "unknown": http.SameSiteLaxMode,
		"Strict": http.SameSiteStrictMode, "None": http.SameSiteNoneMode,
	}
	for input, want := range tests {
		if got := parseSameSite(input); got != want {
			t.Errorf("%q: got %v, want %v", input, got, want)
		}
	}
}

func TestCompleteTokenInjectionCannotReplaceMissingPart(t *testing.T) {
	manager := &Manager{cfg: config.AuthConfig{
		Enabled: true, AccessTokenPart2Cookie: "at2", MaxTokenPartBytes: 100,
		MaxReconstructedTokenBytes: 200,
	}}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+strings.Repeat("x", 20))
	if _, _, err := manager.authenticate(r); err == nil {
		t.Fatal("a complete bearer token without the matching cookie half must be rejected")
	}
}

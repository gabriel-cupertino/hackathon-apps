package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler_Returns200(t *testing.T) {
	app := &App{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	app.HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestHealthHandler_ResponseBody(t *testing.T) {
	app := &App{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	app.HealthHandler(rr, req)

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
	if resp["service"] != "donation-service" {
		t.Errorf("expected service=donation-service, got %q", resp["service"])
	}
}

func TestDonationHandler_MethodNotAllowed(t *testing.T) {
	app := &App{}

	for _, method := range []string{http.MethodPut, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/donations", nil)
		rr := httptest.NewRecorder()

		app.DonationHandler(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("[%s] expected 405, got %d", method, rr.Code)
		}
	}
}

func TestDonationHandler_InvalidJSON(t *testing.T) {
	app := &App{}

	req := httptest.NewRequest(http.MethodPost, "/donations", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.DonationHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestDonationStruct_JSONRoundtrip(t *testing.T) {
	d := Donation{
		ID:        1,
		NgoID:     42,
		Amount:    150.75,
		DonorName: "Maria Silva",
		Status:    "APPROVED",
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Donation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.NgoID != d.NgoID || decoded.Amount != d.Amount || decoded.DonorName != d.DonorName {
		t.Errorf("roundtrip mismatch: got %+v, want %+v", decoded, d)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	app := &App{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	app.HealthHandler(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}

package http_handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestParamsPathIDWinsOverQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/gateway/gw1?id=gw2", nil)
	r = mux.SetURLVars(r, map[string]string{"id": "gw1"})

	filters, _, err := Params(r)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	for _, f := range filters {
		if f.Key == "id" {
			id = f.Value.(string)
		}
	}
	if id != "gw1" {
		t.Fatalf("path id should win, got %q", id)
	}
}

func TestParamsQueryUsedWhenNoPathVar(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/gateway?id=gw2", nil)
	filters, _, err := Params(r)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	for _, f := range filters {
		if f.Key == "id" {
			id = f.Value.(string)
		}
	}
	if id != "gw2" {
		t.Fatalf("query id should apply on collection, got %q", id)
	}
}

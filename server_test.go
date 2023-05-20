package main

/*
Testing server's HTTP endpoints
*/

import (
	"testing"
    "net/http/httptest"
    "net/http"
)

func TestGetProduct(t *testing.T) {
    req, err := http.NewRequest("GET", "/product", nil)
    if err != nil {
        t.Fatal(err)
    }
    rr := httptest.NewRecorder()
    p := newProductHandler()
    handler := http.HandlerFunc(p.get)
    handler.ServeHTTP(rr, req)
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v",
            status, http.StatusOK)
    }
    expected := `[{"name":"Shoes","price":25},{"name":"Webcam","price":50},{"name":"Mic","price":20}]`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v",
            rr.Body.String(), expected)
    }
}


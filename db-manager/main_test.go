package main

import (
	"log"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	log.Println("Setup: Add any setup below...")

	// Run all tests
	code := m.Run()

	log.Println("Teardown: Add any teardown below...")
	os.Exit(code)
}

func TestHttpGet(t *testing.T) {
	log.Println("testing http get!")
	// t.Error("This fails TestHttpGet() but continues execution")
	// t.Fatal("This fails TestHttpGet() and halts execution")

	var resp *http.Response
	var err error
	resp, err = http.Get("http://localhost:8080/beans")
	if err != nil {
		t.Error("Error performing Get(). Is the server running?", err)
	}
	const expectedStatusCode = 200
	if resp.StatusCode != expectedStatusCode {
		t.Errorf("GET responded with status %s, code %d (expected %d)", resp.Status, resp.StatusCode, expectedStatusCode)
	}
}

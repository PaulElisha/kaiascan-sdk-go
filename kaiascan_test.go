package kaiascan

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfigureSDK(t *testing.T) {
	ConfigureSDK(true)
	if BASE_URL != BASE_URL_TESTNET || CHAIN_ID != CHAIN_ID_TESTNET {
		t.Errorf("Testnet configuration failed. Got BASE_URL: %s, CHAIN_ID: %s", BASE_URL, CHAIN_ID)
	}

	ConfigureSDK(false)
	if BASE_URL != BASE_URL_MAINNET || CHAIN_ID != CHAIN_ID_MAINNET {
		t.Errorf("Mainnet configuration failed. Got BASE_URL: %s, CHAIN_ID: %s", BASE_URL, CHAIN_ID)
	}
}

func mockApiResponse[T any](data T, code int, msg string) []byte {
	response := ApiResponse[T]{Code: code, Data: data, Msg: msg}
	jsonResponse, _ := json.Marshal(response)
	return jsonResponse
}

func TestGetFungibleToken(t *testing.T) {
	ConfigureSDK(true)
	mockToken := TokenInfo{
		ContractType:   "ERC20",
		Name:           "TestToken",
		Symbol:         "TT",
		Decimal:        18,
		TotalSupply:    1000000,
		TotalTransfers: 500,
		OfficialSite:   "https://example.com",
		BurnAmount:     1000,
		TotalBurns:     10,
	}
	mockResponse := mockApiResponse(mockToken, 0, "Success")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Mock server received request: %s", r.URL.String())
		if r.URL.Path != "/api/v1/tokens" {
			t.Fatalf("Unexpected API path: %s", r.URL.Path)
		}
		w.Write(mockResponse)
	}))
	defer server.Close()

	BASE_URL = server.URL + "/"

	resp, err := GetFungibleToken("0x1234567890abcdef")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Code != 0 {
		t.Fatalf("Expected response code 0, got %d", resp.Code)
	}
	if resp.Data.Name != "TestToken" {
		t.Errorf("Expected token name 'TestToken', got %s", resp.Data.Name)
	}
	log.Printf("Response: %+v", resp.Data)
}

func TestGetFungibleToken_Error(t *testing.T) {
	mockResponse := mockApiResponse(TokenInfo{}, 1, "Invalid token address")

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockResponse)
	}))
	defer server.Close()

	BASE_URL = server.URL + "/"

	resp, err := GetFungibleToken("0xInvalidAddress")
	if err == nil {
		t.Fatal("Expected an error but got none")
	}
	if resp != nil && resp.Code == 0 {
		t.Fatalf("Expected non-zero response code, got %d", resp.Code)
	}
	log.Printf("Error Response: %v", err)
}

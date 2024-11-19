package kaiascan

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockApiResponse[T any](data T, code int, msg string) []byte {
	response := ApiResponse[T]{Code: code, Data: data, Msg: msg}
	jsonResponse, _ := json.Marshal(response)
	return jsonResponse
}

func TestSetEnvironment(t *testing.T) {
	SetEnvironment(true)
	if BASE_URL != "https://kairos-oapi.kaiascan.io/" || CHAIN_ID != "1001" {
		t.Fatalf("Testnet environment not set correctly")
	}

	SetEnvironment(false)
	if BASE_URL != "https://mainnet-oapi.kaiascan.io/" || CHAIN_ID != "8217" {
		t.Fatalf("Mainnet environment not set correctly")
	}

	log.Println("Environment switching test passed.")
}

func TestGetFungibleToken(t *testing.T) {
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

func TestEnvironmentIntegration(t *testing.T) {
	SetEnvironment(true)
	if BASE_URL != "https://kairos-oapi.kaiascan.io/" {
		t.Fatalf("Failed to switch to Testnet environment")
	}

	SetEnvironment(false)
	if BASE_URL != "https://mainnet-oapi.kaiascan.io/" {
		t.Fatalf("Failed to switch to Mainnet environment")
	}
	log.Println("Environment integration test passed.")
}

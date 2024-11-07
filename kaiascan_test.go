package oklink

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockApiResponse[T any](data T, code int, msg string) []byte {
	response := ApiResponse[T]{Code: code, Data: data, Msg: msg}
	jsonResponse, _ := json.Marshal(response)
	return jsonResponse
}

func TestGetFungibleToken(t *testing.T) {
	mockToken := TokenInfo{
		ContractType:   "ERC20",
		Name:           "TestToken",
		Symbol:         "TT",
		Icon:           "https://example.com/icon.png",
		Decimal:        18,
		TotalSupply:    1000000,
		TotalTransfers: 500,
		OfficialSite:   "https://example.com",
		BurnAmount:     1000,
		TotalBurns:     10,
	}
	mockResponse := mockApiResponse(mockToken, 0, "Success")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/tokens" {
			t.Fatalf("Unexpected API path: %s", r.URL.Path)
		}
		w.Write(mockResponse)
	}))
	defer server.Close()

	// Override the BASE_URL with the test server URL
	BASE_URL = server.URL + "/"

	// Call the function
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
}

func TestGetFungibleToken_ApiError(t *testing.T) {
	mockResponse := mockApiResponse(TokenInfo{}, 1, "Invalid token address")

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
}

func TestGetNftItem(t *testing.T) {
	// Mock server
	mockNft := map[string]interface{}{"nftName": "TestNFT"}
	mockResponse := mockApiResponse(mockNft, 0, "Success")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nfts" {
			t.Fatalf("Unexpected API path: %s", r.URL.Path)
		}
		w.Write(mockResponse)
	}))
	defer server.Close()

	BASE_URL = server.URL + "/"

	resp, err := GetNftItem("0x1234567890abcdef", "1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("Expected response code 0, got %d", resp.Code)
	}
	if resp.Data.(map[string]interface{})["nftName"] != "TestNFT" {
		t.Errorf("Expected NFT name 'TestNFT', got %v", resp.Data.(map[string]interface{})["nftName"])
	}
}

func TestGetContractCreationCode(t *testing.T) {
	mockData := map[string]interface{}{"creationCode": "0x600160005260206000f3"}
	mockResponse := mockApiResponse(mockData, 0, "Success")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/contracts/creation-code" {
			t.Fatalf("Unexpected API path: %s", r.URL.Path)
		}
		w.Write(mockResponse)
	}))
	defer server.Close()

	BASE_URL = server.URL + "/"

	resp, err := GetContractCreationCode("0xContractAddress")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("Expected response code 0, got %d", resp.Code)
	}
	if resp.Data.(map[string]interface{})["creationCode"] != "0x600160005260206000f3" {
		t.Errorf("Expected creation code '0x600160005260206000f3', got %v", resp.Data.(map[string]interface{})["creationCode"])
	}
}

func TestFetchApi_HttpError(t *testing.T) {
	BASE_URL = "http://invalid-url"

	_, err := GetFungibleToken("0x1234567890abcdef")
	if err == nil {
		t.Fatal("Expected an HTTP error but got none")
	}
}

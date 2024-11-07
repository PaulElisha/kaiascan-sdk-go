package oklink

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockFungibleTokenResponse() *ApiResponse[map[string]interface{}] {
	return &ApiResponse[map[string]interface{}]{
		Code: 0,
		Msg:  "success",
		Data: map[string]interface{}{
			"contractType":   "ERC20",
			"name":           "Mock Token",
			"symbol":         "MTK",
			"icon":           "https://mockToken.com/icon.png",
			"decimal":        18,
			"totalSupply":    float64(1000000), // Explicitly use float64 for numerical consistency
			"totalTransfers": float64(500),
			"officialSite":   "https://mck.com",
			"burnAmount":     float64(100),
			"totalBurns":     float64(5),
		},
	}
}

func setupMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestGetFungibleToken(t *testing.T) {
	ts := setupMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := mockFungibleTokenResponse()
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}

		oldBaseURL := BASE_URL
		BASE_URL = ts.URL + "/"
		defer func() { BASE_URL = oldBaseURL }()

		tokenAddress := Address("0x1234567890abcdef")
		response, err := GetFungibleToken(tokenAddress)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if response.Code != 0 {
			t.Errorf("Expected code 0, got %d", response.Code)
		}
		if response.Data["name"] != "Mock Token" {
			t.Errorf("Expected token name 'Mock Token', got %s", response.Data["name"])
		}
		if response.Data["symbol"] != "MTK" {
			t.Errorf("Expected token symbol 'MTK', got %s", response.Data["symbol"])
		}
		if response.Data["totalSupply"] != float64(1000000) {
			t.Errorf("Expected total supply 1000000, got %v", response.Data["totalSupply"])
		}
		if response.Data["totalTransfers"] != float64(500) {
			t.Errorf("Expected total transfers 500, got %v", response.Data["totalTransfers"])
		}

	})
	defer ts.Close()
}

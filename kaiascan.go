package oklink

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	BASE_URL string = "https://mainnet-oapi.kaiascan.io/"
	CHAIN_ID string = "8217"
)

type Address string

type ApiResponse[T any] struct {
	Code int    `json:"code"`
	Data T      `json:"data"`
	Msg  string `json:"msg"`
}

func main() {
	tokenInfo, err := GetFungibleToken("0x1234567890abcdef")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Fungible Token Information: %+v\n", tokenInfo)
}

func fetchApi[T any](urlStr string) (*ApiResponse[T], error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error! status: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var apiResponse ApiResponse[T]
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}
	if apiResponse.Code != 0 {
		return nil, fmt.Errorf("API error! code: %d, message: %s", apiResponse.Code, apiResponse.Msg)
	}
	return &apiResponse, nil
}

func GetFungibleToken(tokenAddress Address) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("tokenAddress", string(tokenAddress))

	urlStr := fmt.Sprintf("%sapi/v1/tokens?%s", BASE_URL, params.Encode())
	return fetchApi[any](urlStr)
}

func GetNftItem(nftAddress Address, tokenId string) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("nftAddress", string(nftAddress))
	params.Add("tokenId", tokenId)

	urlStr := fmt.Sprintf("%sapi/v1/nfts?%s&tokenId=%s", BASE_URL, params.Encode(), tokenId)
	return fetchApi[any](urlStr)
}

func GetContractCreationCode(contractAddress Address) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("contractAddress", string(contractAddress))

	urlStr := fmt.Sprintf("%sapi/v1/contracts/creation-code?%s", BASE_URL, params.Encode())
	return fetchApi[any](urlStr)
}

func GetLatestBlock() (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%sapi/v1/blocks/latest", BASE_URL)
	return fetchApi[any](urlStr)
}

func GetBlock(blockNumber int64) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("blockNumber", fmt.Sprintf("%d", blockNumber))

	urlStr := fmt.Sprintf("%sapi/v1/blocks?%s", BASE_URL, params.Encode())
	return fetchApi[any](urlStr)
}

func GetBlocks() (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%sapi/v1/blocks", BASE_URL)
	return fetchApi[any](urlStr)
}

func GetTransactionsOfBlock(blockNumber int64) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("blockNumber", fmt.Sprintf("%d", blockNumber))

	urlStr := fmt.Sprintf("%sapi/v1/blocks/%d/transactions", BASE_URL, blockNumber)
	return fetchApi[any](urlStr)
}

func GetTransactionReceiptStatus(transactionHash string) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("transactionHash", transactionHash)

	urlStr := fmt.Sprintf("%sapi/v1/transaction-receipts/status?%s", BASE_URL, params.Encode())
	return fetchApi[any](urlStr)
}

func GetTransaction(transactionHash string) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("transactionHash", transactionHash)

	urlStr := fmt.Sprintf("%sapi/v1/transactions/%s", BASE_URL, transactionHash)
	return fetchApi[any](urlStr)
}

func GetContractSourceCode(contractAddress Address) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("contractAddress", string(contractAddress))

	urlStr := fmt.Sprintf("%sapi/v1/contracts/source-code?%s", BASE_URL, params.Encode())
	return fetchApi[any](urlStr)
}

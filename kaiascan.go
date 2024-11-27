package kaiascan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	BASE_URL = "https://mainnet-oapi.kaiascan.io/"
	CHAIN_ID = "8217"

	tokensEndpoint      = "api/v1/tokens"
	nftsEndpoint        = "api/v1/nfts"
	blocksEndpoint      = "api/v1/blocks"
	transactionEndpoint = "api/v1/transactions"
	contractEndpoint    = "api/v1/contracts"
	transactionReceipts = "api/v1/transaction-receipts"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type Address = string

type ApiResponse[T any] struct {
	Code int    `json:"code"`
	Data T      `json:"data"`
	Msg  string `json:"msg"`
}

type TokenInfo struct {
	ContractType   string  `json:"contractType"`
	Name           string  `json:"name"`
	Symbol         string  `json:"symbol"`
	Icon           string  `json:"icon"`
	Decimal        int32   `json:"decimal"`
	TotalSupply    float64 `json:"totalSupply"`
	TotalTransfers int64   `json:"totalTransfers"`
	OfficialSite   string  `json:"officialSite"`
	BurnAmount     float64 `json:"burnAmount"`
	TotalBurns     int64   `json:"totalBurns"`
}

func fetchApi[T any](urlStr string) (*ApiResponse[T], error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for %s: %w", urlStr, err)
	}
	req.Header.Add("Content-Type", "application/json")

	response, err := httpClient.Do(req)
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

func GetAccountKeyHistories(accountAddress string, page int, size int) (*ApiResponse[any], error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	encodedAddress := url.PathEscape(accountAddress)

	urlStr := fmt.Sprintf("https://mainnet-oapi.kaiascan.io/api/v1/accounts/%s/key-histories?%s", encodedAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetFungibleToken(tokenAddress Address) (*ApiResponse[TokenInfo], error) {
	params := url.Values{}
	params.Add("tokenAddress", string(tokenAddress))

	urlStr := fmt.Sprintf("%s%s?%s", BASE_URL, tokensEndpoint, params.Encode())
	return fetchApi[TokenInfo](urlStr)
}

func GetNftItem(nftAddress Address, tokenId string) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("nftAddress", string(nftAddress))
	params.Add("tokenId", tokenId)

	urlStr := fmt.Sprintf("%s%s?%s", BASE_URL, nftsEndpoint, params.Encode())
	return fetchApi[any](urlStr)
}

func GetContractCreationCode(contractAddress Address) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("contractAddress", string(contractAddress))

	urlStr := fmt.Sprintf("%s%s/creation-code?%s", BASE_URL, contractEndpoint, params.Encode())
	return fetchApi[any](urlStr)
}

func GetContractSourceCode(contractAddress Address) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("contractAddress", string(contractAddress))

	urlStr := fmt.Sprintf("%s%s/source-code?%s", BASE_URL, contractEndpoint, params.Encode())
	return fetchApi[any](urlStr)
}

func GetLatestBlock() (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/latest", BASE_URL, blocksEndpoint)
	return fetchApi[any](urlStr)
}

func GetLatestBlockBurns(page int, size int) (*ApiResponse[any], error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than or equal to 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s%s/latest/burns?%s", BASE_URL, blocksEndpoint, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetLatestBlockRewards(blockNumber int) (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/latest/rewards?blockNumber=%d", BASE_URL, blocksEndpoint, blockNumber)

	return fetchApi[any](urlStr)
}

func GetBlock(blockNumber int64) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("blockNumber", fmt.Sprintf("%d", blockNumber))

	urlStr := fmt.Sprintf("%s%s?%s", BASE_URL, blocksEndpoint, params.Encode())
	return fetchApi[any](urlStr)
}

func GetBlocks(
	blockNumber int,
	blockNumberStart *int,
	blockNumberEnd *int,
	page int,
	size int,
) (*ApiResponse[any], error) {
	queryParams := url.Values{}

	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}

	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}

	if page >= 1 {
		queryParams.Add("page", fmt.Sprintf("%d", page))
	}

	if size >= 1 && size <= 2000 {
		queryParams.Add("size", fmt.Sprintf("%d", size))
	}

	urlStr := fmt.Sprintf("%s%s?blockNumber=%d&%s", BASE_URL, blocksEndpoint, blockNumber, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetBlockBurns(blockNumber int) (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/%d/burns", BASE_URL, blocksEndpoint, blockNumber)
	return fetchApi[any](urlStr)
}

func GetBlockRewards(blockNumber int) (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/%d/rewards", BASE_URL, blocksEndpoint, blockNumber)
	return fetchApi[any](urlStr)
}

func GetInternalTransactionsOfBlock(blockNumber int, page int, size int) (*ApiResponse[any], error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s%s/%d/internal-transactions?%s", BASE_URL, blocksEndpoint, blockNumber, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetTransactionsOfBlock(blockNumber int, transactionType *string, page int, size int) (*ApiResponse[any], error) {
	queryParams := url.Values{}

	if transactionType != nil {
		queryParams.Add("type", *transactionType)
	}

	if page >= 1 {
		queryParams.Add("page", fmt.Sprintf("%d", page))
	}

	if size >= 1 && size <= 2000 {
		queryParams.Add("size", fmt.Sprintf("%d", size))
	}

	urlStr := fmt.Sprintf("%s%s/%d/transactions?%s", BASE_URL, blocksEndpoint, blockNumber, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetTransaction(transactionHash string) (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/%s", BASE_URL, transactionEndpoint, transactionHash)
	return fetchApi[any](urlStr)
}

func GetTransactionReceiptStatus(transactionHash string) (*ApiResponse[any], error) {
	params := url.Values{}
	params.Add("transactionHash", transactionHash)

	urlStr := fmt.Sprintf("%s%s/status?%s", BASE_URL, transactionReceipts, params.Encode())
	return fetchApi[any](urlStr)
}

func GetTransactionStatus(transactionHash string) (*ApiResponse[any], error) {
	urlStr := fmt.Sprintf("%s%s/%s/status", BASE_URL, transactionEndpoint, url.PathEscape(transactionHash))

	return fetchApi[any](urlStr)
}

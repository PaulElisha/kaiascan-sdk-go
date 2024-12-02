package kaiascan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	BASE_URL_MAINNET = "https://mainnet-oapi.kaiascan.io/"
	CHAIN_ID_MAINNET = "8217"

	BASE_URL_TESTNET = "https://kairos-oapi.kaiascan.io/"
	CHAIN_ID_TESTNET = "1001"

	BASE_URL = BASE_URL_MAINNET
	CHAIN_ID = CHAIN_ID_MAINNET

	tokensEndpoint      = "api/v1/tokens"
	nftsEndpoint        = "api/v1/nfts"
	blocksEndpoint      = "api/v1/blocks"
	transactionEndpoint = "api/v1/transactions"
	contractEndpoint    = "api/v1/contracts"
	transactionReceipts = "api/v1/transaction-receipts"
	accountEndpoint     = "api/v1/accounts"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func ConfigureSDK(isTestnet bool) {
	if isTestnet {
		BASE_URL = BASE_URL_TESTNET
		CHAIN_ID = CHAIN_ID_TESTNET
	} else {
		BASE_URL = BASE_URL_MAINNET
		CHAIN_ID = CHAIN_ID_MAINNET
	}
}

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

	urlStr := fmt.Sprintf("%s/%s/%s/key-histories?%s", BASE_URL, accountEndpoint, encodedAddress, queryParams.Encode())

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

func GetTokenHolders(
	tokenAddress string,
	page int,
	size int,
	holderAddress *string,
) (*ApiResponse[any], error) {
	queryParams := url.Values{}

	if holderAddress != nil {
		queryParams.Add("holderAddress", *holderAddress)
	}

	if page >= 1 {
		queryParams.Add("page", fmt.Sprintf("%d", page))
	}

	if size >= 1 && size <= 2000 {
		queryParams.Add("size", fmt.Sprintf("%d", size))
	}

	encodedTokenAddress := url.PathEscape(tokenAddress)

	urlStr := fmt.Sprintf("%s/%s/%sholders?%s", BASE_URL, tokensEndpoint, encodedTokenAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetBlocksByTimestamp(timestamp int64) (*ApiResponse[any], error) {
	if timestamp <= 0 {
		return nil, fmt.Errorf("timestamp must be a positive integer")
	}

	urlStr := fmt.Sprintf("%s/%s/timestamps/%d", BASE_URL, blocksEndpoint, timestamp)

	return fetchApi[any](urlStr)
}

func GetContractInfo(contractAddress string) (*ApiResponse[any], error) {
	if contractAddress == "" {
		return nil, fmt.Errorf("contract address is required")
	}

	urlStr := fmt.Sprintf("%s/%s/%s", BASE_URL, contractEndpoint, contractAddress)

	return fetchApi[any](urlStr)
}

func GetContractsInfo(contractAddresses []string) (*ApiResponse[any], error) {
	if len(contractAddresses) == 0 {
		return nil, fmt.Errorf("contract address list is required")
	}

	contractAddressesStr := strings.Join(contractAddresses, ",")

	queryParams := url.Values{}
	queryParams.Add("contractAddresses", contractAddressesStr)

	urlStr := fmt.Sprintf("%s/%s?%s", BASE_URL, contractEndpoint, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetContractAbi(contractAddress string) (*ApiResponse[any], error) {
	if contractAddress == "" {
		return nil, fmt.Errorf("contract address is required")
	}

	urlStr := fmt.Sprintf("%s/%s/%s/abi", BASE_URL, contractEndpoint, contractAddress)

	return fetchApi[any](urlStr)
}

func GetNftInfo(tokenAddress string) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}

	urlStr := fmt.Sprintf("%s/%s/%s", BASE_URL, nftsEndpoint, tokenAddress)

	return fetchApi[any](urlStr)
}

func GetNftHolders(
	tokenAddress string,
	page int,
	size int,
	tokenId *string,
) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if tokenId != nil {
		queryParams.Add("tokenId", *tokenId)
	}

	encodedTokenAddress := url.PathEscape(tokenAddress)

	urlStr := fmt.Sprintf("%s/%s/%s/holders?%s", BASE_URL, nftsEndpoint, encodedTokenAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetNftTransfers(
	tokenAddress string,
	page int,
	size int,
	tokenId *string,
	blockNumberStart *int,
	blockNumberEnd *int,
) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if tokenId != nil {
		queryParams.Add("tokenId", *tokenId)
	}
	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}

	encodedTokenAddress := url.PathEscape(tokenAddress)

	urlStr := fmt.Sprintf("%s/%s/%s/transfers?%s", BASE_URL, nftsEndpoint, encodedTokenAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetNftInventories(tokenAddress string, page int, size int, keyword *string) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	if keyword != nil {
		queryParams = append(queryParams, fmt.Sprintf("keyword=%s", *keyword))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/inventories?%s", BASE_URL, nftsEndpoint, tokenAddress, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTokenBurns(tokenAddress string, page int, size int, blockNumberStart *int, blockNumberEnd *int) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	if blockNumberStart != nil {
		queryParams = append(queryParams, fmt.Sprintf("blockNumberStart=%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams = append(queryParams, fmt.Sprintf("blockNumberEnd=%d", *blockNumberEnd))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/burns?%s", BASE_URL, tokensEndpoint, tokenAddress, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTokenTransfers(tokenAddress string, page int, size int, blockNumberStart *int, blockNumberEnd *int) (*ApiResponse[any], error) {
	if tokenAddress == "" {
		return nil, fmt.Errorf("token address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	if blockNumberStart != nil {
		queryParams = append(queryParams, fmt.Sprintf("blockNumberStart=%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams = append(queryParams, fmt.Sprintf("blockNumberEnd=%d", *blockNumberEnd))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/transfers?%s", BASE_URL, accountEndpoint, tokenAddress, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTransactionInputData(transactionHash string) (*ApiResponse[any], error) {
	if transactionHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}

	urlStr := fmt.Sprintf("%s/%s/%s/input-data", BASE_URL, transactionEndpoint, transactionHash)

	return fetchApi[any](urlStr)
}

func GetTransactionEventLogs(transactionHash string, page int, size int, signature *string) (*ApiResponse[any], error) {
	if transactionHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	if signature != nil {
		queryParams = append(queryParams, fmt.Sprintf("signature=%s", *signature))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/event-logs?%s", BASE_URL, transactionEndpoint, transactionHash, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTransactionInternalTransactions(transactionHash string, page int, size int) (*ApiResponse[any], error) {
	if transactionHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	urlStr := fmt.Sprintf("%s/%s/%s/internal-transactions?%s", BASE_URL, transactionEndpoint, transactionHash, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTransactionTokenTransfers(transactionHash string, page int, size int) (*ApiResponse[any], error) {
	if transactionHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	urlStr := fmt.Sprintf("%s/%s/%s/token-transfers?%s", BASE_URL, transactionEndpoint, transactionHash, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetTransactionNftTransfers(transactionHash string, page int, size int) (*ApiResponse[any], error) {
	if transactionHash == "" {
		return nil, fmt.Errorf("transaction hash is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := []string{
		fmt.Sprintf("page=%d", page),
		fmt.Sprintf("size=%d", size),
	}

	urlStr := fmt.Sprintf("%s/%s/%s/nft-transfers?%s", BASE_URL, transactionEndpoint, transactionHash, strings.Join(queryParams, "&"))

	return fetchApi[any](urlStr)
}

func GetAccountTokenBalances(accountAddress string, page int, size int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s/%s/%s/token-balances?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountNftTransfers(accountAddress string, page int, size int, contractAddress *string, blockNumberStart *int, blockNumberEnd *int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if contractAddress != nil {
		queryParams.Add("contractAddress", *contractAddress)
	}
	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/nft-transfers?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountKIP37NftBalances(accountAddress string, page int, size int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s/%s/%s/nft-balances/kip37?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountKIP17NftBalances(accountAddress string, page int, size int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s/%s/%s/nft-balances/kip17?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountEventLogs(accountAddress string, page int, size int, signature *string, blockNumberStart *int, blockNumberEnd *int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if signature != nil {
		queryParams.Add("signature", *signature)
	}
	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/event-logs?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountTokenTransfers(accountAddress string, page int, size int, contractAddress *string, blockNumberStart *int, blockNumberEnd *int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if contractAddress != nil {
		queryParams.Add("contractAddress", *contractAddress)
	}
	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/token-transfers?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountInfo(accountAddress string) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}

	urlStr := fmt.Sprintf("%s/%s/%s", BASE_URL, accountEndpoint, accountAddress)

	return fetchApi[any](urlStr)
}

func GetFeePaidTransactions(accountAddress string, page int, size int, blockNumberStart *int, blockNumberEnd *int, txType *string) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}
	if txType != nil {
		queryParams.Add("type", *txType)
	}

	urlStr := fmt.Sprintf("%s/%s/%s/fee-paid-transactions?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountTransactions(accountAddress string, page int, size int, blockNumberStart *int, blockNumberEnd *int, txType *string, directions []string) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	if blockNumberStart != nil {
		queryParams.Add("blockNumberStart", fmt.Sprintf("%d", *blockNumberStart))
	}
	if blockNumberEnd != nil {
		queryParams.Add("blockNumberEnd", fmt.Sprintf("%d", *blockNumberEnd))
	}
	if txType != nil {
		queryParams.Add("type", *txType)
	}
	if len(directions) > 0 {
		queryParams.Add("directions", strings.Join(directions, ","))
	}

	urlStr := fmt.Sprintf("%s/%s/%s/transactions?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

func GetAccountTokenDetails(accountAddress string, page int, size int) (*ApiResponse[any], error) {
	if accountAddress == "" {
		return nil, fmt.Errorf("account address is required")
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be >= 1")
	}
	if size < 1 || size > 2000 {
		return nil, fmt.Errorf("size must be between 1 and 2000")
	}

	queryParams := url.Values{}
	queryParams.Add("page", fmt.Sprintf("%d", page))
	queryParams.Add("size", fmt.Sprintf("%d", size))

	urlStr := fmt.Sprintf("%s/%s/%s/token-details?%s", BASE_URL, accountEndpoint, accountAddress, queryParams.Encode())

	return fetchApi[any](urlStr)
}

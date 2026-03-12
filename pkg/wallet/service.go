package wallet

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"path/filepath"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Service provides wallet functionality
type Service struct {
	config      Config
	workspace   string
	keystoreDir string
	clients     map[int]*ethclient.Client
	clientsMu   sync.RWMutex
	keystore    *keystore.KeyStore
	logger      *slog.Logger
}

// NewService creates a new wallet service
func NewService(config Config, workspace string) *Service {
	keystoreDir := filepath.Join(workspace, "wallets")

	return &Service{
		config:      config,
		workspace:   workspace,
		keystoreDir: keystoreDir,
		clients:     make(map[int]*ethclient.Client),
		keystore:    keystore.NewKeyStore(keystoreDir, keystore.StandardScryptN, keystore.StandardScryptP),
		logger:      slog.Default().With("component", "wallet"),
	}
}

// GetWorkspace returns the workspace path
func (s *Service) GetWorkspace() string {
	return s.workspace
}

// Initialize sets up the wallet service
func (s *Service) Initialize(ctx context.Context) error {
	s.logger.Info("initializing wallet service")

	// Validate configuration
	if err := s.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize blockchain clients for each chain
	for _, chain := range s.config.Chains {
		client, err := ethclient.Dial(chain.RPC)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", chain.Name, err)
		}

		s.clientsMu.Lock()
		s.clients[chain.ChainID] = client
		s.clientsMu.Unlock()

		s.logger.Info("connected to blockchain", "chain", chain.Name, "chain_id", chain.ChainID)
	}

	return nil
}

// Close cleans up resources
func (s *Service) Close() error {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for chainID, client := range s.clients {
		if client != nil {
			client.Close()
			s.logger.Info("disconnected from blockchain", "chain_id", chainID)
		}
	}

	return nil
}

// CreateWallet creates a new wallet
func (s *Service) CreateWallet(ctx context.Context, password string) (*WalletInfo, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// Restrict to only one wallet
	if len(s.keystore.Accounts()) > 0 {
		return nil, fmt.Errorf("only one wallet is allowed - cannot create multiple wallets")
	}

	account, err := s.keystore.NewAccount(password)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &WalletInfo{
		Address: account.Address.Hex(),
		Path:    account.URL.Path,
	}, nil
}

// GetWalletInfo returns wallet information
func (s *Service) GetWalletInfo(ctx context.Context, address common.Address) (*WalletInfo, error) {
	// Get the default chain (first configured chain)
	if len(s.config.Chains) == 0 {
		return nil, fmt.Errorf("no chains configured")
	}

	chain := s.config.Chains[0]
	return s.GetWalletInfoForChain(ctx, address, chain.ChainID)
}

// GetWalletInfoForChain returns wallet information for a specific chain
func (s *Service) GetWalletInfoForChain(ctx context.Context, address common.Address, chainID int) (*WalletInfo, error) {
	chain, err := s.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(chainID)
	if err != nil {
		return nil, err
	}

	balance, err := s.getBalance(ctx, client, address, chain)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &WalletInfo{
		Address: address.Hex(),
		Balance: balance,
		Chain:   chain.Name,
		ChainID: chainID,
	}, nil
}

// Transfer transfers funds to a recipient
func (s *Service) Transfer(ctx context.Context, from common.Address, to common.Address, amount *big.Int, password string, chainID int) (*types.Transaction, error) {
	chain, err := s.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	// For non-native chains, use ERC20 transfer
	if !chain.IsNative {
		return s.TransferToken(ctx, from, to, amount, password, chainID, common.HexToAddress(chain.GasToken))
	}

	return s.transferNative(ctx, from, to, amount, password, chain)
}

// TransferToken transfers ERC20 tokens
func (s *Service) TransferToken(ctx context.Context, from common.Address, to common.Address, amount *big.Int, password string, chainID int, tokenAddress common.Address) (*types.Transaction, error) {
	chain, err := s.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(chainID)
	if err != nil {
		return nil, err
	}

	// Find the account in keystore
	account, err := s.findAccount(from)
	if err != nil {
		return nil, err
	}

	// Unlock account
	if err := s.keystore.Unlock(account, password); err != nil {
		return nil, fmt.Errorf("failed to unlock account: %w", err)
	}
	defer s.keystore.Lock(account.Address)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Build ERC20 transfer transaction
	data, err := buildERC20TransferData(to, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to build transfer data: %w", err)
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &tokenAddress,
		Value: big.NewInt(0),
		Data:  data,
	})
	if err != nil {
		// If estimation fails, use a reasonable default with buffer
		s.logger.Warn("Failed to estimate gas limit, using default", "error", err)
		gasLimit = 200000 // Default gas limit with buffer
	} else {
		// Add 20% buffer to estimated gas limit
		gasLimit = uint64(float64(gasLimit) * 1.2)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // Value is 0 for token transfers
		gasLimit,
		gasPrice,
		data,
	)

	// Get chain ID
	chainIDBig := big.NewInt(int64(chainID))

	// Sign transaction
	signedTx, err := s.keystore.SignTx(account, tx, chainIDBig)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	s.logger.Info("token transfer sent",
		"tx_hash", signedTx.Hash().Hex(),
		"from", from.Hex(),
		"to", to.Hex(),
		"amount", amount.String(),
		"chain", chain.Name,
	)

	return signedTx, nil
}

// GetAccounts returns all accounts in the keystore
func (s *Service) GetAccounts() []accounts.Account {
	return s.keystore.Accounts()
}

// GetConfig returns the wallet configuration
func (s *Service) GetConfig() Config {
	return s.config
}

// GetChainByID returns chain configuration by ID
func (s *Service) GetChainByID(chainID int) (*ChainConfig, error) {
	for _, chain := range s.config.Chains {
		if chain.ChainID == chainID {
			return &chain, nil
		}
	}
	return nil, fmt.Errorf("chain with ID %d not found", chainID)
}

// CallContractMethod calls a read-only smart contract function
func (s *Service) CallContractMethod(ctx context.Context, contractAddress common.Address, abiType string, method string, params ...interface{}) (interface{}, error) {
	var abiJson abi.ABI

	switch abiType {
	case "erc20":
		abiJson = erc20ABIJson
	default:
		return nil, fmt.Errorf("unsupported ABI type: %s", abiType)
	}

	chainID := 7441 // Default to ClawSwift chain
	client, err := s.getClient(chainID)
	if err != nil {
		return nil, err
	}

	// Pack the method call data
	data, err := abiJson.Pack(method, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method call: %w", err)
	}

	// Call the contract
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract method: %w", err)
	}

	// Unpack the result
	var unpacked interface{}
	err = abiJson.UnpackIntoInterface(&unpacked, method, result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack method result: %w", err)
	}

	return unpacked, nil
}

// ExecuteContractMethod executes a write smart contract function
func (s *Service) ExecuteContractMethod(ctx context.Context, from common.Address, contractAddress common.Address, abiType string, method string, value *big.Int, password string, params ...interface{}) (*types.Transaction, error) {
	var abiJson abi.ABI

	switch abiType {
	case "erc20":
		abiJson = erc20ABIJson
	default:
		return nil, fmt.Errorf("unsupported ABI type: %s", abiType)
	}

	chainID := 7441 // Default to ClawSwift chain
	chain, err := s.GetChainByID(chainID)
	if err != nil {
		return nil, err
	}

	client, err := s.getClient(chainID)
	if err != nil {
		return nil, err
	}

	// Find the account in keystore
	account, err := s.findAccount(from)
	if err != nil {
		return nil, err
	}

	// Unlock account
	if err := s.keystore.Unlock(account, password); err != nil {
		return nil, fmt.Errorf("failed to unlock account: %w", err)
	}
	defer s.keystore.Lock(account.Address)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Pack the method call data
	data, err := abiJson.Pack(method, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method call: %w", err)
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &contractAddress,
		Value: value,
		Data:  data,
	})
	if err != nil {
		// If estimation fails, use a reasonable default with buffer
		s.logger.Warn("Failed to estimate gas limit, using default", "error", err)
		gasLimit = 200000 // Default gas limit with buffer
	} else {
		// Add 20% buffer to estimated gas limit
		gasLimit = uint64(float64(gasLimit) * 1.2)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		contractAddress,
		value,
		gasLimit,
		gasPrice,
		data,
	)

	// Get chain ID
	chainIDBig := big.NewInt(int64(chainID))

	// Sign transaction
	signedTx, err := s.keystore.SignTx(account, tx, chainIDBig)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	s.logger.Info("contract method executed",
		"tx_hash", signedTx.Hash().Hex(),
		"from", from.Hex(),
		"contract", contractAddress.Hex(),
		"method", method,
		"chain", chain.Name,
	)

	return signedTx, nil
}

// Helper methods

func (s *Service) getClient(chainID int) (*ethclient.Client, error) {
	s.clientsMu.RLock()
	client, exists := s.clients[chainID]
	s.clientsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client for chain %d not found", chainID)
	}

	return client, nil
}

func (s *Service) findAccount(address common.Address) (accounts.Account, error) {
	for _, account := range s.keystore.Accounts() {
		if account.Address == address {
			return account, nil
		}
	}
	return accounts.Account{}, fmt.Errorf("account %s not found", address.Hex())
}

func (s *Service) getBalance(ctx context.Context, client *ethclient.Client, address common.Address, chain *ChainConfig) (*big.Float, error) {
	if chain.IsNative {
		// Get native balance
		balance, err := client.BalanceAt(ctx, address, nil)
		if err != nil {
			return nil, err
		}
		return new(big.Float).SetInt(balance), nil
	}

	// Get ERC20 token balance
	tokenAddress := common.HexToAddress(chain.GasToken)
	return s.getERC20Balance(ctx, client, address, tokenAddress)
}

func (s *Service) transferNative(ctx context.Context, from, to common.Address, amount *big.Int, password string, chain *ChainConfig) (*types.Transaction, error) {
	client, err := s.getClient(chain.ChainID)
	if err != nil {
		return nil, err
	}

	// Find the account in keystore
	account, err := s.findAccount(from)
	if err != nil {
		return nil, err
	}

	// Unlock account
	if err := s.keystore.Unlock(account, password); err != nil {
		return nil, fmt.Errorf("failed to unlock account: %w", err)
	}
	defer s.keystore.Lock(account.Address)

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: amount,
		Data:  nil,
	})
	if err != nil {
		// If estimation fails, use a reasonable default with buffer
		s.logger.Warn("Failed to estimate gas limit, using default", "error", err)
		gasLimit = 21000 // Standard transfer gas limit
	} else {
		// Add 20% buffer to estimated gas limit
		gasLimit = uint64(float64(gasLimit) * 1.2)
	}

	// Create transaction
	tx := types.NewTransaction(
		nonce,
		to,
		amount,
		gasLimit,
		gasPrice,
		nil, // No data for simple transfers
	)

	// Get chain ID
	chainIDBig := big.NewInt(int64(chain.ChainID))

	// Sign transaction
	signedTx, err := s.keystore.SignTx(account, tx, chainIDBig)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	s.logger.Info("native transfer sent",
		"tx_hash", signedTx.Hash().Hex(),
		"from", from.Hex(),
		"to", to.Hex(),
		"amount", amount.String(),
		"chain", chain.Name,
	)

	return signedTx, nil
}

package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sipeed/domeclaw/pkg/blockchain"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// WalletService manages Ethereum wallet operations
type WalletService struct {
	walletDir        string
	keystore         *keystore.KeyStore
	blockchainClient *blockchain.Client
	transferService  *blockchain.TransferService
	contractService  *blockchain.ContractService
	abiManager       *blockchain.ABIManager
	chainConfig      *config.EVMChain
}

// NewWalletService creates a new wallet service
func NewWalletService(workspace string, cfg *config.WalletConfig) *WalletService {
	walletDir := filepath.Join(workspace, "wallet")
	os.MkdirAll(walletDir, 0o700)

	ks := keystore.NewKeyStore(walletDir, keystore.StandardScryptN, keystore.StandardScryptP)

	// Initialize blockchain client
	bcClient := blockchain.NewClient()

	var chainConfig *config.EVMChain
	if cfg != nil && cfg.Enabled && len(cfg.Chains) > 0 {
		chain := &cfg.Chains[0]
		logger.InfoCF("wallet", "Connecting to blockchain", map[string]any{
			"chain": chain.Name,
			"rpc":   chain.RPC,
		})
		if err := bcClient.AddChain(chain); err != nil {
			logger.ErrorCF("wallet", "Failed to connect to blockchain", map[string]any{
				"chain": chain.Name,
				"rpc":   chain.RPC,
				"error": err.Error(),
			})
		} else {
			chainConfig = chain
			logger.InfoCF("wallet", "Connected to blockchain successfully", map[string]any{
				"chain":   chain.Name,
				"chainId": chain.ChainID,
			})
		}
	} else {
		logger.WarnCF("wallet", "Blockchain not configured", map[string]any{
			"enabled": cfg != nil && cfg.Enabled,
			"chains":  len(cfg.Chains),
		})
	}

	// Initialize services
	transferService := blockchain.NewTransferService(bcClient)

	// Initialize ABI manager
	abiManager, err := blockchain.NewABIManager(workspace)
	if err != nil {
		logger.ErrorCF("wallet", "Failed to initialize ABI manager", map[string]any{
			"error": err.Error(),
		})
	}

	contractService := blockchain.NewContractService(bcClient, abiManager)

	return &WalletService{
		walletDir:        walletDir,
		keystore:         ks,
		blockchainClient: bcClient,
		transferService:  transferService,
		contractService:  contractService,
		abiManager:       abiManager,
		chainConfig:      chainConfig,
	}
}

// WalletExists checks if wallet already exists
func (ws *WalletService) WalletExists() bool {
	accounts := ws.keystore.Accounts()
	return len(accounts) > 0
}

// CreateWallet creates a new wallet with PIN
func (ws *WalletService) CreateWallet(pin string) (common.Address, error) {
	if ws.WalletExists() {
		return common.Address{}, ErrWalletAlreadyExists
	}

	if !ValidatePIN(pin) {
		return common.Address{}, ErrInvalidPINFormat
	}

	account, err := ws.keystore.NewAccount(pin)
	if err != nil {
		return common.Address{}, ErrKeystoreFailed
	}

	// Save PIN
	pinFile := filepath.Join(ws.walletDir, "pin.json")
	pinData := PINStore{
		PIN:       pin,
		CreatedAt: time.Now(),
	}
	pinJSON, _ := json.MarshalIndent(pinData, "", "  ")
	os.WriteFile(pinFile, pinJSON, 0o600)

	// Save wallet info
	walletFile := filepath.Join(ws.walletDir, "wallet.json")
	walletInfo := WalletInfo{
		Address:   account.Address,
		CreatedAt: time.Now(),
		Encrypted: true,
	}
	walletJSON, _ := json.MarshalIndent(walletInfo, "", "  ")
	os.WriteFile(walletFile, walletJSON, 0o600)

	logger.InfoCF("wallet", "Wallet created", map[string]any{
		"address": account.Address.Hex(),
	})

	return account.Address, nil
}

// GetAddress returns the wallet address
func (ws *WalletService) GetAddress() (common.Address, error) {
	accounts := ws.keystore.Accounts()
	if len(accounts) == 0 {
		return common.Address{}, ErrWalletNotCreated
	}
	return accounts[0].Address, nil
}

// Unlock unlocks the wallet with PIN
func (ws *WalletService) Unlock(pin string) error {
	accounts := ws.keystore.Accounts()
	if len(accounts) == 0 {
		return ErrWalletNotCreated
	}

	if err := ws.keystore.Unlock(accounts[0], pin); err != nil {
		return ErrInvalidPIN
	}

	return nil
}

// Lock locks the wallet
func (ws *WalletService) Lock() error {
	accounts := ws.keystore.Accounts()
	if len(accounts) == 0 {
		return ErrWalletNotCreated
	}

	return ws.keystore.Lock(accounts[0].Address)
}

// GetBalance returns wallet balance
func (ws *WalletService) GetBalance() (string, error) {
	if ws.blockchainClient == nil {
		logger.WarnCF("wallet", "Blockchain client not initialized", nil)
		return "0", nil
	}

	if ws.chainConfig == nil {
		logger.WarnCF("wallet", "Chain config not set", nil)
		return "0", nil
	}

	address, err := ws.GetAddress()
	if err != nil {
		return "0", err
	}

	logger.InfoCF("wallet", "Getting balance", map[string]any{
		"address": address.Hex(),
		"chain":   ws.chainConfig.Name,
	})

	balance, err := ws.blockchainClient.GetBalance(context.Background(), ws.chainConfig.ChainID, address)
	if err != nil {
		logger.ErrorCF("wallet", "Failed to get balance", map[string]any{
			"address": address.Hex(),
			"error":   err.Error(),
		})
		return "0", err
	}

	logger.InfoCF("wallet", "Balance retrieved", map[string]any{
		"balance": balance.FormattedBalance(),
		"token":   balance.TokenName,
	})

	return balance.FormattedBalance(), nil
}

// Transfer sends tokens
func (ws *WalletService) Transfer(to common.Address, amount *big.Int, pin string) (common.Hash, error) {
	if ws.chainConfig == nil {
		return common.Hash{}, fmt.Errorf("blockchain not configured")
	}

	address, err := ws.GetAddress()
	if err != nil {
		return common.Hash{}, err
	}

	// Unlock wallet
	if err := ws.Unlock(pin); err != nil {
		return common.Hash{}, err
	}
	defer ws.Lock()

	// Create signer function
	signer := ws.createSigner(address)

	// Transfer
	if ws.chainConfig.IsNative {
		return ws.transferService.TransferNative(
			context.Background(),
			ws.chainConfig.ChainID,
			address,
			to,
			amount,
			signer,
		)
	} else {
		tokenAddress := common.HexToAddress(ws.chainConfig.GasToken)
		return ws.transferService.TransferERC20(
			context.Background(),
			ws.chainConfig.ChainID,
			address,
			tokenAddress,
			to,
			amount,
			signer,
		)
	}
}

// CallContract calls a read-only contract function
func (ws *WalletService) CallContract(contractAddress common.Address, abiName, method string, args []interface{}) (interface{}, error) {
	if ws.chainConfig == nil {
		return nil, fmt.Errorf("blockchain not configured")
	}

	return ws.contractService.CallContract(
		context.Background(),
		ws.chainConfig.ChainID,
		contractAddress,
		abiName,
		method,
		args,
	)
}

// WriteContract calls a state-changing contract function
func (ws *WalletService) WriteContract(contractAddress common.Address, abiName, method string, args []interface{}, value *big.Int, pin string) (common.Hash, error) {
	if ws.chainConfig == nil {
		return common.Hash{}, fmt.Errorf("blockchain not configured")
	}

	address, err := ws.GetAddress()
	if err != nil {
		return common.Hash{}, err
	}

	// Unlock wallet
	if err := ws.Unlock(pin); err != nil {
		return common.Hash{}, err
	}
	defer ws.Lock()

	signer := ws.createSigner(address)

	return ws.contractService.WriteContract(
		context.Background(),
		ws.chainConfig.ChainID,
		address,
		contractAddress,
		abiName,
		method,
		args,
		value,
		signer,
	)
}

// UploadABI uploads a new ABI
func (ws *WalletService) UploadABI(name, abiJSON string) error {
	return ws.abiManager.UploadABI(name, abiJSON)
}

// ListABIs lists all available ABIs
func (ws *WalletService) ListABIs() ([]string, error) {
	return ws.abiManager.ListABIs()
}

// createSigner creates a transaction signer function
func (ws *WalletService) createSigner(address common.Address) blockchain.SignerFunc {
	return func(ctx context.Context, chainID int64, tx *types.Transaction) (*types.Transaction, error) {
		chainIDBig := big.NewInt(chainID)

		return ws.keystore.SignTx(
			accounts.Account{Address: address},
			tx,
			chainIDBig,
		)
	}
}

package wallet

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// WalletInfo represents wallet information
type WalletInfo struct {
	Address string    `json:"address"`
	Balance *big.Float `json:"balance,omitempty"`
	Chain   string    `json:"chain,omitempty"`
	ChainID int       `json:"chain_id,omitempty"`
	Path    string    `json:"path,omitempty"`
}

// TransferRequest represents a transfer request
type TransferRequest struct {
	From     common.Address `json:"from"`
	To       common.Address `json:"to"`
	Amount   *big.Int       `json:"amount"`
	Password string         `json:"password"`
	ChainID  int            `json:"chain_id"`
}

// TransferTokenRequest represents an ERC20 token transfer request
type TransferTokenRequest struct {
	From         common.Address `json:"from"`
	To           common.Address `json:"to"`
	Amount       *big.Int       `json:"amount"`
	Password     string         `json:"password"`
	ChainID      int            `json:"chain_id"`
	TokenAddress common.Address `json:"token_address"`
}
package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// TransferService handles token transfers
type TransferService struct {
	client *Client
}

// NewTransferService creates a new transfer service
func NewTransferService(client *Client) *TransferService {
	return &TransferService{
		client: client,
	}
}

// TransferNative sends native tokens
func (ts *TransferService) TransferNative(
	ctx context.Context,
	chainID int64,
	from common.Address,
	to common.Address,
	amount *big.Int,
	signer SignerFunc,
) (common.Hash, error) {
	client, ok := ts.client.GetClient(chainID)
	if !ok {
		return common.Hash{}, fmt.Errorf("chain %d not found", chainID)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: amount,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, nil)

	// Sign transaction
	signedTx, err := signer(ctx, chainID, tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash(), nil
}

// TransferERC20 sends ERC20 tokens
func (ts *TransferService) TransferERC20(
	ctx context.Context,
	chainID int64,
	from common.Address,
	tokenAddress common.Address,
	to common.Address,
	amount *big.Int,
	signer SignerFunc,
) (common.Hash, error) {
	client, ok := ts.client.GetClient(chainID)
	if !ok {
		return common.Hash{}, fmt.Errorf("chain %d not found", chainID)
	}

	logger.InfoCF("blockchain", "TransferERC20 started", map[string]any{
		"from":   from.Hex(),
		"to":     to.Hex(),
		"token":  tokenAddress.Hex(),
		"amount": amount.String(),
	})

	// ERC20 transfer function: transfer(address,uint256)
	transferSig := []byte{0xa9, 0x05, 0x9c, 0xbb}

	// Prepare call data: function selector + to (32 bytes) + amount (32 bytes)
	callData := make([]byte, 0, 4+32+32)
	callData = append(callData, transferSig...)
	callData = append(callData, common.LeftPadBytes(to.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(amount.Bytes(), 32)...)

	logger.DebugCF("blockchain", "Call data prepared", map[string]any{
		"data": common.Bytes2Hex(callData),
	})

	// Get nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas (use default if estimation fails)
	gasLimit := uint64(100000) // Default for ERC20 transfer
	estimatedGas, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &tokenAddress,
		Data: callData,
	})
	if err != nil {
		logger.WarnCF("blockchain", "Gas estimation failed, using default", map[string]any{
			"error": err.Error(),
		})
		// Try to get more info about why it failed
		// This might be because the transfer would fail (insufficient balance, etc.)
	} else {
		gasLimit = estimatedGas + 10000 // Add buffer
		logger.InfoCF("blockchain", "Gas estimated", map[string]any{
			"estimated": estimatedGas,
			"final":     gasLimit,
		})
	}

	// Create transaction
	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, callData)

	// Sign transaction
	signedTx, err := signer(ctx, chainID, tx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		logger.ErrorCF("blockchain", "Send transaction failed", map[string]any{
			"error": err.Error(),
		})
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	logger.InfoCF("blockchain", "TransferERC20 successful", map[string]any{
		"tx_hash": signedTx.Hash().Hex(),
	})

	return signedTx.Hash(), nil
}

// SignerFunc is a function that signs transactions
type SignerFunc func(ctx context.Context, chainID int64, tx *types.Transaction) (*types.Transaction, error)

// EstimateGasCost estimates gas cost for a transaction
func (ts *TransferService) EstimateGasCost(
	ctx context.Context,
	chainID int64,
	from common.Address,
	to *common.Address,
	amount *big.Int,
	data []byte,
) (*big.Int, error) {
	client, ok := ts.client.GetClient(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: amount,
		Data:  data,
	})
	if err != nil {
		return nil, err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	gasCost := new(big.Int).Mul(big.NewInt(0).SetUint64(gasLimit), gasPrice)
	return gasCost, nil
}

// GetTransactionStatus gets transaction status
func (ts *TransferService) GetTransactionStatus(ctx context.Context, chainID int64, txHash common.Hash) (*TransactionStatus, error) {
	client, ok := ts.client.GetClient(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	receipt, err := client.TransactionReceipt(ctx, txHash)
	if err != nil {
		if err == ethereum.NotFound {
			return &TransactionStatus{
				Hash:    txHash,
				Status:  "pending",
				Success: false,
			}, nil
		}
		return nil, err
	}

	return &TransactionStatus{
		Hash:        txHash,
		Status:      "confirmed",
		Success:     receipt.Status == types.ReceiptStatusSuccessful,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
	}, nil
}

// TransactionStatus contains transaction status information
type TransactionStatus struct {
	Hash        common.Hash
	Status      string
	Success     bool
	BlockNumber uint64
	GasUsed     uint64
}

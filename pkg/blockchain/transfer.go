package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

	// ERC20 transfer function: transfer(address,uint256)
	transferSig := common.HexToHash("0xa9059cbb")

	// Prepare call data
	callData := append(transferSig.Bytes(), common.LeftPadBytes(to.Bytes(), 32)...)
	callData = append(callData, common.LeftPadBytes(amount.Bytes(), 32)...)

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
		From: from,
		To:   &tokenAddress,
		Data: callData,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
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
		return common.Hash{}, fmt.Errorf("failed to send transaction: %w", err)
	}

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

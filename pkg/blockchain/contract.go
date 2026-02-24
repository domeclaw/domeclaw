package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ContractService handles smart contract interactions
type ContractService struct {
	client     *Client
	abiManager *ABIManager
}

// NewContractService creates a new contract service
func NewContractService(client *Client, abiManager *ABIManager) *ContractService {
	return &ContractService{
		client:     client,
		abiManager: abiManager,
	}
}

// CallContract calls a read-only contract function
func (cs *ContractService) CallContract(
	ctx context.Context,
	chainID int64,
	contractAddress common.Address,
	abiName string,
	method string,
	args []interface{},
) (interface{}, error) {
	// Get ABI
	parsedABI, err := cs.abiManager.GetABI(abiName)
	if err != nil {
		return nil, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Get client
	client, ok := cs.client.GetClient(chainID)
	if !ok {
		return nil, fmt.Errorf("chain %d not found", chainID)
	}

	// Pack method call
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method call: %w", err)
	}

	// Call contract
	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	// Unpack result
	var methodABI abi.Method
	found := false
	for _, m := range parsedABI.Methods {
		if m.Name == method {
			methodABI = m
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	if len(methodABI.Outputs) == 0 {
		return nil, nil
	}

	outputs, err := methodABI.Outputs.Unpack(result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	if len(outputs) == 1 {
		return outputs[0], nil
	}

	return outputs, nil
}

// WriteContract calls a state-changing contract function
func (cs *ContractService) WriteContract(
	ctx context.Context,
	chainID int64,
	from common.Address,
	contractAddress common.Address,
	abiName string,
	method string,
	args []interface{},
	value *big.Int,
	signer SignerFunc,
) (common.Hash, error) {
	// Get ABI
	parsedABI, err := cs.abiManager.GetABI(abiName)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Get client
	client, ok := cs.client.GetClient(chainID)
	if !ok {
		return common.Hash{}, fmt.Errorf("chain %d not found", chainID)
	}

	// Pack method call
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack method call: %w", err)
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
		To:    &contractAddress,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, contractAddress, value, gasLimit, gasPrice, data)

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

// DeployContract deploys a new smart contract
func (cs *ContractService) DeployContract(
	ctx context.Context,
	chainID int64,
	from common.Address,
	abiName string,
	bytecode []byte,
	constructorArgs []interface{},
	signer SignerFunc,
) (common.Hash, error) {
	// Get ABI
	parsedABI, err := cs.abiManager.GetABI(abiName)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get ABI: %w", err)
	}

	// Get client
	client, ok := cs.client.GetClient(chainID)
	if !ok {
		return common.Hash{}, fmt.Errorf("chain %d not found", chainID)
	}

	// Pack constructor arguments
	var data []byte
	if len(parsedABI.Constructor.Inputs) > 0 {
		data, err = parsedABI.Pack("", constructorArgs...)
		if err != nil {
			return common.Hash{}, fmt.Errorf("failed to pack constructor args: %w", err)
		}
	}

	// Append constructor data to bytecode
	if len(data) > 0 {
		bytecode = append(bytecode, data...)
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
		From: from,
		Data: bytecode,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create deployment transaction
	tx := types.NewContractCreation(nonce, big.NewInt(0), gasLimit, gasPrice, bytecode)

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

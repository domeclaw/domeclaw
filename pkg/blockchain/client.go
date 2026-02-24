package blockchain

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sipeed/domeclaw/pkg/config"
	"github.com/sipeed/domeclaw/pkg/logger"
)

// Client manages connections to multiple EVM blockchains
type Client struct {
	mu         sync.RWMutex
	rpcClients map[int64]*ethclient.Client
	chains     map[int64]*config.EVMChain
}

// NewClient creates a new blockchain client
func NewClient() *Client {
	return &Client{
		rpcClients: make(map[int64]*ethclient.Client),
		chains:     make(map[int64]*config.EVMChain),
	}
}

// AddChain adds a new EVM chain configuration and connects to it
func (c *Client) AddChain(chain *config.EVMChain) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if _, exists := c.rpcClients[chain.ChainID]; exists {
		logger.InfoCF("blockchain", "Chain already connected", map[string]any{
			"chain": chain.Name,
		})
		return nil
	}

	// Connect to RPC
	client, err := ethclient.Dial(chain.RPC)
	if err != nil {
		return fmt.Errorf("failed to connect to %s RPC: %w", chain.Name, err)
	}

	// Verify chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to get chain ID for %s: %w", chain.Name, err)
	}

	if chainID.Int64() != chain.ChainID {
		client.Close()
		return fmt.Errorf("chain ID mismatch: expected %d, got %d", chain.ChainID, chainID.Int64())
	}

	c.rpcClients[chain.ChainID] = client
	c.chains[chain.ChainID] = chain

	logger.InfoCF("blockchain", "Connected to chain", map[string]any{
		"name":    chain.Name,
		"chainId": chain.ChainID,
		"rpc":     chain.RPC,
		"native":  chain.IsNative,
	})

	return nil
}

// GetClient returns the RPC client for a specific chain
func (c *Client) GetClient(chainID int64) (*ethclient.Client, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client, ok := c.rpcClients[chainID]
	return client, ok
}

// GetChain returns chain configuration
func (c *Client) GetChain(chainID int64) (*config.EVMChain, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chain, ok := c.chains[chainID]
	return chain, ok
}

// ListChains returns all configured chains
func (c *Client) ListChains() []*config.EVMChain {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chains := make([]*config.EVMChain, 0, len(c.chains))
	for _, chain := range c.chains {
		chains = append(chains, chain)
	}
	return chains
}

// Close closes all RPC connections
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for chainID, client := range c.rpcClients {
		client.Close()
		logger.InfoCF("blockchain", "Disconnected from chain", map[string]any{
			"chainId": chainID,
		})
	}
}

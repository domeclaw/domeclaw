package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// ABIManager manages smart contract ABIs
type ABIManager struct {
	mu           sync.RWMutex
	workspaceDir string
	abis         map[string]*abi.ABI
}

// NewABIManager creates a new ABI manager
func NewABIManager(workspaceDir string) (*ABIManager, error) {
	abisDir := filepath.Join(workspaceDir, "abis")
	if err := os.MkdirAll(abisDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create ABIs directory: %w", err)
	}

	manager := &ABIManager{
		workspaceDir: workspaceDir,
		abis:         make(map[string]*abi.ABI),
	}

	// Load existing ABIs
	if err := manager.loadAllABIs(); err != nil {
		return nil, fmt.Errorf("failed to load ABIs: %w", err)
	}

	return manager, nil
}

// UploadABI uploads a new ABI
func (m *ABIManager) UploadABI(name, abiJSON string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return fmt.Errorf("invalid ABI JSON: %w", err)
	}

	// Save to file
	abisDir := filepath.Join(m.workspaceDir, "abis")
	abiFile := filepath.Join(abisDir, fmt.Sprintf("%s.json", name))

	if err := ioutil.WriteFile(abiFile, []byte(abiJSON), 0o644); err != nil {
		return fmt.Errorf("failed to save ABI file: %w", err)
	}

	// Store in memory
	m.abis[name] = &parsedABI

	return nil
}

// GetABI gets an ABI by name
func (m *ABIManager) GetABI(name string) (*abi.ABI, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	abi, ok := m.abis[name]
	if !ok {
		return nil, fmt.Errorf("ABI '%s' not found", name)
	}

	return abi, nil
}

// DeleteABI deletes an ABI
func (m *ABIManager) DeleteABI(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Delete from memory
	delete(m.abis, name)

	// Delete from file
	abisDir := filepath.Join(m.workspaceDir, "abis")
	abiFile := filepath.Join(abisDir, fmt.Sprintf("%s.json", name))

	if err := os.Remove(abiFile); err != nil {
		return fmt.Errorf("failed to delete ABI file: %w", err)
	}

	return nil
}

// ListABIs lists all available ABIs
func (m *ABIManager) ListABIs() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.abis))
	for name := range m.abis {
		names = append(names, name)
	}

	return names, nil
}

// loadAllABIs loads all ABIs from disk
func (m *ABIManager) loadAllABIs() error {
	abisDir := filepath.Join(m.workspaceDir, "abis")

	files, err := ioutil.ReadDir(abisDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		name := file.Name()[:len(file.Name())-5] // Remove .json extension

		abiFile := filepath.Join(abisDir, file.Name())
		data, err := ioutil.ReadFile(abiFile)
		if err != nil {
			continue
		}

		var parsedABI abi.ABI
		if err := json.Unmarshal(data, &parsedABI); err != nil {
			continue
		}

		m.abis[name] = &parsedABI
	}

	return nil
}

# Ethereum Wallet Integration Implementation Plan

## Overview
This plan outlines the step-by-step implementation of Ethereum wallet functionality for Picoclaw, including support for both native ETH transfers and ERC20 token transfers, with comprehensive access control and multi-chain configuration support.

## Implementation Strategy

### Phase 1: Foundation Setup (Tasks 1-2)
**Objective**: Establish the basic infrastructure and dependencies

#### 1.1 Add Go Ethereum Dependencies
- Add `github.com/ethereum/go-ethereum` to go.mod
- Add required sub-packages for keystore, accounts, and RPC interactions
- Verify dependency compatibility with existing project dependencies

#### 1.2 Create Blockchain Configuration Module
- Create `pkg/blockchain/config.go` for chain configuration parsing
- Implement structs for chain configuration with fields:
  - `name`, `chain_id`, `rpc`, `explorer`, `currency`
  - `is_native`, `gas_token`, `gas_token_name`, `decimal`
- Add validation for required fields and proper formatting

#### 1.3 Setup RPC Connection Infrastructure
- Create `pkg/blockchain/client.go` for Ethereum client management
- Implement connection pooling and retry logic
- Add support for multiple RPC endpoints per chain
- Implement health checking for RPC connections

#### 1.4 Add ERC20 Contract Interaction
- Create `pkg/blockchain/erc20.go` for ERC20 token operations
- Implement standard ERC20 ABI interface
- Add methods for balance queries and transfer transactions
- Create contract interaction utilities

#### 2.1 Create Wallet Command Package Structure
- Create directory structure: `cmd/picoclaw/internal/wallet/`
- Implement base command structure following existing patterns
- Create separate files for each command: `create.go`, `transfer.go`, `info.go`, `transfer_token.go`
- Implement shared utilities in `helpers.go`

#### 2.2 Integrate with Existing Command System
- Study existing command registration patterns in `cmd/picoclaw/internal/`
- Implement wallet command registration in the main command registry
- Add proper command routing and argument parsing
- Ensure integration with existing help system

#### 2.3 Implement Command Routing
- Create command parser for `/wallet` commands
- Implement subcommand routing (create, transfer, transfer_token, info)
- Add argument validation and error handling
- Implement command aliases and shortcuts

#### 2.4 Add Access Control Middleware
- Create access control middleware in `pkg/auth/wallet_access.go`
- Implement user ID extraction from different channel types
- Create authorization check logic against `allow_from` lists
- Add error handling for unauthorized access attempts

#### 2.5 Implement User ID Validation
- Create user ID extraction utilities for different channels (Telegram, Discord, etc.)
- Implement channel-specific ID parsing
- Add validation for user ID formats
- Create mapping between channel user IDs and internal representations

### Phase 2: Core Wallet Functionality (Tasks 3-5)
**Objective**: Implement the core wallet operations

#### 3.1 Implement Wallet Creation Logic
- Create wallet generation using go-ethereum keystore
- Implement secure random account generation
- Add BIP39 mnemonic support (optional enhancement)
- Create wallet metadata storage

#### 3.2 Implement JSON Keystore Generation
- Create keystore files compatible with geth format
- Implement proper encryption using user-provided passwords
- Add keystore file naming and organization
- Create backup and recovery mechanisms

#### 3.3 Add Password Encryption
- Implement secure password handling using bcrypt or similar
- Add password strength validation
- Create password confirmation flows
- Implement secure password storage (hashed, never plain text)

#### 3.4 Create Wallet File Management
- Implement wallet file storage in user directory
- Create wallet indexing and lookup system
- Add wallet backup and restore functionality
- Implement wallet file encryption at rest

#### 4.1 Implement Native ETH Transfer Logic
- Create transaction building for native ETH transfers
- Implement gas estimation and fee calculation
- Add nonce management for transaction ordering
- Create transaction signing with private key

#### 4.2 Implement Transaction Signing
- Create secure transaction signing process
- Implement private key decryption for signing
- Add signature validation and verification
- Create transaction encoding and serialization

#### 4.3 Add Transaction Broadcasting
- Implement RPC transaction submission
- Add transaction hash generation and tracking
- Create transaction status monitoring
- Implement retry logic for failed broadcasts

#### 4.4 Implement Balance Checking
- Create balance query functionality
- Add pending transaction balance adjustments
- Implement balance caching for performance
- Create insufficient balance error handling

#### 4.5 Add Automatic Transfer Type Detection
- Implement chain configuration analysis
- Create logic to detect native vs ERC20 chains
- Add automatic routing between transfer types
- Implement user notification for automatic switching

#### 5.1 Create ERC20 Contract Interaction Module
- Implement ERC20 ABI interface
- Create contract call builders
- Add token decimal handling
- Implement gas estimation for ERC20 transfers

#### 5.2 Implement Token Transfer Building
- Create ERC20 transfer transaction builders
- Implement token amount conversion with decimals
- Add approval flow for token transfers (if needed)
- Create token transfer validation

#### 5.3 Add Token Balance Checking
- Implement ERC20 balanceOf calls
- Add token balance formatting with decimals
- Create multi-token balance queries
- Implement token balance caching

#### 5.4 Create Transfer Token Command Handler
- Implement `/wallet transfer_token` command
- Add argument parsing for token transfers
- Create token transfer confirmation flows
- Add token transfer error handling

### Phase 3: Advanced Features (Tasks 6-7)
**Objective**: Add advanced wallet features and multi-chain support

#### 6.1 Implement Enhanced Balance Queries
- Create multi-chain balance aggregation
- Implement historical balance tracking
- Add balance change notifications
- Create balance export functionality

#### 6.2 Implement Address Display Features
- Add address QR code generation
- Implement address book functionality
- Create address validation and verification
- Add address labeling and management

#### 6.3 Add Transaction History
- Implement transaction history retrieval
- Create transaction filtering and search
- Add transaction categorization
- Implement transaction export functionality

#### 6.4 Create Formatted Output
- Implement pretty-printing for wallet information
- Add customizable output formats (JSON, table, etc.)
- Create transaction summary displays
- Implement balance trend visualizations

#### 6.5 Add Chain-Specific Balance Display
- Implement chain-aware balance formatting
- Add multi-currency balance display
- Create chain-specific token symbols
- Implement real-time price conversions

#### 7.1 Create Chain Configuration Parser
- Implement configuration file parsing
- Add validation for chain configurations
- Create chain configuration hot-reloading
- Implement configuration migration tools

#### 7.2 Implement Chain Detection Logic
- Create automatic chain detection from RPC
- Implement chain ID verification
- Add chain fork detection and handling
- Create chain health monitoring

#### 7.3 Add Advanced Configuration Fields
- Implement support for `is_native`, `gas_token`, `decimal` fields
- Add custom gas price configurations
- Implement chain-specific transaction parameters
- Create configuration inheritance and overrides

#### 7.4 Create ClawSwift Configuration Example
- Document ClawSwift chain setup process
- Create example configuration files
- Add ClawSwift-specific optimizations
- Implement ClawSwift network monitoring

#### 7.5 Add Channel Configuration Parsing
- Implement channel-specific configuration parsing
- Add `allow_from` list management
- Create channel configuration validation
- Implement dynamic channel configuration updates

#### 7.6 Implement Channel-Specific Authorization
- Create per-channel authorization logic
- Implement user ID mapping for different channels
- Add authorization caching for performance
- Create authorization audit logging

### Phase 4: Error Handling and Testing (Tasks 8-9)
**Objective**: Implement comprehensive error handling and testing

#### 8.1 Implement Input Validation
- Create Ethereum address validation
- Implement amount validation with decimals
- Add password strength validation
- Create transaction parameter validation

#### 8.2 Add Network Error Handling
- Implement RPC connection error handling
- Add network timeout and retry logic
- Create graceful degradation for network issues
- Implement offline mode functionality

#### 8.3 Create User-Friendly Error Messages
- Design clear error message templates
- Implement error code system
- Add error context and suggestions
- Create multi-language error support

#### 8.4 Add Comprehensive Logging
- Implement structured logging for all operations
- Add debug, info, warning, and error log levels
- Create log rotation and management
- Implement log analysis and monitoring

#### 8.5 Add ERC20-Specific Error Handling
- Implement token contract error parsing
- Add insufficient allowance error handling
- Create token-specific error messages
- Implement token transfer failure analysis

#### 8.6 Add Access Control Error Handling
- Create unauthorized access error messages
- Implement user-friendly access denied messages
- Add audit logging for access attempts
- Create admin notification system for violations

#### 9.1 Create Unit Tests for Wallet Creation
- Test wallet generation and keystore creation
- Test password encryption and decryption
- Test error handling for invalid inputs
- Test wallet file management

#### 9.2 Create Unit Tests for ETH Transfers
- Test transaction building and signing
- Test balance checking and validation
- Test error handling for insufficient funds
- Test network error scenarios

#### 9.3 Create Unit Tests for ERC20 Transfers
- Test token transfer transaction building
- Test token balance queries
- Test token decimal handling
- Test contract interaction errors

#### 9.4 Create Integration Tests
- Test full wallet lifecycle operations
- Test multi-chain configuration switching
- Test real blockchain interactions (testnet)
- Test concurrent wallet operations

#### 9.5 Test Error Scenarios
- Test network failure scenarios
- Test invalid input handling
- Test authentication and authorization failures
- Test blockchain-specific error conditions

#### 9.6 Test ClawSwift Chain Configuration
- Test ClawSwift network connectivity
- Test CLAW token transfers
- Test ClawSwift-specific error conditions
- Test ClawSwift block confirmation logic

#### 9.7 Test Access Control Functionality
- Test authorized user access
- Test unauthorized user blocking
- Test channel-specific authorization
- Test dynamic authorization updates

#### 9.8 Test User Authorization Scenarios
- Test different channel types (Telegram, Discord, etc.)
- Test user ID format variations
- Test authorization caching behavior
- Test authorization bypass scenarios

## Implementation Timeline

### Week 1: Foundation (Phase 1)
- Complete dependency setup and basic infrastructure
- Implement basic command structure and routing
- Add access control framework

### Week 2: Core Functionality (Phase 2 - Part 1)
- Complete wallet creation functionality
- Implement native ETH transfers
- Add basic error handling

### Week 3: Advanced Features (Phase 2 - Part 2)
- Complete ERC20 token transfer functionality
- Implement multi-chain configuration
- Add comprehensive balance and info features

### Week 4: Polish and Testing (Phases 3-4)
- Complete error handling and logging
- Implement comprehensive test suite
- Add documentation and examples
- Perform integration testing

## Risk Mitigation

### Technical Risks
1. **Go-Ethereum Compatibility**: Test dependency compatibility early
2. **Network Reliability**: Implement robust retry and fallback mechanisms
3. **Security Vulnerabilities**: Conduct security audits for key management
4. **Performance Issues**: Implement caching and optimization strategies

### Operational Risks
1. **User Experience**: Provide clear error messages and guidance
2. **Blockchain Changes**: Design for configurability and adaptability
3. **Regulatory Compliance**: Ensure proper access controls and audit trails
4. **Maintenance Burden**: Create comprehensive documentation and monitoring

## Success Criteria

### Functional Requirements
- All wallet commands work as specified
- Multi-chain support including ClawSwift
- Access control prevents unauthorized usage
- Error handling provides clear user feedback

### Performance Requirements
- Wallet operations complete within reasonable time limits
- System handles concurrent users efficiently
- Network failures don't cause system instability

### Security Requirements
- Private keys are never exposed in plain text
- Access control prevents unauthorized transactions
- Audit logs capture all sensitive operations
- System follows security best practices

### Quality Requirements
- Comprehensive test coverage (>80%)
- Clear documentation for users and developers
- Consistent error handling and logging
- Maintainable and extensible code structure

This implementation plan provides a comprehensive roadmap for adding Ethereum wallet functionality to Picoclaw while maintaining security, reliability, and user experience standards.
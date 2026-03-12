# Minimal Core Wallet Integration Spec

## Why
To add Ethereum wallet functionality to Picoclaw while maintaining minimal impact on the core codebase for easy future syncing with the main repository stream.

## Design Principles
- **Minimal Core Changes**: Modify core Picoclaw files as little as possible
- **Plugin Architecture**: Use existing extension points and patterns
- **Modular Design**: Self-contained wallet functionality
- **Configuration-Driven**: Enable/disable features without code changes
- **Backward Compatible**: Don't break existing functionality

## What Changes
- Add wallet functionality as optional plugin/module
- Create minimal command registration hooks
- Implement configuration-based feature toggles
- Add wallet-specific configuration section
- Use existing channel and command infrastructure
- Implement access control through existing auth systems

## Impact
- **Minimal Core Impact**: Only essential integration points modified
- **Plugin Architecture**: Wallet features isolated in separate modules
- **Configuration Driven**: Features can be enabled/disabled without code changes
- **Easy Maintenance**: Future syncs won't conflict with wallet functionality

## ADDED Requirements

### Requirement: Plugin-Based Wallet Architecture
The system SHALL provide wallet functionality as an optional plugin that integrates with existing Picoclaw infrastructure.

#### Scenario: Wallet plugin initialization
- **WHEN** system starts with wallet plugin enabled in configuration
- **THEN** wallet commands are registered through existing command system
- **AND** wallet functionality is available without modifying core files
- **AND** plugin can be disabled without affecting core functionality

### Requirement: Minimal Command Registration
The system SHALL register wallet commands using existing command registration patterns without modifying core command processing.

#### Scenario: Register wallet commands
- **WHEN** wallet plugin is loaded
- **THEN** commands are registered through existing registry system
- **AND** no core command processing logic is modified
- **AND** commands follow existing argument parsing patterns

### Requirement: Configuration-Driven Feature Toggle
The system SHALL allow wallet functionality to be enabled/disabled through configuration without code changes.

#### Scenario: Disable wallet functionality
- **WHEN** wallet.enabled is set to false in configuration
- **THEN** wallet commands are not registered
- **AND** wallet-related code is not loaded
- **AND** system operates normally without wallet features

### Requirement: Channel-Specific Access Control
The system SHALL use existing channel access control mechanisms for wallet command authorization.

#### Scenario: Channel access control integration
- **WHEN** user attempts to use wallet commands
- **THEN** existing channel authorization is checked
- **AND** wallet-specific allow_from lists are used if configured
- **AND** existing access control patterns are followed

### Requirement: Modular Blockchain Configuration
The system SHALL support multiple blockchain configurations through a dedicated wallet configuration section.

#### Scenario: Configure ClawSwift chain
- **WHEN** wallet.chains configuration includes ClawSwift
- **THEN** ClawSwift chain is available for wallet operations
- **AND** ERC20 token transfers are supported for non-native chains
- **AND** chain switching is handled automatically

## Configuration Example
```json
{
  "wallet": {
    "enabled": true,
    "chains": [
      {
        "name": "Ethereum",
        "chain_id": 1,
        "rpc": "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
        "explorer": "https://etherscan.io",
        "currency": "ETH",
        "is_native": true,
        "decimal": 18
      },
      {
        "name": "ClawSwift",
        "chain_id": 7441,
        "rpc": "https://exp.clawswift.net/rpc",
        "explorer": "https://exp.clawswift.net",
        "currency": "CLAW",
        "is_native": false,
        "gas_token": "0x20c0000000000000000000000000000000000000",
        "gas_token_name": "CLAW",
        "decimal": 16
      }
    ]
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "allow_from": ["user123", "admin456"]
    }
  }
}
```

## Integration Strategy

### Minimal Core Modifications
1. **Command Registration**: Add wallet commands through existing registry hooks
2. **Configuration Loading**: Extend existing config system with wallet section
3. **Channel Integration**: Use existing channel infrastructure for access control
4. **Error Handling**: Leverage existing error handling patterns

### Plugin Architecture Benefits
- **Isolation**: Wallet functionality is self-contained
- **Maintainability**: Easy to update without affecting core
- **Testability**: Can be tested independently
- **Flexibility**: Easy to add new chains or features
- **Sync Compatibility**: Future core updates won't conflict

## Security Considerations
- Private keys stored securely using existing encryption patterns
- Access control integrated with existing authorization systems
- Audit logging follows existing logging infrastructure
- Network security uses existing connection patterns
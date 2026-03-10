# Picoclaw Wallet Integration

This package provides Ethereum wallet functionality for Picoclaw, including support for both native ETH transfers and ERC20 token transfers, with a focus on minimal core impact and easy maintenance.

## Features

- **Wallet Creation**: Create Ethereum wallets with password-protected JSON keystore format
- **Native ETH Transfers**: Transfer ETH on native chains like Ethereum mainnet
- **ERC20 Token Transfers**: Transfer ERC20 tokens, including gas tokens on non-native chains
- **Multi-Chain Support**: Support for multiple blockchain networks including ClawSwift
- **Access Control**: Channel-specific access control using existing Picoclaw infrastructure
- **Minimal Core Impact**: Plugin architecture that doesn't modify core Picoclaw functionality

## Architecture

### Plugin-Based Design

The wallet functionality is implemented as an isolated plugin that integrates with existing Picoclaw infrastructure:

```
pkg/wallet/
├── config.go          # Wallet configuration structures
├── service.go         # Core wallet service implementation
├── types.go           # Type definitions
├── erc20.go           # ERC20 contract interaction utilities
├── errors.go          # Error handling and user-friendly messages
└── commands/          # CLI command implementations
    ├── create.go      # Wallet creation command
    ├── transfer.go    # ETH/ERC20 transfer command
    ├── transfer_token.go # Explicit ERC20 transfer command
    └── info.go        # Wallet information command
```

### Integration Points

1. **Command Registration**: Added to main command registry following existing patterns
2. **Configuration**: Extended existing config system with wallet section
3. **Access Control**: Uses existing channel `allow_from` mechanisms
4. **Error Handling**: Follows established Picoclaw error patterns

## Configuration

Add the following to your Picoclaw configuration file:

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
  }
}
```

### Configuration Fields

- **enabled**: Enable/disable wallet functionality
- **chains**: Array of blockchain configurations
  - **name**: Chain name for display
  - **chain_id**: Unique chain identifier
  - **rpc**: RPC endpoint URL
  - **explorer**: Block explorer URL
  - **currency**: Native currency symbol
  - **is_native**: Whether the chain has native gas token (ETH)
  - **gas_token**: ERC20 contract address for gas token (required if `is_native: false`)
  - **gas_token_name**: Display name for the gas token
  - **decimal**: Token decimal places (18 for ETH, varies for ERC20)

## Usage

### Create Wallet

```bash
# Create a new wallet
picoclaw wallet create mypassword

# Create wallet for specific chain
picoclaw wallet create mypassword --chain-id 7441
```

### Transfer Funds

```bash
# Transfer ETH on Ethereum
picoclaw wallet transfer 0x123... 0x456... 1.5 mypassword

# Transfer on specific chain (automatically uses ERC20 for non-native chains)
picoclaw wallet transfer 0x123... 0x456... 100 mypassword --chain-id 7441

# Explicit ERC20 token transfer
picoclaw wallet transfer_token 0x123... 0x456... 50 mypassword --chain-id 7441 --token 0x20c0000000000000000000000000000000000000
```

### Check Wallet Info

```bash
# Show info for first wallet
picoclaw wallet info

# Show info for specific address
picoclaw wallet info --address 0x123...

# Show info for specific chain
picoclaw wallet info --address 0x123... --chain-id 7441
```

## Access Control

Wallet commands respect channel-specific `allow_from` configuration:

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "allow_from": ["user123", "admin456"]
    }
  }
}
```

Only users in the `allow_from` list can use wallet commands. If no `allow_from` is configured, all users can access wallet functionality (backward compatibility).

## Security Considerations

1. **Private Key Storage**: Private keys are stored as encrypted JSON keystore files using industry-standard encryption
2. **Password Protection**: All wallet operations require password authentication
3. **Access Control**: Channel-specific authorization prevents unauthorized access
4. **Network Security**: Uses secure RPC connections for blockchain interactions
5. **Error Sanitization**: User-friendly error messages don't expose sensitive information

## Development

### Adding New Chains

To add support for a new blockchain:

1. Add chain configuration to your config file
2. Ensure the chain supports Ethereum-compatible JSON-RPC
3. For ERC20-gas chains, specify the gas token contract address

### Testing

Run the test suite:

```bash
go test ./pkg/wallet -v
```

### Error Handling

The wallet system provides comprehensive error handling with user-friendly messages:

- **Invalid Configuration**: Clear validation messages for configuration issues
- **Network Errors**: Retry logic and connection error handling
- **Access Denied**: Proper authorization error messages
- **Transaction Failures**: Detailed transaction error information

## Future Enhancements

Potential improvements that maintain the minimal-core philosophy:

1. **Hardware Wallet Support**: Integration with Ledger/Trezor devices
2. **Multi-Signature Wallets**: Support for multi-sig transactions
3. **Transaction History**: Enhanced transaction tracking and history
4. **Price Feeds**: Integration with price oracles for balance display
5. **Custom Token Support**: User-defined ERC20 token management

## Maintenance and Updates

The plugin architecture ensures easy maintenance:

- **Isolated Code**: All wallet functionality is contained in `pkg/wallet/`
- **Configuration Driven**: Features can be enabled/disabled without code changes
- **Existing Patterns**: Uses established Picoclaw patterns for consistency
- **Minimal Core Impact**: Future core updates won't conflict with wallet functionality

## Troubleshooting

### Common Issues

1. **"Wallet functionality is disabled"**: Check that `wallet.enabled: true` in configuration
2. **"Chain not found"**: Verify chain ID exists in configuration
3. **"Account not found"**: Ensure wallet exists in keystore directory
4. **"Insufficient balance"**: Check balance before attempting transfers
5. **"Access denied"**: Verify user ID is in channel's `allow_from` list

### Debug Mode

Enable debug logging to troubleshoot issues:

```json
{
  "log_level": "debug"
}
```

## License

This wallet integration follows the same license as the main Picoclaw project.
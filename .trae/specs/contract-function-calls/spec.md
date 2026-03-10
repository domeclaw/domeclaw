# Contract Function Calls Spec

## Why
Need to add support for calling and executing smart contract functions from Picoclaw using Go Ethereum (geth) via RPC, as required by the dc-hotwallet skill. This will allow users to interact with smart contracts directly through Telegram commands.

## What Changes
- Add `query_contract_call` tool for reading contract functions
- Add `execute_contract_write` tool for writing to contract functions
- Add `/wallet call` command for contract read operations
- Add `/wallet write` command for contract write operations
- Implement support for common ERC20 methods (balanceOf, transfer, etc.)
- Create ABI management system to support contract interactions

## Impact
- Affected specs: minimal-core-wallet-integration
- Affected code:
  - /pkg/wallet/ - Add contract interaction methods
  - /pkg/tools/ - Add query_contract_call and execute_contract_write tools
  - /pkg/commands/ - Add /wallet call and /wallet write commands
  - /cmd/picoclaw/internal/wallet/ - Add CLI commands

## ADDED Requirements
### Requirement: Query Contract Call
The system SHALL provide a tool and command to call read-only smart contract functions.

#### Scenario: Call ERC20 balanceOf
- **WHEN** user calls `/wallet call 0x20c0000000000000000000000000000000000000 erc20 balanceOf 0x44c2db1fc0986ca3c173403701c909874badc0d0`
- **THEN** system returns the token balance of the address

### Requirement: Execute Contract Write
The system SHALL provide a tool and command to execute write smart contract functions.

#### Scenario: Execute ERC20 transfer
- **WHEN** user calls `/wallet write 0x20c0000000000000000000000000000000000000 erc20 transfer 0 1234 0xRecipientAddress 1000000000000000000`
- **THEN** system executes the transfer and returns transaction hash

## MODIFIED Requirements
### Requirement: Wallet Service
The wallet service SHALL be modified to support contract interactions.
- Add method to call contract read functions
- Add method to execute contract write functions
- Add ABI management system for common contract types

## REMOVED Requirements
None

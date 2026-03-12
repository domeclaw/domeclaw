# Tasks

## Phase 1: Core Contract Interaction Methods
- [x] Task 1: Add contract read function call method to wallet service
  - [x] SubTask 1.1: Implement CallContractMethod in /pkg/wallet/service.go
  - [x] SubTask 1.2: Handle ABI lookup for standard contract types (erc20)
  - [x] SubTask 1.3: Implement parameter encoding/decoding
- [x] Task 2: Add contract write function execution method to wallet service
  - [x] SubTask 2.1: Implement ExecuteContractMethod in /pkg/wallet/service.go
  - [x] SubTask 2.2: Handle gas estimation for write operations
  - [x] SubTask 2.3: Add transaction signing and sending logic

## Phase 2: Tools for AI Integration
- [x] Task 3: Create query_contract_call tool
  - [x] SubTask 3.1: Implement tool in /pkg/tools/wallet.go
  - [x] SubTask 3.2: Define tool schema and parameters
  - [x] SubTask 3.3: Implement tool execution logic
- [x] Task 4: Create execute_contract_write tool
  - [x] SubTask 4.1: Implement tool in /pkg/tools/wallet.go
  - [x] SubTask 4.2: Define tool schema and parameters
  - [x] SubTask 4.3: Implement tool execution logic

## Phase 3: Telegram Commands
- [x] Task 5: Add /wallet call command to commands package
  - [x] SubTask 5.1: Implement handleWalletCall in /pkg/commands/cmd_wallet.go
  - [x] SubTask 5.2: Parse command arguments
  - [x] SubTask 5.3: Handle response formatting
- [x] Task 6: Add /wallet write command to commands package
  - [x] SubTask 6.1: Implement handleWalletWrite in /pkg/commands/cmd_wallet.go
  - [x] SubTask 6.2: Parse command arguments
  - [x] SubTask 6.3: Handle response formatting

## Phase 4: CLI Commands
- [x] Task 7: Add CLI commands for contract calls
  - [x] SubTask 7.1: Implement call.go in /cmd/picoclaw/internal/wallet/
  - [x] SubTask 7.2: Implement write.go in /cmd/picoclaw/internal/wallet/
  - [x] SubTask 7.3: Add commands to main wallet command

## Phase 5: Testing and Verification
- [ ] Task 8: Write unit tests for contract interaction methods
- [ ] Task 9: Test with actual contract on ClawSwift network
- [ ] Task 10: Verify integration with dc-hotwallet skill

# Task Dependencies
- [Task 3] and [Task 4] depend on [Task 1] and [Task 2]
- [Task 5] and [Task 6] depend on [Task 3] and [Task 4]
- [Task 7] depends on [Task 5] and [Task 6]
- [Task 8], [Task 9], and [Task 10] depend on all other tasks

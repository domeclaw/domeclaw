package channels

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/sipeed/domeclaw/pkg/wallet"
)

// WalletCommander handles wallet-related commands
type WalletCommander interface {
	Create(ctx context.Context, message telego.Message, pin string) error
	Info(ctx context.Context, message telego.Message) error
	Balance(ctx context.Context, message telego.Message, tokenAddress string) error
	Transfer(ctx context.Context, message telego.Message, args []string) error
	TransferToken(ctx context.Context, message telego.Message, args []string) error
	CallContract(ctx context.Context, message telego.Message, args []string) error
	WriteContract(ctx context.Context, message telego.Message, args []string) error
	UploadABI(ctx context.Context, message telego.Message, name, abiJSON string) error
	ListABIs(ctx context.Context, message telego.Message) error
}

type walletCmd struct {
	walletService *wallet.WalletService
	bot           *telego.Bot
}

// NewWalletCommands creates wallet command handler
func NewWalletCommands(ws *wallet.WalletService, bot *telego.Bot) WalletCommander {
	return &walletCmd{
		walletService: ws,
		bot:           bot,
	}
}

// Create handles wallet creation
func (wc *walletCmd) Create(ctx context.Context, message telego.Message, pin string) error {
	// Check if wallet already exists
	if wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID),
			"❌ Wallet already exists!\n\nUse /wallet info to view your wallet."))
		return err
	}

	// If no PIN provided, ask user to set one
	if pin == "" {
		_, err := wc.bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID),
			"🔐 Please set a 4-digit PIN for your wallet.\n\n"+
				"Usage: `/wallet create 1234`\n\n"+
				"⚠️ Keep your PIN safe! You'll need it for transactions."))
		return err
	}

	// Create wallet
	address, err := wc.walletService.CreateWallet(pin)
	if err != nil {
		msg := fmt.Sprintf("❌ Failed to create wallet: %v", err)
		_, sendErr := wc.bot.SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID), msg))
		return sendErr
	}

	response := fmt.Sprintf(
		"✅ Wallet created successfully!\n\n"+
			"📍 Address: `%s`\n"+
			"🔐 PIN: Set\n\n"+
			"⚠️ **Important**: This is a hot wallet for convenient use via Telegram.\n"+
			"Keep your PIN safe - you'll need it for all transactions!\n\n"+
			"Use `/wallet info` to view your wallet details.",
		address.Hex(),
	)

	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      response,
		ParseMode: "Markdown",
	})
	return err
}

// Info displays wallet information
func (wc *walletCmd) Info(ctx context.Context, message telego.Message) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	address, err := wc.walletService.GetAddress()
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Error: %v", err),
		})
		return sendErr
	}

	balance, _ := wc.walletService.GetBalance()

	response := fmt.Sprintf(
		"🦐 **DomeClaw Wallet**\n\n"+
			"📍 Address: `%s`\n"+
			"💰 Balance: %s CLAW\n"+
			"🔗 Chain: ClawSwift (7441)\n\n"+
			"🔒 Status: Locked\n\n"+
			"Use `/wallet unlock [PIN]` to unlock for transactions.",
		address.Hex(),
		balance,
	)

	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      response,
		ParseMode: "Markdown",
	})
	return err
}

// Balance displays token balance
func (wc *walletCmd) Balance(ctx context.Context, message telego.Message, tokenAddress string) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	// Default token address if not provided
	if tokenAddress == "" {
		tokenAddress = "0x20c0000000000000000000000000000000000000"
	}

	// Validate address format
	if len(tokenAddress) != 42 || tokenAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid token address format.\n\nPlease provide a valid Ethereum address (0x...).",
		})
		return err
	}

	walletAddress, err := wc.walletService.GetAddress()
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Error: %v", err),
		})
		return sendErr
	}

	// Get token balance
	balanceInfo, err := wc.walletService.GetTokenBalance(tokenAddress)
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Failed to get token balance: %v", err),
		})
		return sendErr
	}

	response := fmt.Sprintf(
		"💰 **Token Balance**\n\n"+
			"👛 Wallet: `%s`\n"+
			"🪙 Token: `%s`\n"+
			"🏷️ Symbol: **%s**\n"+
			"📊 Decimals: %d\n\n"+
			"💵 Balance: `%s %s`",
		walletAddress.Hex(),
		balanceInfo.Address,
		balanceInfo.Symbol,
		balanceInfo.Decimals,
		balanceInfo.Balance,
		balanceInfo.Symbol,
	)

	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      response,
		ParseMode: "Markdown",
	})
	return err
}

// TransferNative sends native/default chain token
func (wc *walletCmd) Transfer(ctx context.Context, message telego.Message, args []string) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	// Args: [to_address, amount, pin]
	if len(args) != 3 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text: "❌ Invalid arguments.\n\n" +
				"Usage: `/wallet transfer <to_address> <amount> <pin>`\n\n" +
				"Example: `/wallet transfer 0xABC... 100 1234`",
			ParseMode: "Markdown",
		})
		return err
	}

	toAddress := args[0]
	amountStr := args[1]
	pin := args[2]

	// Validate address
	if len(toAddress) != 42 || toAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid recipient address format. Must be 42 chars starting with 0x.",
		})
		return err
	}

	return wc.executeTransfer(ctx, message, "", toAddress, amountStr, pin)
}

// TransferToken sends ERC20 tokens
func (wc *walletCmd) TransferToken(ctx context.Context, message telego.Message, args []string) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	// Args: [token_address, to_address, amount, pin]
	if len(args) != 4 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text: "❌ Invalid arguments.\n\n" +
				"Usage: `/wallet transfertoken <token_address> <to_address> <amount> <pin>`\n\n" +
				"Example: `/wallet transfertoken 0xTOKEN... 0xABC... 100 1234`",
			ParseMode: "Markdown",
		})
		return err
	}

	tokenAddress := args[0]
	toAddress := args[1]
	amountStr := args[2]
	pin := args[3]

	// Validate addresses
	if len(tokenAddress) != 42 || tokenAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid token address format. Must be 42 chars starting with 0x.",
		})
		return err
	}

	if len(toAddress) != 42 || toAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid recipient address format. Must be 42 chars starting with 0x.",
		})
		return err
	}

	return wc.executeTransfer(ctx, message, tokenAddress, toAddress, amountStr, pin)
}

// executeTransfer performs the actual transfer
func (wc *walletCmd) executeTransfer(ctx context.Context, message telego.Message, tokenAddress, toAddress, amountStr, pin string) error {
	// Parse amount
	amountFloat := new(big.Float)
	_, ok := amountFloat.SetString(amountStr)
	if !ok {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid amount format.",
		})
		return err
	}

	// Get token decimals for proper conversion
	var decimals int32 = 18 // default
	if tokenAddress != "" {
		tokenInfo, err := wc.walletService.GetTokenBalance(tokenAddress)
		if err == nil && tokenInfo != nil {
			decimals = tokenInfo.Decimals
		}
	}

	// Convert amount to wei/smallest unit
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amount := new(big.Float).Mul(amountFloat, new(big.Float).SetInt(multiplier))
	amountInt, _ := amount.Int(nil)

	if amountInt.Sign() <= 0 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Amount must be greater than 0.",
		})
		return err
	}

	// Send confirmation message
	walletAddr, _ := wc.walletService.GetAddress()
	tokenDisplay := "CLAW (default)"
	if tokenAddress != "" {
		tokenDisplay = tokenAddress[:6] + "..." + tokenAddress[38:]
	}

	_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text: fmt.Sprintf(
			"🔄 **Sending Transaction**\n\n"+
				"From: `%s`\n"+
				"To: `%s`\n"+
				"Amount: %s\n"+
				"Token: %s\n\n"+
				"Processing...",
			walletAddr.Hex(),
			toAddress,
			amountStr,
			tokenDisplay,
		),
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}

	// Perform transfer
	to := common.HexToAddress(toAddress)
	var txHash common.Hash
	var txErr error

	if tokenAddress == "" {
		// Use default token transfer
		txHash, txErr = wc.walletService.Transfer(to, amountInt, pin)
	} else {
		// Use specific token transfer
		tokenAddr := common.HexToAddress(tokenAddress)
		txHash, txErr = wc.walletService.TransferToken(tokenAddr, to, amountInt, pin)
	}

	if txErr != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Transfer failed: %v", txErr),
		})
		return sendErr
	}

	// Success response
	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      fmt.Sprintf("✅ **Transfer Successful!**\n\n📤 Transaction Hash:\n`%s`", txHash.Hex()),
		ParseMode: "Markdown",
	})
	return err
}

// ListABIs lists all uploaded ABIs
func (wc *walletCmd) ListABIs(ctx context.Context, message telego.Message) error {
	abis, err := wc.walletService.ListABIs()
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Failed to list ABIs: %v", err),
		})
		return sendErr
	}

	if len(abis) == 0 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text: "📋 **No ABIs uploaded**\n\n" +
				"Use `/wallet abiupload <name>` with a JSON file reply to upload ABI.",
			ParseMode: "Markdown",
		})
		return err
	}

	var list strings.Builder
	list.WriteString("📋 **Available ABIs:**\n\n")
	for i, abi := range abis {
		list.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, abi))
	}

	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      list.String(),
		ParseMode: "Markdown",
	})
	return err
}

// UploadABI uploads an ABI from a JSON file
func (wc *walletCmd) UploadABI(ctx context.Context, message telego.Message, name, abiJSON string) error {
	if name == "" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Please provide ABI name.\n\nUsage: Reply to a JSON file with `/wallet abiupload <name>`",
		})
		return err
	}

	if abiJSON == "" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Please reply to a JSON file containing the ABI.",
		})
		return err
	}

	if err := wc.walletService.UploadABI(name, abiJSON); err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Failed to upload ABI: %v", err),
		})
		return sendErr
	}

	_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      fmt.Sprintf("✅ **ABI uploaded successfully!**\n\nName: `%s`", name),
		ParseMode: "Markdown",
	})
	return err
}

// CallContract calls a read-only contract function
func (wc *walletCmd) CallContract(ctx context.Context, message telego.Message, args []string) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	// Args: <contract_address> <abi_name> <method> [arg1 arg2 ...]
	if len(args) < 3 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text: "❌ Invalid arguments.\n\n" +
				"Usage: `/wallet call <contract_address> <abi_name> <method> [args...]`\n\n" +
				"Example:\n" +
				"`/wallet call 0xContract... erc20 balanceOf 0xWallet...`",
			ParseMode: "Markdown",
		})
		return err
	}

	contractAddress := args[0]
	abiName := args[1]
	method := args[2]
	var callArgs []interface{}

	// Parse remaining args as strings (simple types only for now)
	for i := 3; i < len(args); i++ {
		arg := args[i]
		// Try to parse as number first
		if num, ok := new(big.Int).SetString(arg, 10); ok {
			callArgs = append(callArgs, num)
		} else if len(arg) == 42 && arg[:2] == "0x" {
			// Address
			callArgs = append(callArgs, common.HexToAddress(arg))
		} else {
			// String
			callArgs = append(callArgs, arg)
		}
	}

	// Validate address
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid contract address format.",
		})
		return err
	}

	contract := common.HexToAddress(contractAddress)

	// Call contract
	result, err := wc.walletService.CallContract(contract, abiName, method, callArgs)
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Call failed: %v", err),
		})
		return sendErr
	}

	// Format result
	resultStr := fmt.Sprintf("%v", result)
	if result == nil {
		resultStr = "(no return value)"
	}

	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      fmt.Sprintf("📤 **Contract Call Result**\n\nContract: `%s`\nMethod: `%s`\n\nResult: `%s`", contractAddress, method, resultStr),
		ParseMode: "Markdown",
	})
	return err
}

// WriteContract calls a state-changing contract function
func (wc *walletCmd) WriteContract(ctx context.Context, message telego.Message, args []string) error {
	if !wc.walletService.WalletExists() {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ No wallet found.\n\nUse `/wallet create [PIN]` to create one.",
		})
		return err
	}

	// Args: <contract_address> <abi_name> <method> <value> <pin> [arg1 arg2 ...]
	if len(args) < 5 {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text: "❌ Invalid arguments.\n\n" +
				"Usage: `/wallet write <contract> <abi> <method> <value> <pin> [args...]`\n\n" +
				"Example:\n" +
				"`/wallet write 0xContract... erc20 transfer 0 1234 0xTo... 1000`",
			ParseMode: "Markdown",
		})
		return err
	}

	contractAddress := args[0]
	abiName := args[1]
	method := args[2]
	valueStr := args[3]
	pin := args[4]
	var callArgs []interface{}

	// Parse remaining args (from index 5)
	for i := 5; i < len(args); i++ {
		arg := args[i]
		if num, ok := new(big.Int).SetString(arg, 10); ok {
			callArgs = append(callArgs, num)
		} else if len(arg) == 42 && arg[:2] == "0x" {
			callArgs = append(callArgs, common.HexToAddress(arg))
		} else {
			callArgs = append(callArgs, arg)
		}
	}

	// Validate address
	if len(contractAddress) != 42 || contractAddress[:2] != "0x" {
		_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "❌ Invalid contract address format.",
		})
		return err
	}

	// Parse value
	value := big.NewInt(0)
	if valueStr != "0" {
		v, ok := new(big.Int).SetString(valueStr, 10)
		if !ok {
			_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: message.Chat.ID},
				Text:   "❌ Invalid value format. Use 0 for no ETH transfer.",
			})
			return err
		}
		value = v
	}

	contract := common.HexToAddress(contractAddress)

	// Send confirmation
	_, err := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      fmt.Sprintf("🔄 **Writing to Contract**\n\nContract: `%s`\nMethod: `%s`\n\nProcessing...", contractAddress, method),
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}

	// Execute write
	txHash, err := wc.walletService.WriteContract(contract, abiName, method, callArgs, value, pin)
	if err != nil {
		_, sendErr := wc.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   fmt.Sprintf("❌ Write failed: %v", err),
		})
		return sendErr
	}

	// Success response
	_, err = wc.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      fmt.Sprintf("✅ **Transaction Sent!**\n\n📤 Transaction Hash:\n`%s`", txHash.Hex()),
		ParseMode: "Markdown",
	})
	return err
}

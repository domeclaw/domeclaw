package channels

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/sipeed/domeclaw/pkg/wallet"
)

// WalletCommander handles wallet-related commands
type WalletCommander interface {
	Create(ctx context.Context, message telego.Message, pin string) error
	Info(ctx context.Context, message telego.Message) error
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

// commandArgs extracts arguments from command text (already defined in telegram_commands.go)

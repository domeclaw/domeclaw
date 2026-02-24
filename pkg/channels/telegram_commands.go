package channels

import (
	"context"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"

	"github.com/sipeed/domeclaw/pkg/config"
)

type TelegramCommander interface {
	Help(ctx context.Context, message telego.Message) error
	Start(ctx context.Context, message telego.Message) error
	Model(ctx context.Context, message telego.Message) error
	Status(ctx context.Context, message telego.Message) error
	Show(ctx context.Context, message telego.Message) error
	List(ctx context.Context, message telego.Message) error
}

type cmd struct {
	bot    *telego.Bot
	config *config.Config
}

func NewTelegramCommands(bot *telego.Bot, cfg *config.Config) TelegramCommander {
	return &cmd{
		bot:    bot,
		config: cfg,
	}
}

func commandArgs(text string) string {
	parts := strings.SplitN(text, " ", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (c *cmd) Help(ctx context.Context, message telego.Message) error {
	msg := `🐈‍⬛🦐✨ *DomeClaw Bot Commands*

/start - Start the bot
/help - Show this help message
/model - Show current model info
/status - Show bot status and configuration
/wallet create [PIN] - Create Ethereum wallet
/wallet info - View wallet info
/show [model|channel] - Show specific configuration
/list [models|channels] - List available options

*Examples:*
/model - See which AI model is being used
/wallet create 1234 - Create wallet with PIN 1234
/wallet info - View wallet address and balance
/show model - Same as /model
/list models - See all configured models`
	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      msg,
		ParseMode: "Markdown",
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) Start(ctx context.Context, message telego.Message) error {
	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text:   "Hello! I am DomeClaw 🐈‍⬛🦐✨\n\nUse /help to see available commands.",
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) Model(ctx context.Context, message telego.Message) error {
	model := c.config.Agents.Defaults.Model
	provider := c.config.Agents.Defaults.Provider

	// Find model details from model_list
	var modelDetails string
	for _, mc := range c.config.ModelList {
		if mc.ModelName == model || mc.Model == provider+"/"+model {
			if mc.APIBase != "" {
				modelDetails = fmt.Sprintf("\n📡 API: %s", mc.APIBase)
			}
			break
		}
	}

	msg := fmt.Sprintf("🤖 *Current AI Model*\n\n"+
		"*Model:* `%s`%s\n"+
		"*Provider:* `%s`%s",
		model,
		modelDetails,
		provider,
		"",
	)

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      msg,
		ParseMode: "Markdown",
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) Status(ctx context.Context, message telego.Message) error {
	model := c.config.Agents.Defaults.Model
	provider := c.config.Agents.Defaults.Provider
	workspace := c.config.Agents.Defaults.Workspace

	msg := fmt.Sprintf("🐈‍⬛🦐✨ *DomeClaw Status*\n\n"+
		"*Model:* `%s`\n"+
		"*Provider:* `%s`\n"+
		"*Workspace:* `%s`\n"+
		"*Channel:* Telegram",
		model,
		provider,
		workspace,
	)

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      msg,
		ParseMode: "Markdown",
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) Show(ctx context.Context, message telego.Message) error {
	args := commandArgs(message.Text)
	if args == "" {
		_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "Usage: /show [model|channel]\n\nUse /model for detailed model info.",
			ReplyParameters: &telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		})
		return err
	}

	var response string
	switch args {
	case "model":
		response = fmt.Sprintf("🤖 Current Model: `%s`\nProvider: `%s`",
			c.config.Agents.Defaults.Model,
			c.config.Agents.Defaults.Provider)
	case "channel":
		response = "📱 Current Channel: `telegram`"
	default:
		response = fmt.Sprintf("❌ Unknown parameter: %s. Try 'model' or 'channel'.", args)
	}

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    telego.ChatID{ID: message.Chat.ID},
		Text:      response,
		ParseMode: "Markdown",
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

func (c *cmd) List(ctx context.Context, message telego.Message) error {
	args := commandArgs(message.Text)
	if args == "" {
		_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: message.Chat.ID},
			Text:   "Usage: /list [models|channels]",
			ReplyParameters: &telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		})
		return err
	}

	var response string
	switch args {
	case "models":
		provider := c.config.Agents.Defaults.Provider
		if provider == "" {
			provider = "configured default"
		}
		response = fmt.Sprintf("Configured Model: %s\nProvider: %s\n\nTo change models, update config.yaml",
			c.config.Agents.Defaults.Model, provider)

	case "channels":
		var enabled []string
		if c.config.Channels.Telegram.Enabled {
			enabled = append(enabled, "telegram")
		}
		if c.config.Channels.WhatsApp.Enabled {
			enabled = append(enabled, "whatsapp")
		}
		if c.config.Channels.Feishu.Enabled {
			enabled = append(enabled, "feishu")
		}
		if c.config.Channels.Discord.Enabled {
			enabled = append(enabled, "discord")
		}
		if c.config.Channels.Slack.Enabled {
			enabled = append(enabled, "slack")
		}
		response = fmt.Sprintf("Enabled Channels:\n- %s", strings.Join(enabled, "\n- "))

	default:
		response = fmt.Sprintf("Unknown parameter: %s. Try 'models' or 'channels'.", args)
	}

	_, err := c.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: message.Chat.ID},
		Text:   response,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	return err
}

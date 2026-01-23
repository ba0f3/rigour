package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Bot struct {
	Token  string
	ChatID int64
}

func NewBot(token string, chatID int64) *Bot {
	return &Bot{Token: token, ChatID: chatID}
}

func (b *Bot) Notify(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.Token)

	payload := map[string]interface{}{
		"chat_id":    b.ChatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api returned status: %s", resp.Status)
	}

	return nil
}

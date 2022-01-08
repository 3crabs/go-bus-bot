package tg

import (
	"fmt"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type tg struct {
	bot *tgbot.BotAPI
}

func NewTg(bot *tgbot.BotAPI) *tg {
	return &tg{bot: bot}
}

func (t *tg) SendPage(chatId int64, title, description, action string, keyboard [][]tgbot.InlineKeyboardButton) {
	text := ""
	if description != "" {
		text = fmt.Sprintf("%s\n\n%s\n\n%s", title, description, action)
	} else {
		text = fmt.Sprintf("%s\n\n%s", title, action)
	}
	msg := tgbot.NewMessage(chatId, text)
	if keyboard != nil {
		msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
	}
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
}

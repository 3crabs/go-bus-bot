package tg

import (
	"fmt"
	"github.com/3crabs/go-bus-bot/nav"
	"github.com/3crabs/go-bus-bot/user"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

type tg struct {
	bot *tgbot.BotAPI
}

func NewTg(bot *tgbot.BotAPI) *tg {
	return &tg{bot: bot}
}

func (t *tg) SendPageError(chatId int64, userMessageId int, description, action string) {
	t.SendPage(chatId, userMessageId, "Ошибка", description, action, nil)
}

var messageId = 0

func (t *tg) SendPage(chatId int64, userMessageId int, title, description, action string, keyboard [][]tgbot.InlineKeyboardButton) {
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
	if messageId != 0 {
		_, _ = t.bot.DeleteMessage(tgbot.DeleteMessageConfig{
			ChatID:    chatId,
			MessageID: messageId,
		})
		_, _ = t.bot.DeleteMessage(tgbot.DeleteMessageConfig{
			ChatID:    chatId,
			MessageID: userMessageId,
		})
	}
	m, err := t.bot.Send(msg)
	messageId = m.MessageID
	if err != nil {
		log.Println(err)
	}
}

func (t *tg) ShowPageMain(chatId int64, userMessageId int, u *user.User) {
	keyboard := [][]tgbot.InlineKeyboardButton{
		{{Text: "Рейсы", CallbackData: nav.PageFindRaces.Link()}},
	}
	description := ""
	if u.Login {
		description = "Сейчас вам доступны все функции"
		keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Пассажиры", CallbackData: nav.PagePassengers.Link()}})
		keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Отзывы", CallbackData: nav.PageFeedback.Link()}})
	}
	if !u.Login {
		description = "Сейчас вы можете только:\n- смотреть рейсы\n\nДля получения доступа ко всем функциям нужно войти"
		keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Вход", CallbackData: nav.PageLogin.Link()}})
	}
	t.SendPage(
		chatId,
		userMessageId,
		"Главная",
		description,
		"Меню:",
		keyboard,
	)
	u.SetState(nav.Menu)
}

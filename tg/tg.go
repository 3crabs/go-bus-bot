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

func (t *tg) SendPageError(chatId int64, description, action string) {
	t.SendPage(chatId, "Ошибка", description, action, nil)
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

func (t *tg) ShowPageMain(chatId int64, u *user.User) {
	keyboard := [][]tgbot.InlineKeyboardButton{
		{{Text: "Рейсы", CallbackData: nav.PageFindRaces.Link()}},
	}
	description := ""
	if u.Login {
		description = "Сейчас вам доступны все функции"
		keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Пассажиры", CallbackData: nav.PagePassengers.Link()}})
	}
	if !u.Login {
		description = "Сейчас вы можете только:\n- смотреть рейсы\n\nДля получения доступа ко всем функциям нужно войти"
		keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Вход", CallbackData: nav.PageLogin.Link()}})
	}
	t.SendPage(
		chatId,
		"Главная",
		description,
		"Меню:",
		keyboard,
	)
	u.SetState(nav.Menu)
}

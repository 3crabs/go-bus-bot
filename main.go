package main

import (
	"context"
	"fmt"
	"github.com/3crabs/go-bus-api/bus"
	"github.com/3crabs/go-bus-bot/nav"
	"github.com/3crabs/go-bus-bot/normalize"
	"github.com/3crabs/go-bus-bot/tg"
	user "github.com/3crabs/go-bus-bot/user"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

var backKeyboard = [][]tgbot.InlineKeyboardButton{
	{tgbot.NewInlineKeyboardButtonData("Назад", "back")},
}

var users map[int64]*user.User

func getUser(chatID int64) *user.User {
	_, ok := users[chatID]
	if !ok {
		users[chatID] = user.NewUser()
	}
	u, _ := users[chatID]
	return u
}

func main() {
	users = make(map[int64]*user.User)
	b := bus.NewBus("http", "185.119.59.74:8090")

	bot, err := tgbot.NewBotAPI("5087528840:AAFSQGdR2zxUI6PzEiac9UoWJees1s74Ap4")
	if err != nil {
		log.Fatalln(err)
	}

	t := tg.NewTg(bot)

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Bot is start up!")

	for update := range updates {
		if update.CallbackQuery == nil && update.Message == nil {
			continue
		}

		var chatId int64
		var text string
		var buttonID string

		if update.CallbackQuery != nil {
			chatId = int64(update.CallbackQuery.From.ID)
			text = update.CallbackQuery.Data

			if strings.HasPrefix(text, "page") || text == "back" {
				getUser(chatId).SetPage(text)
			}
			if strings.Contains(text, "_") {
				words := strings.Split(text, "_")
				buttonID = words[1]
			}
		}

		if update.Message != nil {
			chatId = update.Message.Chat.ID
			text = update.Message.Text
		}

		u := getUser(chatId)
		switch u.Page {

		case nav.PageMain:
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

		case nav.PagePassengers:
			passengers, err := b.GetPassengers(context.Background(), getUser(chatId).AccessToken)
			if err != nil {
				msg := tgbot.NewMessage(chatId, "Что то пошло не так(\n\nПопробуйте позже")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				continue
			}
			getUser(chatId).Passengers = *passengers
			var keyboard [][]tgbot.InlineKeyboardButton
			for _, p := range *passengers {
				name := fmt.Sprintf("%s %s", p.LastName, p.FirstName)
				if p.Owner {
					name += " (Вы)"
				}
				keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
					tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("pageOnePassenger_%d", p.Id)),
				})
			}
			if len(*passengers) == 0 {
				keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
					tgbot.NewInlineKeyboardButtonData("Ввести свои данные", "pageAddMainPassenger"),
				})
			}
			keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData("Назад", "back"),
			})
			msg := tgbot.NewMessage(chatId, fmt.Sprintf("Пассажиры %d", len(*passengers)))
			msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
			_, _ = bot.Send(msg)

		case nav.PageAddMainPassenger:
			switch u.State {
			case nav.Menu:
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите фамилию")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitLastName)
			case nav.WaitLastName:
				lastName, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageAddPassengerData.LastName = lastName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите имя")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitFirstName)
			case nav.WaitFirstName:
				firstName, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageAddPassengerData.FirstName = firstName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите отчество")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitMiddleName)
			case nav.WaitMiddleName:
				middleName, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("мужской", "man")},
					{tgbot.NewInlineKeyboardButtonData("женский", "woman")},
				}
				getUser(chatId).PageAddPassengerData.MiddleName = middleName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВыберите пол")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitGender)
			case nav.WaitGender:
				gender, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				if gender == "man" {
					getUser(chatId).PageAddPassengerData.Gender = "M"
				} else {
					getUser(chatId).PageAddPassengerData.Gender = "F"
				}
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите серию паспорта")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitDocSeries)
			case nav.WaitDocSeries:
				docSeries, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageAddPassengerData.DocSeries = docSeries
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите номер паспорта")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitDocNum)
			case nav.WaitDocNum:
				docNum, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageAddPassengerData.DocNum = docNum
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите почту")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitEmail)
			case nav.WaitEmail:
				email, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageAddPassengerData.Email = email
				getUser(chatId).PageAddPassengerData.Phone = getUser(chatId).PageLoginData.Phone
				getUser(chatId).PageAddPassengerData.Owner = true
				getUser(chatId).PageAddPassengerData.Citizenship = "РОССИЯ"
				getUser(chatId).PageAddPassengerData.DocTypeCode = "1" // паспорт
				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Все верно", "submit")},
				}
				p := getUser(chatId).PageAddPassengerData
				gender := "мужской"
				if getUser(chatId).PageAddPassengerData.Gender == "F" {
					gender = "женский"
				}
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nПроверьте свои данные\n\n"+
					fmt.Sprintf("%s %s %s\n", p.LastName, p.FirstName, p.MiddleName)+
					fmt.Sprintf("Пол: %s\n", gender)+
					fmt.Sprintf("Паспорт: %s %s\n", p.DocSeries, p.DocNum)+
					fmt.Sprintf("Email: %s\n", p.Email)+
					fmt.Sprintf("Телефон: %s\n", p.Phone),
				)
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitSubmit)
			case nav.WaitSubmit:
				if text == "submit" {
					_, err := b.AddPassenger(context.Background(), getUser(chatId).AccessToken, getUser(chatId).PageAddPassengerData)
					if err != nil {
						msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось добавить пассажира, попробуйте позже")
						_, _ = bot.Send(msg)
						continue
					}
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nПассажир успешно добавлен")
					msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.Menu)
				}
			}

		case nav.PageOnePassenger:
			var p bus.PassengerDTO
			for _, passenger := range getUser(chatId).Passengers {
				if passenger.Id == getUser(chatId).PageOnePassengerData.Id {
					p = passenger
					break
				}
			}
			gender := "мужской"
			if getUser(chatId).PageAddPassengerData.Gender == "F" {
				gender = "женский"
			}
			msg := tgbot.NewMessage(chatId, "Информация о пассажире\n\n"+
				fmt.Sprintf("%s %s %s\n", p.LastName, p.FirstName, p.MiddleName)+
				fmt.Sprintf("Пол: %s\n", gender)+
				fmt.Sprintf("Паспорт: %s %s\n", p.DocSeries, p.DocNum)+
				fmt.Sprintf("Email: %s\n", p.Email)+
				fmt.Sprintf("Телефон: %s\n", p.Phone),
			)
			msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
			_, _ = bot.Send(msg)

		case nav.PageFindRaces:
			switch u.State {
			case nav.Menu:
				msg := tgbot.NewMessage(chatId, "Рейсы\n\nВведите название точки отправления или ее часть")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitFromPattern)
			case nav.WaitFromPattern:
				fromPattern, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				fromPoints, err := b.GetPointsFrom(context.Background(), fromPattern)
				if err != nil {
					log.Println(err)
				}
				if len(*fromPoints) == 0 {
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nТочек отправления не найдено\n\nВведите название точки отправления или ее часть")
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitFromPattern)
				}
				if len(*fromPoints) == 1 {
					id := (*fromPoints)[0].Id
					getUser(chatId).PageFindRacesData.From = id
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nВведите название точки прибытия или ее часть")
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitToPattern)
				}
				if len(*fromPoints) > 1 {
					var keyboard [][]tgbot.InlineKeyboardButton
					for _, p := range *fromPoints {
						name := fmt.Sprintf("%s %s", p.Name, p.Address)
						keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
							tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitFrom_%d", p.Id)),
						})
					}
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nВыберите точку отправления")
					msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitFrom)
				}
			case nav.WaitFrom:
				id, _ := strconv.Atoi(buttonID)
				getUser(chatId).PageFindRacesData.From = id
				msg := tgbot.NewMessage(chatId, "Рейсы\n\nВведите название точки прибытия или ее часть")
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitToPattern)
			case nav.WaitToPattern:
				toPattern, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				toPoints, err := b.GetPointsTo(context.Background(), getUser(chatId).PageFindRacesData.From, toPattern)
				if err != nil {
					log.Println(err)
				}
				if len(*toPoints) == 0 {
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nТочек прибытия не найдено\n\nВведите название точки отправления или ее часть")
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitToPattern)
				}
				if len(*toPoints) == 1 {
					id := (*toPoints)[0].Id
					getUser(chatId).PageFindRacesData.To = id
					date := getUser(chatId).PageFindRacesData
					races, err := b.GetRaces(context.Background(), date.From, date.To, "10.01.2022", 1)
					if err != nil {
						log.Println(err)
					}
					var keyboard [][]tgbot.InlineKeyboardButton
					for _, r := range *races {
						name := fmt.Sprintf("%v %.2fруб", r.ArrivalDate.Format("15:04"), r.Price)
						keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
							tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitRace_%s", r.Uid)),
						})
					}
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nВыберите рейс")
					msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
					getUser(chatId).SetState(nav.WaitRace)
					_, _ = bot.Send(msg)
				}
				if len(*toPoints) > 1 {
					var keyboard [][]tgbot.InlineKeyboardButton
					for _, p := range *toPoints {
						name := fmt.Sprintf("%s %s", p.Name, p.Address)
						keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
							tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitTo_%d", p.Id)),
						})
					}
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nВыберите точку прибытия")
					msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitTo)
				}
			case nav.WaitTo:
				id, _ := strconv.Atoi(buttonID)
				getUser(chatId).PageFindRacesData.To = id
				date := getUser(chatId).PageFindRacesData
				races, err := b.GetRaces(context.Background(), date.From, date.To, "10.01.2022", 1)
				if err != nil {
					log.Println(err)
				}
				var keyboard [][]tgbot.InlineKeyboardButton
				for _, r := range *races {
					name := fmt.Sprintf("%v %.2fруб", r.ArrivalDate.Format("15:04"), r.Price)
					keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
						tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitRace_%s", r.Uid)),
					})
				}
				msg := tgbot.NewMessage(chatId, "Рейсы\n\nВыберите рейс")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				getUser(chatId).SetState(nav.WaitRace)
				_, _ = bot.Send(msg)
			case nav.WaitRace:
			}

		case nav.PageLogin:
			switch u.State {
			case nav.Menu:
				msg := tgbot.NewMessage(chatId, "Вход\n\nВведите номер телефона")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitPhone)
			case nav.WaitPhone:
				phone, err := normalize.Phone(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалост разобрать номер, повторите попытку ввода")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).PageLoginData.Phone = phone

				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Да, я помню", "loginWithoutSMS")},
					{tgbot.NewInlineKeyboardButtonData("Нет, пришлите по СМС", "loginWithSMS")},
				}
				msg := tgbot.NewMessage(chatId, "Вход\n\nПомните пароль?")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.WaitSelectLogin)
			case nav.WaitSelectLogin:
				if text == "loginWithSMS" {
					if err := b.Register(context.Background(), getUser(chatId).PageLoginData.Phone); err != nil {
						log.Println(err)
					}
					msg := tgbot.NewMessage(chatId, fmt.Sprintf("Вход\n\nНа номер %s отправлено смс с паролем\n\nВведите пароль", getUser(chatId).PageLoginData.Phone))
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitPassword)
				}
				if text == "loginWithoutSMS" {
					msg := tgbot.NewMessage(chatId, "Вход\n\nВведите пароль")
					_, _ = bot.Send(msg)
					getUser(chatId).SetState(nav.WaitPassword)
				}
			case nav.WaitPassword:
				password, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось разобрать пароль, повторите попытку ввода")
					_, _ = bot.Send(msg)
					continue
				}
				login, err := b.Login(context.Background(), getUser(chatId).PageLoginData.Phone, password)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось авторизоваться, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).AccessToken = login.AccessToken
				getUser(chatId).Login = true

				msg := tgbot.NewMessage(chatId, "Вход\n\nВы успешно вошли")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).SetState(nav.Menu)
			}
		}
	}
}

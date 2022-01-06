package main

import (
	"bus_tg_bot/normalize"
	"context"
	"fmt"
	"github.com/3crabs/go-bus-api/bus"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

var backKeyboard = [][]tgbot.InlineKeyboardButton{
	{tgbot.NewInlineKeyboardButtonData("Назад", "back")},
}

var users map[int64]*user

func getUser(chatID int64) *user {
	_, ok := users[chatID]
	if !ok {
		users[chatID] = newUser()
	}
	u, _ := users[chatID]
	return u
}

func main() {
	users = make(map[int64]*user)
	b := bus.NewBus("http", "185.119.59.74:8090")

	bot, err := tgbot.NewBotAPI("5087528840:AAFSQGdR2zxUI6PzEiac9UoWJees1s74Ap4")
	if err != nil {
		log.Fatalln(err)
	}

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
				getUser(chatId).setPage(text)
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
		switch u.page {

		case pageMain:
			keyboard := [][]tgbot.InlineKeyboardButton{
				{tgbot.NewInlineKeyboardButtonData("Рейсы", "pageFindRaces")},
			}
			if u.login {
				keyboard = append(keyboard, []tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Пассажиры", "pagePassengers")})
			}
			if !u.login {
				keyboard = append(keyboard, []tgbot.InlineKeyboardButton{tgbot.NewInlineKeyboardButtonData("Вход", "pageLogin")})
			}
			msg := tgbot.NewMessage(chatId, "Главная")
			msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
			_, _ = bot.Send(msg)

		case pagePassengers:
			passengers, err := b.GetPassengers(context.Background(), getUser(chatId).accessToken)
			if err != nil {
				msg := tgbot.NewMessage(chatId, "Что то пошло не так(\n\nПопробуйте позже")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				continue
			}
			getUser(chatId).passengers = *passengers
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

		case pageAddMainPassenger:
			switch u.state {
			case menu:
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите фамилию")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitLastName)
			case waitLastName:
				lastName, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageAddPassengerData.LastName = lastName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите имя")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitFirstName)
			case waitFirstName:
				firstName, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageAddPassengerData.FirstName = firstName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите отчество")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitMiddleName)
			case waitMiddleName:
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
				getUser(chatId).pageAddPassengerData.MiddleName = middleName
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВыберите пол")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitGender)
			case waitGender:
				gender, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				if gender == "man" {
					getUser(chatId).pageAddPassengerData.Gender = "M"
				} else {
					getUser(chatId).pageAddPassengerData.Gender = "F"
				}
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите серию паспорта")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitDocSeries)
			case waitDocSeries:
				docSeries, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageAddPassengerData.DocSeries = docSeries
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите номер паспорта")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitDocNum)
			case waitDocNum:
				docNum, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageAddPassengerData.DocNum = docNum
				msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nВведите почту")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitEmail)
			case waitEmail:
				email, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageAddPassengerData.Email = email
				getUser(chatId).pageAddPassengerData.Phone = getUser(chatId).pageLoginData.Phone
				getUser(chatId).pageAddPassengerData.Owner = true
				getUser(chatId).pageAddPassengerData.Citizenship = "РОССИЯ"
				getUser(chatId).pageAddPassengerData.DocTypeCode = "1" // паспорт
				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Все верно", "submit")},
				}
				p := getUser(chatId).pageAddPassengerData
				gender := "мужской"
				if getUser(chatId).pageAddPassengerData.Gender == "F" {
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
				getUser(chatId).setState(waitSubmit)
			case waitSubmit:
				if text == "submit" {
					_, err := b.AddPassenger(context.Background(), getUser(chatId).accessToken, getUser(chatId).pageAddPassengerData)
					if err != nil {
						msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось добавить пассажира, попробуйте позже")
						_, _ = bot.Send(msg)
						continue
					}
					msg := tgbot.NewMessage(chatId, "Создание пассажира\n\nПассажир успешно добавлен")
					msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
					_, _ = bot.Send(msg)
					getUser(chatId).setState(menu)
				}
			}

		case pageOnePassenger:
			var p bus.PassengerDTO
			for _, passenger := range getUser(chatId).passengers {
				if passenger.Id == getUser(chatId).pageOnePassengerData.id {
					p = passenger
					break
				}
			}
			gender := "мужской"
			if getUser(chatId).pageAddPassengerData.Gender == "F" {
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

		case pageFindRaces:
			switch u.state {
			case menu:
				msg := tgbot.NewMessage(chatId, "Рейсы\n\nВведите название точки отправления или ее часть")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitFromPattern)
			case waitFromPattern:
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
				getUser(chatId).setState(waitFrom)
			case waitFrom:
				id, _ := strconv.Atoi(buttonID)
				getUser(chatId).pageFindRacesData.from = id
				msg := tgbot.NewMessage(chatId, "Рейсы\n\nВведите название точки прибытия или ее часть")
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitToPattern)
			case waitToPattern:
				toPattern, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Рейсы\n\nНе смог разобрать текст, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				toPoints, err := b.GetPointsTo(context.Background(), getUser(chatId).pageFindRacesData.from, toPattern)
				if err != nil {
					log.Println(err)
				}
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
				getUser(chatId).setState(waitTo)
			case waitTo:
				id, _ := strconv.Atoi(buttonID)
				getUser(chatId).pageFindRacesData.to = id
				date := getUser(chatId).pageFindRacesData
				races, err := b.GetRaces(context.Background(), date.from, date.to, "08.01.2022", 1)
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
				getUser(chatId).setState(waitRace)
				_, _ = bot.Send(msg)
			case waitRace:

			}

		case pageLogin:
			switch u.state {
			case menu:
				msg := tgbot.NewMessage(chatId, "Вход\n\nВведите номер телефона")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitPhone)
			case waitPhone:
				phone, err := normalize.Phone(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалост разобрать номер, повторите попытку ввода")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).pageLoginData.Phone = phone

				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Да, я помню", "loginWithoutSMS")},
					{tgbot.NewInlineKeyboardButtonData("Нет, пришлите по СМС", "loginWithSMS")},
				}
				msg := tgbot.NewMessage(chatId, "Вход\n\nПомните пароль?")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(keyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(waitSelectLogin)
			case waitSelectLogin:
				if text == "loginWithSMS" {
					if err := b.Register(context.Background(), getUser(chatId).pageLoginData.Phone); err != nil {
						log.Println(err)
					}
					msg := tgbot.NewMessage(chatId, fmt.Sprintf("Вход\n\nНа номер %s отправлено смс с паролем\n\nВведите пароль", getUser(chatId).pageLoginData.Phone))
					_, _ = bot.Send(msg)
					getUser(chatId).setState(waitPassword)
				}
				if text == "loginWithoutSMS" {
					msg := tgbot.NewMessage(chatId, "Вход\n\nВведите пароль")
					_, _ = bot.Send(msg)
					getUser(chatId).setState(waitPassword)
				}
			case waitPassword:
				password, err := normalize.String(text)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось разобрать пароль, повторите попытку ввода")
					_, _ = bot.Send(msg)
					continue
				}
				login, err := b.Login(context.Background(), getUser(chatId).pageLoginData.Phone, password)
				if err != nil {
					msg := tgbot.NewMessage(chatId, "Вход\n\nНе удалось авторизоваться, попробуйте позже")
					_, _ = bot.Send(msg)
					continue
				}
				getUser(chatId).accessToken = login.AccessToken
				getUser(chatId).login = true

				msg := tgbot.NewMessage(chatId, "Вход\n\nВы успешно вошли")
				msg.ReplyMarkup = tgbot.NewInlineKeyboardMarkup(backKeyboard...)
				_, _ = bot.Send(msg)
				getUser(chatId).setState(menu)
			}
		}
	}
}

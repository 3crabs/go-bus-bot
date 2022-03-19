package main

import (
	"context"
	"fmt"
	"github.com/3crabs/go-bus-api/bus"
	"github.com/3crabs/go-bus-bot/nav"
	"github.com/3crabs/go-bus-bot/normalize"
	"github.com/3crabs/go-bus-bot/server"
	"github.com/3crabs/go-bus-bot/tg"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

var backKeyboard = [][]tgbot.InlineKeyboardButton{
	{tgbot.NewInlineKeyboardButtonData("Назад", "back")},
}

func main() {
	s := server.NewServer()

	b := bus.NewBus("https", "passenger.busbonus.ru", "")
	log.Println("Create Bus SDK")

	bot, err := tgbot.NewBotAPI("5087528840:AAFSQGdR2zxUI6PzEiac9UoWJees1s74Ap4")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Create Tg SDK")

	t := tg.NewTg(bot)
	log.Println("Create Tg WRAPPER")

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("RUN")

	for update := range updates {
		if update.CallbackQuery == nil && update.Message == nil {
			continue
		}

		var chatId int64
		var text string
		var buttonID string
		u := s.GetUser(chatId)

		if update.CallbackQuery != nil {
			chatId = int64(update.CallbackQuery.From.ID)
			text = update.CallbackQuery.Data

			if strings.HasPrefix(text, "page") || text == "back" {
				u.SetPage(text)
			}
			if strings.Contains(text, "_") {
				words := strings.Split(text, "_")
				buttonID = words[1]
			}
		}

		if update.Message != nil {
			chatId = update.Message.Chat.ID
			text = update.Message.Text

			if text == "/start" {
				u.SetPage(string(nav.PageMain))
				u.SetState(nav.Menu)
			}
		}

		switch u.Page {

		case nav.PageMain:
			t.ShowPageMain(chatId, u)

		case nav.PagePassengers:
			passengers, err := b.GetPassengers(context.Background(), u.AccessToken)
			if err != nil {
				t.SendPageError(chatId, "Не удалось получить список пассажиров", "Попробуйте позже")
				continue
			}
			u.Passengers = *passengers
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
				keyboard = append(keyboard, []tgbot.InlineKeyboardButton{{Text: "Ввести свои данные", CallbackData: nav.PageAddMainPassenger.Link()}})
			}
			keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
				tgbot.NewInlineKeyboardButtonData("Назад", "back"),
			})
			t.SendPage(
				chatId,
				"Пассажиры",
				"Количество пассажиров: "+strconv.Itoa(len(*passengers)),
				"Выберите пассажира",
				keyboard,
			)

		case nav.PageAddMainPassenger:
			switch u.State {
			case nav.Menu:
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите фамилию",
					backKeyboard,
				)
				u.SetState(nav.WaitLastName)
			case nav.WaitLastName:
				lastName, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				u.PageAddPassengerData.LastName = lastName
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите имя",
					nil,
				)
				u.SetState(nav.WaitFirstName)
			case nav.WaitFirstName:
				firstName, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				u.PageAddPassengerData.FirstName = firstName
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите отчество",
					nil,
				)
				u.SetState(nav.WaitMiddleName)
			case nav.WaitMiddleName:
				middleName, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("мужской", "man")},
					{tgbot.NewInlineKeyboardButtonData("женский", "woman")},
				}
				u.PageAddPassengerData.MiddleName = middleName
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Выберите пол",
					keyboard,
				)
				u.SetState(nav.WaitGender)
			case nav.WaitGender:
				gender, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				if gender == "man" {
					u.PageAddPassengerData.Gender = "M"
				} else {
					u.PageAddPassengerData.Gender = "F"
				}
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите серию паспорта",
					nil,
				)
				u.SetState(nav.WaitDocSeries)
			case nav.WaitDocSeries:
				docSeries, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				u.PageAddPassengerData.DocSeries = docSeries
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите номер паспорта",
					nil,
				)
				u.SetState(nav.WaitDocNum)
			case nav.WaitDocNum:
				docNum, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				u.PageAddPassengerData.DocNum = docNum
				t.SendPage(
					chatId,
					"Создание пассажира",
					"",
					"Введите почту",
					nil,
				)
				u.SetState(nav.WaitEmail)
			case nav.WaitEmail:
				email, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				u.PageAddPassengerData.Email = email
				u.PageAddPassengerData.Phone = u.PageLoginData.Phone
				u.PageAddPassengerData.Owner = true
				u.PageAddPassengerData.Citizenship = "РОССИЯ"
				u.PageAddPassengerData.DocTypeCode = "1" // паспорт
				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Все верно", "submit")},
				}
				p := u.PageAddPassengerData
				gender := "мужской"
				if u.PageAddPassengerData.Gender == "F" {
					gender = "женский"
				}
				t.SendPage(
					chatId,
					"Создание пассажира",
					"Проверьте свои данные\n\n"+
						fmt.Sprintf("%s %s %s\n", p.LastName, p.FirstName, p.MiddleName)+
						fmt.Sprintf("Пол: %s\n", gender)+
						fmt.Sprintf("Паспорт: %s %s\n", p.DocSeries, p.DocNum)+
						fmt.Sprintf("Email: %s\n", p.Email)+
						fmt.Sprintf("Телефон: %s\n", p.Phone),
					"Выберите действие",
					keyboard,
				)
				u.SetState(nav.WaitSubmit)
			case nav.WaitSubmit:
				if text == "submit" {
					_, err := b.AddPassenger(context.Background(), u.AccessToken, u.PageAddPassengerData)
					if err != nil {
						t.SendPageError(chatId, "Не удалось добавить пассажира", "Проверьте данные и попробуйте позже")
						continue
					}
					t.SendPage(
						chatId,
						"Создание пассажира",
						"",
						"Пассажир успешно добавлен",
						backKeyboard,
					)
					u.SetState(nav.Menu)
				}
			}

		case nav.PageOnePassenger:
			var p bus.PassengerDTO
			for _, passenger := range u.Passengers {
				if passenger.Id == u.PageOnePassengerData.Id {
					p = passenger
					break
				}
			}
			gender := "мужской"
			if u.PageAddPassengerData.Gender == "F" {
				gender = "женский"
			}
			t.SendPage(
				chatId,
				"Информация о пассажире",
				"Информация о пассажире\n\n"+
					fmt.Sprintf("%s %s %s\n", p.LastName, p.FirstName, p.MiddleName)+
					fmt.Sprintf("Пол: %s\n", gender)+
					fmt.Sprintf("Паспорт: %s %s\n", p.DocSeries, p.DocNum)+
					fmt.Sprintf("Email: %s\n", p.Email)+
					fmt.Sprintf("Телефон: %s\n", p.Phone),
				"Выберите действие",
				backKeyboard,
			)

		case nav.PageFindRaces:
			switch u.State {
			case nav.Menu:
				t.SendPage(
					chatId,
					"Рейсы",
					"",
					"Введите название точки отправления или ее часть",
					backKeyboard,
				)
				u.SetState(nav.WaitFromPattern)
			case nav.WaitFromPattern:
				fromPattern, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				fromPoints, err := b.GetPointsFrom(context.Background(), fromPattern)
				if err != nil {
					log.Println(err)
					continue
				}
				if len(*fromPoints) == 0 {
					t.SendPage(
						chatId,
						"Рейсы",
						"Точек отправления не найдено",
						"Введите название точки отправления или ее часть",
						nil,
					)
					u.SetState(nav.WaitFromPattern)
				}
				if len(*fromPoints) == 1 {
					id := (*fromPoints)[0].Id
					u.PageFindRacesData.From = id
					t.SendPage(
						chatId,
						"Рейсы",
						"",
						"Введите название точки прибытия или ее часть",
						nil,
					)
					u.SetState(nav.WaitToPattern)
				}
				if len(*fromPoints) > 1 {
					var keyboard [][]tgbot.InlineKeyboardButton
					for _, p := range *fromPoints {
						name := fmt.Sprintf("%s %s", p.Name, p.Address)
						keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
							tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitFrom_%d", p.Id)),
						})
					}
					t.SendPage(
						chatId,
						"Рейсы",
						"",
						"Выберите точку отправления",
						keyboard,
					)
					u.SetState(nav.WaitFrom)
				}
			case nav.WaitFrom:
				id, _ := strconv.Atoi(buttonID)
				u.PageFindRacesData.From = id
				t.SendPage(
					chatId,
					"Рейсы",
					"",
					"Введите название точки прибытия или ее часть",
					nil,
				)
				u.SetState(nav.WaitToPattern)
			case nav.WaitToPattern:
				toPattern, err := normalize.String(text)
				if err != nil {
					t.SendPageError(chatId, "Не смог разобрать текст", "Проверьте данные и попробуйте снова")
					continue
				}
				toPoints, err := b.GetPointsTo(context.Background(), u.PageFindRacesData.From, toPattern)
				if err != nil {
					log.Println(err)
				}
				if len(*toPoints) == 0 {
					t.SendPage(
						chatId,
						"Рейсы",
						"Точек прибытия не найдено",
						"Введите название точки отправления или ее часть",
						nil,
					)
					u.SetState(nav.WaitToPattern)
				}
				if len(*toPoints) == 1 {
					id := (*toPoints)[0].Id
					u.PageFindRacesData.To = id
					date := u.PageFindRacesData
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
					t.SendPage(
						chatId,
						"Рейсы",
						"",
						"Выберите рейс",
						keyboard,
					)
					u.SetState(nav.WaitRace)
				}
				if len(*toPoints) > 1 {
					var keyboard [][]tgbot.InlineKeyboardButton
					for _, p := range *toPoints {
						name := fmt.Sprintf("%s %s", p.Name, p.Address)
						keyboard = append(keyboard, []tgbot.InlineKeyboardButton{
							tgbot.NewInlineKeyboardButtonData(name, fmt.Sprintf("waitTo_%d", p.Id)),
						})
					}
					t.SendPage(
						chatId,
						"Рейсы",
						"",
						"Выберите точку прибытия",
						keyboard,
					)
					u.SetState(nav.WaitTo)
				}
			case nav.WaitTo:
				id, _ := strconv.Atoi(buttonID)
				u.PageFindRacesData.To = id
				date := u.PageFindRacesData
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
				t.SendPage(
					chatId,
					"Рейсы",
					"",
					"Выберите рейс",
					keyboard,
				)
				u.SetState(nav.WaitRace)
			case nav.WaitRace:
			}

		case nav.PageLogin:
			switch u.State {
			case nav.Menu:
				t.SendPage(
					chatId,
					"Вход",
					"",
					"Введите номер телефона",
					backKeyboard,
				)
				u.SetState(nav.WaitPhone)
			case nav.WaitPhone:
				phone, err := normalize.Phone(text)
				if err != nil {
					t.SendPageError(
						chatId,
						"Не удалост разобрать номер",
						"Повторите попытку ввода",
					)
					continue
				}
				u.PageLoginData.Phone = phone

				keyboard := [][]tgbot.InlineKeyboardButton{
					{tgbot.NewInlineKeyboardButtonData("Да, я помню", "loginWithoutSMS")},
					{tgbot.NewInlineKeyboardButtonData("Нет, пришлите по СМС", "loginWithSMS")},
				}
				t.SendPage(
					chatId,
					"Вход",
					"",
					"Помните пароль?",
					keyboard,
				)
				u.SetState(nav.WaitSelectLogin)
			case nav.WaitSelectLogin:
				if text == "loginWithSMS" {
					if err := b.Register(context.Background(), u.PageLoginData.Phone); err != nil {
						log.Println(err)
					}
					t.SendPage(
						chatId,
						"Вход",
						fmt.Sprintf("На номер %s отправлено смс с паролем", u.PageLoginData.Phone),
						"Введите пароль",
						nil,
					)
					u.SetState(nav.WaitPassword)
				}
				if text == "loginWithoutSMS" {
					t.SendPage(
						chatId,
						"Вход",
						"",
						"Введите пароль",
						nil,
					)
					u.SetState(nav.WaitPassword)
				}
			case nav.WaitPassword:
				password, err := normalize.String(text)
				if err != nil {
					t.SendPageError(
						chatId,
						"Не удалось разобрать пароль",
						"Повторите попытку ввода",
					)
					continue
				}
				login, err := b.Login(context.Background(), u.PageLoginData.Phone, password)
				if err != nil {
					t.SendPageError(
						chatId,
						"Не удалось авторизоваться",
						"Попробуйте позже",
					)
					continue
				}
				u.AccessToken = login.AccessToken
				u.Login = true

				t.ShowPageMain(chatId, u)
			}

		case nav.PageFeedback:
			t.SendPage(
				chatId,
				"Отзывы",
				"",
				"Выберите действие",
				nil,
			)
		}
	}
}

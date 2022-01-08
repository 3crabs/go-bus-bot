package main

import (
	"github.com/3crabs/go-bus-api/bus"
	"github.com/3crabs/go-bus-bot/nav"
	"strconv"
	"strings"
)

type user struct {
	page                 nav.Page
	state                nav.State
	pageLoginData        bus.PhoneDTO
	passengers           []bus.PassengerDTO
	pageAddPassengerData bus.PassengerCreateDTO
	pageOnePassengerData struct {
		id int
	}
	pageFindRacesData struct {
		from int
		to   int
	}
	login       bool
	accessToken string
}

func newUser() *user {
	return &user{
		page:  nav.PageMain,
		state: nav.Menu,
		login: false,
	}
}

func (u *user) setPage(data string) {
	if data == "back" {
		switch u.page {
		case nav.PageFindRaces:
			u.page = nav.PageMain
		case nav.PagePassengers:
			u.page = nav.PageMain
		case nav.PageAddMainPassenger:
			u.page = nav.PagePassengers
		case nav.PageOnePassenger:
			u.page = nav.PagePassengers
		case nav.PageLogin:
			u.page = nav.PageMain
		}
	} else {
		if strings.Contains(data, "_") {
			words := strings.Split(data, "_")
			if words[0] == "pageOnePassenger" {
				id, _ := strconv.Atoi(words[1])
				u.pageOnePassengerData.id = id
				data = words[0]
			}
		}
		u.page = nav.Page(data)
	}
	if u.page == nav.PageMain {
		u.state = nav.Menu
	}
}

func (u *user) setState(s nav.State) {
	u.state = s
}

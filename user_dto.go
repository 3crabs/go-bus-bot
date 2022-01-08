package main

import (
	"github.com/3crabs/go-bus-api/bus"
	"github.com/3crabs/go-bus-bot/nav"
	"strconv"
	"strings"
)

type state string

const (
	menu = state("menu")

	// pageLogin
	waitPhone       = state("waitPhone")
	waitSelectLogin = state("waitSelectLogin")
	waitPassword    = state("waitPassword")

	// pageAddPassenger
	waitGender     = state("waitGender")
	waitLastName   = state("waitLastName")
	waitFirstName  = state("waitFirstName")
	waitMiddleName = state("waitMiddleName")
	waitDocSeries  = state("waitDocSeries")
	waitDocNum     = state("waitDocNum")
	waitEmail      = state("waitEmail")
	_              = state("waitPhone") // waitPhone
	waitSubmit     = state("waitSubmit")

	// pageFindRaces
	waitFromPattern = state("waitFromPattern")
	waitFrom        = state("waitFrom")
	waitToPattern   = state("waitToPattern")
	waitTo          = state("waitTo")
	waitDate        = state("waitDate")
	waitRace        = state("waitRace")
)

type user struct {
	page                 nav.Page
	state                state
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
		state: menu,
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
		u.state = menu
	}
}

func (u *user) setState(s state) {
	u.state = s
}

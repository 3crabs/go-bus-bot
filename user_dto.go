package main

import (
	"github.com/3crabs/go-bus-api/bus"
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

type page string

const (
	pageLogin            = page("pageLogin")
	pageMain             = page("pageMain")
	pageFindRaces        = page("pageFindRaces")
	pagePassengers       = page("pagePassengers")
	pageAddMainPassenger = page("pageAddMainPassenger")
	pageOnePassenger     = page("pageOnePassenger")
)

type user struct {
	page                 page
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
		page:  pageMain,
		state: menu,
		login: false,
	}
}

func (u *user) setPage(data string) {
	if data == "back" {
		switch u.page {
		case pageFindRaces:
			u.page = pageMain
		case pagePassengers:
			u.page = pageMain
		case pageAddMainPassenger:
			u.page = pagePassengers
		case pageOnePassenger:
			u.page = pagePassengers
		case pageLogin:
			u.page = pageMain
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
		u.page = page(data)
	}
	if u.page == pageMain {
		u.state = menu
	}
}

func (u *user) setState(s state) {
	u.state = s
}

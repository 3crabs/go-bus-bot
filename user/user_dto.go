package user

import (
	"github.com/3crabs/go-bus-api/bus"
	"github.com/3crabs/go-bus-bot/nav"
	"strconv"
	"strings"
)

type User struct {
	Page                 nav.Page
	State                nav.State
	PageLoginData        bus.PhoneDTO
	Passengers           []bus.PassengerDTO
	PageAddPassengerData bus.PassengerCreateRequest
	PageOnePassengerData struct {
		Id int
	}
	PageFindRacesData struct {
		From int
		To   int
	}
	Login       bool
	AccessToken string
}

func NewUser() *User {
	return &User{
		Page:  nav.PageMain,
		State: nav.Menu,
		Login: false,
	}
}

func (u *User) SetPage(data string) {
	if data == "back" {
		switch u.Page {
		case nav.PageFindRaces:
			u.Page = nav.PageMain
		case nav.PagePassengers:
			u.Page = nav.PageMain
		case nav.PageAddMainPassenger:
			u.Page = nav.PagePassengers
		case nav.PageOnePassenger:
			u.Page = nav.PagePassengers
		case nav.PageLogin:
			u.Page = nav.PageMain
		}
	} else {
		if strings.Contains(data, "_") {
			words := strings.Split(data, "_")
			if words[0] == "pageOnePassenger" {
				id, _ := strconv.Atoi(words[1])
				u.PageOnePassengerData.Id = id
				data = words[0]
			}
		}
		u.Page = nav.Page(data)
	}
	if u.Page == nav.PageMain {
		u.State = nav.Menu
	}
}

func (u *User) SetState(s nav.State) {
	u.State = s
}

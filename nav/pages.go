package nav

const (
	Menu = State("menu")

	PageLogin       = Page("pageLogin")
	WaitPhone       = State("waitPhone")
	WaitSelectLogin = State("waitSelectLogin")
	WaitPassword    = State("waitPassword")

	PageMain = Page("pageMain")

	PageFindRaces   = Page("pageFindRaces")
	WaitFromPattern = State("waitFromPattern")
	WaitFrom        = State("waitFrom")
	WaitToPattern   = State("waitToPattern")
	WaitTo          = State("waitTo")
	WaitDate        = State("waitDate")
	WaitRace        = State("waitRace")

	PagePassengers = Page("pagePassengers")

	PageAddMainPassenger = Page("pageAddMainPassenger")
	WaitGender           = State("waitGender")
	WaitLastName         = State("waitLastName")
	WaitFirstName        = State("waitFirstName")
	WaitMiddleName       = State("waitMiddleName")
	WaitDocSeries        = State("waitDocSeries")
	WaitDocNum           = State("waitDocNum")
	WaitEmail            = State("waitEmail")
	_                    = State("waitPhone") // waitPhone
	WaitSubmit           = State("waitSubmit")

	PageOnePassenger = Page("pageOnePassenger")
)

type Page string

func (p Page) Link() *string {
	s := string(p)
	return &s
}

type State string

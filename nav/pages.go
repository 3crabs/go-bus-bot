package nav

const (
	PageLogin            = Page("pageLogin")
	PageMain             = Page("pageMain")
	PageFindRaces        = Page("pageFindRaces")
	PagePassengers       = Page("pagePassengers")
	PageAddMainPassenger = Page("pageAddMainPassenger")
	PageOnePassenger     = Page("pageOnePassenger")
)

type Page string

func (p Page) Link() *string {
	s := string(p)
	return &s
}

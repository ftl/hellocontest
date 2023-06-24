package newcontest

type View interface {
	Show() bool

	SetContestIdentifiers(ids []string, texts []string)
	SelectContestIdentifier(value string)
	SetContestName(value string)
	SetDataComplete(bool)
}

type ContestProvider interface {
	ContestIdentifiers() ([]string, []string)
	ProposeContestName(string) string
}

type Result struct {
	Identifier string
	Name       string
	Filename   string
}

func NewController(contestProvider ContestProvider, logDirectory string) *Controller {
	result := &Controller{
		logDirectory:    logDirectory,
		contestProvider: contestProvider,
	}

	return result
}

type Controller struct {
	view         View
	logDirectory string

	contestProvider ContestProvider

	selectedIdentifier string
	selectedName       string
	selectedFilename   string
}

func (c *Controller) SetView(view View) {
	if view == nil {
		c.view = new(nullview)
		return
	}
	c.view = view
}

func (c *Controller) Run() (Result, bool) {
	c.view.SetContestIdentifiers(c.contestProvider.ContestIdentifiers())

	accepted := c.view.Show()

	if !accepted {
		return Result{}, false
	}
	return Result{
		Identifier: c.selectedIdentifier,
		Name:       c.selectedName,
		Filename:   c.selectedFilename,
	}, true
}

func (c *Controller) SelectContestIdentifier(identifier string) {
	oldDefaultName := c.contestProvider.ProposeContestName(c.selectedIdentifier)
	proposeNewName := c.selectedName == oldDefaultName || c.selectedName == ""

	c.selectedIdentifier = identifier

	if proposeNewName {
		c.selectedName = c.contestProvider.ProposeContestName(c.selectedIdentifier)
		c.view.SetContestName(c.selectedName)
	}

	c.view.SetDataComplete(c.selectedIdentifier != "")
}

func (c *Controller) EnterContestName(name string) {
	c.selectedName = name
}

type nullview struct{}

func (n *nullview) Show() bool                               { return false }
func (n *nullview) SetContestIdentifiers([]string, []string) {}
func (n *nullview) SelectContestIdentifier(string)           {}
func (n *nullview) SetContestName(string)                    {}
func (n *nullview) SetDataComplete(bool)                     {}

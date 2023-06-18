package newcontest

import (
	"os"
	"path/filepath"
)

type View interface {
	Show() bool

	SetContestIdentifiers(ids []string, texts []string)
	SelectContestIdentifier(value string)
	SetContestName(value string)
	SetContestFilename(value string)
	SetDataComplete(bool)
}

type FileSelector interface {
	SelectSaveFile(title string, dir string, filename string, patterns ...string) (string, bool, error)
	ShowErrorDialog(string, ...interface{})
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
	fileSelector FileSelector
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

func (c *Controller) SetFileSelector(fileSelector FileSelector) {
	if fileSelector == nil {
		c.fileSelector = new(nullFileSelector)
		return
	}
	c.fileSelector = fileSelector
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
}

func (c *Controller) EnterContestName(name string) {
	c.selectedName = name
}

func (c *Controller) ChooseContestFilename() {
	if c.selectedIdentifier != "" && c.selectedName == "" {
		c.selectedName = c.contestProvider.ProposeContestName(c.selectedIdentifier)
		c.view.SetContestName(c.selectedName)
	}

	proposedFilename := c.selectedName + ".log"

	filename, ok, err := c.fileSelector.SelectSaveFile("New Logfile", c.logDirectory, proposedFilename, "*.log")
	if !ok {
		return
	}
	if err != nil {
		c.fileSelector.ShowErrorDialog("Cannot select a file: %v", err)
		return
	}

	c.selectedFilename = filename

	_, err = os.Stat(filepath.Dir(c.selectedFilename))
	complete := err == nil

	c.view.SetContestFilename(c.selectedFilename)
	c.view.SetDataComplete(complete)
}

type nullview struct{}

func (n *nullview) Show() bool                               { return false }
func (n *nullview) SetContestIdentifiers([]string, []string) {}
func (n *nullview) SelectContestIdentifier(string)           {}
func (n *nullview) SetContestName(string)                    {}
func (n *nullview) SetContestFilename(string)                {}
func (n *nullview) SetDataComplete(bool)                     {}

type nullFileSelector struct{}

func (n *nullFileSelector) SelectSaveFile(title string, dir string, filename string, patterns ...string) (string, bool, error) {
	return "", false, nil
}
func (n *nullFileSelector) ShowErrorDialog(string, ...interface{}) {}

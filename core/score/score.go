package score

import (
	"sync"
	"time"

	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show()
	Hide()

	ShowScore(score core.Score)
	SetGoals(points int, multis int)
}

type Controller struct {
	score     core.Score
	scoreLock *sync.Mutex
	view      View

	contestStartTime  time.Time
	contestPointsGoal int
	contestMultisGoal int

	listeners []any
}

func NewController(settings core.Settings) *Controller {
	result := &Controller{
		score: core.NewScore(),
		view:  new(nullView),
	}

	result.setContest(settings.Contest())

	return result
}

func (c *Controller) SetView(view View) {
	if view == nil {
		panic("score.Counter.SetView must not be called with nil")
	}
	if _, ok := c.view.(*nullView); !ok {
		panic("score.Counter.SetView was already called")
	}

	c.view = view
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
	c.view.ShowScore(c.score)
}

func (c *Controller) ContestChanged(contest core.Contest) {
	c.setContest(contest)
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
}

func (c *Controller) setContest(contest core.Contest) {
	c.contestStartTime = contest.StartTime
	c.contestPointsGoal = contest.PointsGoal
	c.contestMultisGoal = contest.MultisGoal
}

func (c *Controller) ScoreChanged(score core.Score) {
	c.score = score
	c.view.ShowScore(c.score)
}

func (c *Controller) Show() {
	c.view.Show()
	c.view.SetGoals(c.contestPointsGoal, c.contestMultisGoal)
	c.view.ShowScore(c.score)
}

func (c *Controller) Hide() {
	c.view.Hide()
}

var _ View = new(nullView)

type nullView struct{}

func (v *nullView) Show()                {}
func (v *nullView) Hide()                {}
func (v *nullView) ShowScore(core.Score) {}
func (v *nullView) SetGoals(int, int)    {}

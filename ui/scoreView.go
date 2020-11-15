package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"github.com/ftl/hellocontest/core"
)

type scoreView struct {
	tableLabel *gtk.Label
}

func setupScoreView(builder *gtk.Builder) *scoreView {
	result := new(scoreView)

	result.tableLabel = getUI(builder, "tableLabel").(*gtk.Label)

	return result
}

func (v *scoreView) ShowScore(score core.Score) {
	if v == nil {
		return
	}

	renderedScore := fmt.Sprintf("<span allow_breaks='true' font_family='monospace'>%s</span>", score)
	v.tableLabel.SetMarkup(renderedScore)
}

package model

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/conval"
)

type Summary struct {
	// comes from the contest definition
	ContestName  string
	CabrilloName string

	// comes from the contest settings
	StartTime   time.Time
	Callsign    Callsign
	MyExchanges string
	QTCsEnabled bool

	// the dialog starts with the value of the contest settings and may change it
	DetailedMultis bool

	// has to be selected by the user in a dialog
	OperatorMode conval.OperatorMode
	Overlay      conval.Overlay
	PowerMode    conval.PowerMode
	Assisted     bool

	// comes from the score counter
	WorkedModes []string
	WorkedBands []string
	Score       Score
	TimeReport  TimeReport
}

func (s Summary) WorkingConditions() string {
	result := []string{}
	if s.OperatorMode != "" {
		result = append(result, fmt.Sprintf("%s operator", string(s.OperatorMode)))
	}
	if s.Overlay != "" {
		result = append(result, fmt.Sprintf("%s overlay", string(s.Overlay)))
	}
	if s.PowerMode != "" {
		result = append(result, fmt.Sprintf("%s power", string(s.PowerMode)))
	}
	if s.Assisted {
		result = append(result, "assisted")
	}
	return strings.Join(result, ", ")
}

func (s Summary) ScoreTable() string {
	columns := ScoreColumns(s.QTCsEnabled, s.DetailedMultis, s.Score.MultiProperties)

	rows := make([][]string, 0, len(Bands)+1)
	for _, band := range Bands {
		if score, ok := s.Score.ScorePerBand[band]; ok {
			rows = append(rows, scoreTableRow(columns, string(band), score))
		}
	}
	totalRow := scoreTableRow(columns, "Tot", s.Score.Result())

	widths := scoreTableWidths(columns, rows, totalRow)
	separator := strings.Repeat("-", scoreTableWidth(widths))

	buf := bytes.NewBufferString("")
	fmt.Fprintln(buf, scoreTableLine(columns, widths, scoreTableHeaders(columns)))
	fmt.Fprintln(buf, separator)
	for _, row := range rows {
		fmt.Fprintln(buf, scoreTableLine(columns, widths, row))
	}
	fmt.Fprintln(buf, separator)
	fmt.Fprintln(buf, scoreTableLine(columns, widths, totalRow))
	return buf.String()
}

func scoreTableHeaders(columns []ScoreColumn) []string {
	result := make([]string, len(columns)+1)
	result[0] = "Band"
	for i, column := range columns {
		result[i+1] = column.Header
	}
	return result
}

func scoreTableRow(columns []ScoreColumn, label string, score BandScore) []string {
	result := make([]string, len(columns)+1)
	result[0] = label
	for i, column := range columns {
		result[i+1] = column.Value(score)
	}
	return result
}

func scoreTableWidths(columns []ScoreColumn, rows [][]string, totalRow []string) []int {
	result := make([]int, len(columns)+1)
	for i, header := range scoreTableHeaders(columns) {
		result[i] = len(header)
	}
	for _, row := range append(rows, totalRow) {
		for i, cell := range row {
			result[i] = max(result[i], len(cell))
		}
	}
	return result
}

func scoreTableWidth(widths []int) int {
	result := 0
	for _, width := range widths {
		result += width + 1
	}
	return result - 1
}

func scoreTableLine(columns []ScoreColumn, widths []int, cells []string) string {
	result := make([]string, len(cells))
	for i, cell := range cells {
		// the band stands on the left, every value on the right
		if i == 0 {
			result[i] = fmt.Sprintf("%-*s", widths[i], cell)
		} else {
			result[i] = fmt.Sprintf("%*s", widths[i], cell)
		}
	}
	return strings.Join(result, " ")
}

type ScoreColumn struct {
	Header string
	Value  func(BandScore) string
}

func ScoreColumns(qtcsEnabled bool, detailedMultis bool, properties []conval.Property) []ScoreColumn {
	result := make([]ScoreColumn, 0, len(properties)+8)
	result = append(result, ScoreColumn{"QSOs", func(score BandScore) string { return strconv.Itoa(score.QSOs) }})
	if qtcsEnabled {
		result = append(result, ScoreColumn{"QTCs", func(score BandScore) string { return strconv.Itoa(score.QTCs) }})
	}
	result = append(result,
		ScoreColumn{"Dupes", func(score BandScore) string { return strconv.Itoa(score.Duplicates) }},
		ScoreColumn{"Points", func(score BandScore) string { return strconv.Itoa(score.Points) }},
		ScoreColumn{"P/Q", func(score BandScore) string { return fmt.Sprintf("%.1f", score.PointsPerQSO()) }},
	)

	// a contest without multiplier properties needs neither a multiplier column nor Q/M
	switch {
	case showMultisPerProperty(detailedMultis, properties):
		for _, property := range sortedMultiProperties(properties) {
			result = append(result, ScoreColumn{string(property), func(score BandScore) string {
				return strconv.Itoa(score.MultisPerProperty[property])
			}})
		}
	case len(properties) > 0:
		result = append(result, ScoreColumn{"Mult", func(score BandScore) string { return strconv.Itoa(score.Multis) }})
	}
	if len(properties) > 0 {
		result = append(result, ScoreColumn{"Q/M", func(score BandScore) string { return fmt.Sprintf("%.1f", score.QSOsPerMulti()) }})
	}

	result = append(result, ScoreColumn{"Result", func(score BandScore) string { return strconv.Itoa(score.Result()) }})
	return result
}

func showMultisPerProperty(detailedMultis bool, properties []conval.Property) bool {
	// one property alone needs no own column, the summed multis say the same
	return detailedMultis && len(properties) > 1
}

func sortedMultiProperties(properties []conval.Property) []conval.Property {
	result := slices.Clone(properties)
	slices.Sort(result)
	return result
}

type TimeReport = conval.TimeReport

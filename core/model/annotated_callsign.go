package model

import "strings"

// AnnotatedCallsign contains a callsign with additional information retrieved from databases and the logbook.
type AnnotatedCallsign struct {
	Callsign          Callsign
	Assembly          MatchingAssembly
	Duplicate         bool
	Worked            bool
	ExactMatch        bool
	Points            int
	Multis            int
	PredictedExchange []string
	Name              string
	UserText          string
	OnFrequency       bool

	Comparable any
	Compare    func(any, any) bool
}

func (c AnnotatedCallsign) LessThan(o AnnotatedCallsign) bool {
	if c.ExactMatch && !o.ExactMatch {
		return true
	}

	if c.Compare == nil {
		return false
	}
	if c.Comparable == nil || o.Comparable == nil {
		return false
	}

	return c.Compare(c.Comparable, o.Comparable)
}

type MatchingOperation int

const (
	Matching MatchingOperation = iota
	Insert
	Delete
	Substitute
	FalseFriend
)

type MatchingPart struct {
	OP    MatchingOperation
	Value string
}

type MatchingAssembly []MatchingPart

func (m MatchingAssembly) String() string {
	var result strings.Builder
	for _, match := range m {
		if match.OP != Delete {
			result.WriteString(match.Value)
		}
	}
	return result.String()
}

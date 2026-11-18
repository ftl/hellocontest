package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/conval"
)

// EntryField represents an entry field in the visual part.
type EntryField string

// The entry fields.
const (
	CallsignField EntryField = "callsign"
	BandField     EntryField = "band"
	ModeField     EntryField = "mode"
	OtherField    EntryField = "other"

	myExchangePrefix    string = "myExchange_"
	theirExchangePrefix string = "theirExchange_"
)

func (f EntryField) IsMyExchange() bool {
	return strings.HasPrefix(string(f), myExchangePrefix)
}

func (f EntryField) IsTheirExchange() bool {
	return strings.HasPrefix(string(f), theirExchangePrefix)
}

func IsExchangeField(name string) bool {
	return strings.HasPrefix(name, myExchangePrefix) || strings.HasPrefix(name, theirExchangePrefix)
}

func (f EntryField) ExchangeIndex() int {
	s := string(f)
	var a string
	switch {
	case strings.HasPrefix(s, myExchangePrefix):
		a = s[len(myExchangePrefix):]
	case strings.HasPrefix(s, theirExchangePrefix):
		a = s[len(theirExchangePrefix):]
	default:
		return -1
	}
	result, err := strconv.Atoi(a)
	if err != nil {
		return -1
	}
	return result
}

func (f EntryField) NextExchangeField() EntryField {
	s := string(f)
	var a string
	var prefix string
	switch {
	case strings.HasPrefix(s, myExchangePrefix):
		prefix = myExchangePrefix
		a = s[len(myExchangePrefix):]
	case strings.HasPrefix(s, theirExchangePrefix):
		prefix = theirExchangePrefix
		a = s[len(theirExchangePrefix):]
	default:
		return ""
	}
	i, err := strconv.Atoi(a)
	if err != nil {
		return ""
	}
	return EntryField(prefix + strconv.Itoa(i+1))
}

func MyExchangeField(index int) EntryField {
	return EntryField(fmt.Sprintf("%s%d", myExchangePrefix, index))
}

func TheirExchangeField(index int) EntryField {
	return EntryField(fmt.Sprintf("%s%d", theirExchangePrefix, index))
}

type ExchangeField struct {
	Field            EntryField
	CanContainSerial bool
	CanContainReport bool
	EmptyAllowed     bool
	Properties       conval.ExchangeField

	Short    string
	Name     string
	Hint     string
	ReadOnly bool
}

func DefinitionsToExchangeFields(fieldDefinitions []conval.ExchangeField, exchangeEntryField func(int) EntryField) []ExchangeField {
	result := make([]ExchangeField, 0, len(fieldDefinitions))
	for i, fieldDefinition := range fieldDefinitions {
		short := strings.Join(fieldDefinition.Strings(), "/")
		field := ExchangeField{
			Field:      exchangeEntryField(i + 1),
			Properties: fieldDefinition,
			Short:      short,
		}
		for _, property := range fieldDefinition {
			if property == conval.SerialNumberProperty {
				field.CanContainSerial = true
			}
			if property == conval.RSTProperty {
				field.CanContainReport = true
			}
			if property == conval.EmptyProperty {
				field.EmptyAllowed = true
			}
		}
		result = append(result, field)
	}
	return result
}

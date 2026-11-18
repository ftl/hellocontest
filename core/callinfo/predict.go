package callinfo

import (
	"strconv"

	"github.com/ftl/conval"
	"github.com/ftl/hellocontest/core/dxcc"
	"github.com/ftl/hellocontest/core/model"
)

func predictExchange(theirExchangeFields []model.ExchangeField, dxccEntity dxcc.Prefix, workedQSOs []model.QSO, currentExchange []string, historicExchange []string, historyAvailable []bool) []string {
	result := make([]string, len(theirExchangeFields))
	if len(currentExchange) > 0 {
		copy(result, currentExchange)
	}

	for i := range result {
		qsoExchange, foundInQSO := findExchangeInQSOs(i, workedQSOs)
		if foundInQSO {
			result[i] = qsoExchange
			continue
		}

		historicExchange, foundInHistory := findExchangeInHistory(theirExchangeFields, i, historicExchange, dxccEntity)
		if foundInHistory {
			result[i] = historicExchange
			continue
		}

		if result[i] == "" && i < len(historyAvailable) && historyAvailable[i] && isNoMemberField(theirExchangeFields[i]) {
			result[i] = string(conval.NoMemberProperty)
			// continue // for symmetry
		}
	}

	return result
}

func historyAvailability(history CallHistoryFinder, theirExchangeFields []model.ExchangeField) []bool {
	result := make([]bool, len(theirExchangeFields))
	for i := range result {
		result[i] = history.AvailableFor(i)
	}
	return result
}

func isNoMemberField(field model.ExchangeField) bool {
	return field.Properties.Contains(conval.MemberNumberProperty) &&
		field.Properties.Contains(conval.NoMemberProperty)
}

func findExchangeInQSOs(exchangeIndex int, workedQSOs []model.QSO) (string, bool) {
	result := ""
	found := false
	for _, qso := range workedQSOs {
		if exchangeIndex >= len(qso.TheirExchange) {
			break
		}
		exchange := qso.TheirExchange[exchangeIndex]
		if result == "" {
			result = exchange
			found = true
		} else if result != exchange {
			result = ""
			found = false
			break
		}
	}
	return result, found
}

func findExchangeInHistory(theirExchangeFields []model.ExchangeField, exchangeIndex int, historicExchange []string, dxccEntity dxcc.Prefix) (string, bool) {
	if exchangeIndex < len(historicExchange) && historicExchange[exchangeIndex] != "" {
		return historicExchange[exchangeIndex], true
	}

	if exchangeIndex >= len(theirExchangeFields) {
		return "", false
	}

	if dxccEntity.PrimaryPrefix != "" {
		field := theirExchangeFields[exchangeIndex]
		switch {
		case field.Properties.Contains(conval.CQZoneProperty):
			return strconv.Itoa(int(dxccEntity.CQZone)), true
		case field.Properties.Contains(conval.ITUZoneProperty):
			return strconv.Itoa(int(dxccEntity.ITUZone)), true
		case field.Properties.Contains(conval.DXCCEntityProperty),
			field.Properties.Contains(conval.DXCCPrefixProperty):
			return dxccEntity.PrimaryPrefix, true
		}
	}

	return "", false
}

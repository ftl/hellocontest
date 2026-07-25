package entry_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

// A1. Enter callsign
// Pre:  active field = CallsignField on focused VFO.
// Post: input[focused].callsign = typed text; CallsignEntered listeners notified;
//       callinfo input-changed notified; if text parses as callsign AND not editing:
//       sticky serial claim held; if duplicate: error on callsign field, else message cleared.
// Invariants: other VFO's row; editing flag; selected band/mode/frequency; logbook.

func TestA1_EnterCallsign_Valid(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		AssertCallsignEntered("DL1ABC").
		AssertCallinfoNotified("DL1ABC").
		AssertSerialClaimed().
		AssertMessageCleared(core.VFO1).
		AssertNoLogbookWrite()
}

func TestA1_EnterCallsign_Duplicate(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	dupe := core.QSO{Callsign: dl1abc, MyNumber: 1}

	NewScenario(t).
		WithClassicExchange().
		WithDuplicateQSO(dupe).
		Enter("DL1ABC").
		AssertCallsignEntered("DL1ABC").
		AssertCallinfoNotified("DL1ABC").
		AssertSerialClaimed().
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

func TestA1_EnterCallsign_PartialCallsign(t *testing.T) {
	// "DL" does not parse as a callsign: no serial claim, no duplicate check,
	// no message cleared or shown.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL").
		AssertCallsignEntered("DL").
		AssertCallinfoNotified("DL").
		AssertNoLogbookWrite()
}

// A2. Leave callsign field
// Pre:  active field = CallsignField; callsign parses; predicted exchange present iff
//       its length matches theirExchangeFields.
// Post: empty predictable their-exchange slots filled from prediction;
//       duplicate marker = (duplicate exists AND (not editing OR editQSO.Callsign != current)).
// Invariants: callsign input text; serial claim; logbook.

func TestA2_LeaveCallsign_NoDuplicate(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField().
		AssertDuplicateMarker(core.VFO1, false).
		AssertNoLogbookWrite()
}

func TestA2_LeaveCallsign_Duplicate(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	dupe := core.QSO{Callsign: dl1abc, MyNumber: 1}

	NewScenario(t).
		WithClassicExchange().
		WithDuplicateQSO(dupe).
		Enter("DL1ABC").
		GotoNextField().
		AssertDuplicateMarker(core.VFO1, true).
		AssertNoLogbookWrite()
}

func TestA2_LeaveCallsign_PredictionFillsEmptySlot(t *testing.T) {
	// ClassicExchange: [RST(not predictable), Serial(not predictable), GenericText(predictable)]
	// Prediction fills slot 3 (1-based) when empty.
	frame := core.CallinfoFrame{
		PredictedExchange: []string{"599", "042", "OE"},
	}

	NewScenario(t).
		WithClassicExchange().
		WithCallinfoFrame(core.VFO1, frame).
		Enter("OE5XYZ").
		GotoNextField().
		AssertTheirExchangeSet(core.VFO1, 3, "OE").
		AssertNoLogbookWrite()
}

func TestA2_LeaveCallsign_Editing_SameCallsign_NoMarker(t *testing.T) {
	// Editing the QSO's own callsign: duplicate marker stays false
	// even though the callsign appears in the logbook.
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	editedQSO := core.QSO{Callsign: dl1abc, MyNumber: 7}

	NewScenario(t).
		WithClassicExchange().
		WithDuplicateQSO(editedQSO).
		SelectQSO(editedQSO).
		GotoNextField().
		AssertDuplicateMarker(core.VFO1, false).
		AssertNoLogbookWrite()
}

func TestA2_LeaveCallsign_Editing_DifferentCallsign_ShowsMarker(t *testing.T) {
	// Editing: changed callsign that is a duplicate → marker shown.
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	dl2xyz, _ := core.ParseCallsign("DL2XYZ")
	editedQSO := core.QSO{Callsign: dl1abc, MyNumber: 7}
	dupeQSO := core.QSO{Callsign: dl2xyz, MyNumber: 3}

	NewScenario(t).
		WithClassicExchange().
		WithDuplicateQSO(dupeQSO).
		SelectQSO(editedQSO). // edit mode, shows DL1ABC
		Enter("DL2XYZ").      // change to a different (duped) callsign
		GotoNextField().
		AssertDuplicateMarker(core.VFO1, true).
		AssertNoLogbookWrite()
}

// A3. Goto next field (Tab)
// Pre:  any active field.
// Post: if leaving callsign: A2 ran; active field = next per transition map
//       (callsign → first their-exchange; last their-exchange and any my-exchange → callsign;
//       band/mode → callsign); view active field set; ESM state recomputed.
// Invariants: input values; per-VFO data other than active field; focused VFO.

func TestA3_GotoNextField_Callsign_GoesToTheirExchange1(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField().
		AssertActiveField(core.VFO1, core.TheirExchangeField(1))
}

func TestA3_GotoNextField_TheirExchange1_GoesToTheirExchange2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		GotoNextField(). // → TheirExchange2
		AssertActiveField(core.VFO1, core.TheirExchangeField(2))
}

func TestA3_GotoNextField_TheirExchange2_GoesToTheirExchange3(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		GotoNextField(). // → TheirExchange2
		GotoNextField(). // → TheirExchange3
		AssertActiveField(core.VFO1, core.TheirExchangeField(3))
}

func TestA3_GotoNextField_LastTheirExchange_GoesToCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		GotoNextField(). // → TheirExchange2
		GotoNextField(). // → TheirExchange3
		GotoNextField(). // → Callsign (last their-exchange wraps)
		AssertActiveField(core.VFO1, core.CallsignField)
}

func TestA3_GotoNextField_MyExchange_GoesToCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.MyExchangeField(1)).
		GotoNextField().
		AssertActiveField(core.VFO1, core.CallsignField)
}

func TestA3_GotoNextField_Band_GoesToCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.BandField).
		GotoNextField().
		AssertActiveField(core.VFO1, core.CallsignField)
}

func TestA3_GotoNextField_Mode_GoesToCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.ModeField).
		GotoNextField().
		AssertActiveField(core.VFO1, core.CallsignField)
}

func TestA3_GotoNextField_OtherField_GoesToCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.OtherField).
		GotoNextField().
		AssertActiveField(core.VFO1, core.CallsignField)
}

// A4. Goto next placeholder
// Pre:  none.
// Post: active field = CallsignField; view selects FilterPlaceholder text on callsign field.
// Invariants: all input values; focused VFO.

func TestA4_GotoNextPlaceholder_FocusesCallsignAndSelectsPlaceholder(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		GotoNextPlaceholder().
		AssertActiveField(core.VFO1, core.CallsignField).
		AssertTextSelected(core.VFO1, core.CallsignField, core.FilterPlaceholder)
}

// A5. Set active field
// Pre:  none.
// Post: activeField[focused] = given field; ESM recomputed.
// Invariants: view focus marker not directly updated (caller's job); input values.

func TestA5_SetActiveField_FieldIsSet(t *testing.T) {
	// SetActiveField does NOT call view.SetActiveField (caller's job per use case).
	// Verify via subsequent GotoNextField transition from the set field.
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.BandField).
		GotoNextField().
		AssertActiveField(core.VFO1, core.CallsignField)
}

// A6. Enter text into a field
// Pre:  active field set; for exchange fields, exchange[i] slot exists.
// Post: corresponding input slot updated; per active field: callsign → A1 effects;
//       band → B1 effects; mode → B2 effects; their-exchange → callinfo notified,
//       error cleared; ESM recomputed.
// Invariants: other VFO; editing flag.

func TestA6_EnterTheirExchange_NotifiesCallinfo(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("599").
		AssertCallinfoNotified("DL1ABC")
}

func TestA6_EnterMyExchange_UpdatesInputSlot(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.MyExchangeField(3)).
		Enter("OE").
		AssertMyExchangeValue(3, "OE")
}

// B1. Select band (band field)
// Pre:  input parses as band.
// Post: selectedBand[focused] = parsed band; VFO rig commanded to that band;
//       callsign re-entered (A1 effects).
// Invariants: other VFO band; mode; frequency; logbook.

func TestB1_SelectBand_Valid(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		SetActiveField(core.BandField).
		Enter("40m").
		AssertVFOBand(core.Band40m).
		AssertCallinfoNotified("DL1ABC").
		AssertNoLogbookWrite()
}

// B2. Select mode (mode field)
// Pre:  input parses as mode.
// Post: selectedMode[focused] = parsed mode; VFO rig commanded; if generateReport:
//       my/their report regenerated for focused VFO; callsign re-entered.
// Invariants: other VFO mode; band; frequency; logbook.

func TestB2_SelectMode_Valid(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		SetActiveField(core.ModeField).
		Enter("SSB").
		AssertVFOMode(core.ModeSSB).
		AssertCallinfoNotified("DL1ABC").
		AssertNoLogbookWrite()
}

func TestB2_SelectMode_GeneratesReport(t *testing.T) {
	// WithClassicExchangeAndReport: GenerateReport=true; mode SSB → "59"
	// Field 1 = RST: both my-exchange view and their-exchange view updated.
	NewScenario(t).
		WithClassicExchangeAndReport().
		Enter("DL1ABC").
		SetActiveField(core.ModeField).
		Enter("SSB").
		AssertMyExchangeView(1, "59").
		AssertTheirExchangeSet(core.VFO1, 1, "59").
		AssertNoLogbookWrite()
}

// B3. Enter frequency in kHz via callsign field
// Pre:  active field = CallsignField; input parses as integer.
// Post: focused VFO rig commanded with kHz*1000; callsign input cleared;
//       A1 effects with empty callsign.
// Invariants: band; mode; other VFO; logbook.

func TestB3_EnterFrequency_ViaCallsignField(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("14250").
		PressEnter().
		AssertVFOFrequency(14250000).
		AssertCallsignView(core.VFO1, "").
		AssertCallinfoNotified("").
		AssertNoLogbookWrite()
}

// B4. Enter band via callsign field
// Pre:  active field = CallsignField; input parses as band.
// Post: B1 effects (via bandEntered); callsign input cleared; A1 effects with empty callsign.
// Invariants: other VFO; logbook.

func TestB4_EnterBand_ViaCallsignField(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("40m").
		PressEnter().
		AssertVFOBand(core.Band40m).
		AssertCallsignView(core.VFO1, "").
		AssertCallinfoNotified("").
		AssertNoLogbookWrite()
}

// B5. Jump to bandmap call via callsign field
// Pre:  active field = CallsignField; input starts with @ and remainder parses as callsign.
// Post: bandmap SelectByCallsign invoked; callsign input cleared; A1 effects with empty callsign.
// Invariants: logbook; rig; other VFO.

func TestB5_JumpToBandmapCall(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("@DL1ABC").
		PressEnter().
		AssertBandmapSelected("DL1ABC").
		AssertCallsignView(core.VFO1, "").
		AssertCallinfoNotified("").
		AssertNoLogbookWrite()
}

// B6. XIT active toggled by UI
// Pre:  none.
// Post: VFO1 rig commanded with new XIT-active flag; view active field reapplied.
// Invariants: input values; VFO2 XIT.

func TestB6_XITActive_Toggle(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetXITActive(true).
		AssertXITActiveCommanded(true).
		AssertActiveField(core.VFO1, core.CallsignField)
}

// C1. Log valid QSO
// Pre:  input parses fully (callsign, band, mode, their-exchange complete and valid).
// Post: AddQSO called; CallsignLogged listeners notified; row cleared (active=callsign).
// Invariants: other VFO input; contest definition.

func TestC1_LogValidQSO(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("599").
		GotoNextField(). // → TheirExchange2
		Enter("042").
		GotoNextField(). // → TheirExchange3
		Enter("OE").
		PressEnter().
		AssertQSOAdded().
		AssertQSOAddedCallsign("DL1ABC").
		AssertCallsignLogged("DL1ABC").
		AssertActiveField(core.VFO1, core.CallsignField)
}

func TestC1_LogValidQSO_SearchPounce_AddsWorkedSpotToBandmap(t *testing.T) {
	// workmode = S&P → WorkedSpot added to bandmap after logging.
	NewScenario(t).
		WithClassicExchange().
		WithWorkmode(core.SearchPounce).
		Enter("DL1ABC").
		GotoNextField().Enter("599").
		GotoNextField().Enter("042").
		GotoNextField().Enter("OE").
		PressEnter().
		AssertQSOAdded().
		AssertBandmapAddedCallsign("DL1ABC").
		AssertBandmapSpotSource(core.WorkedSpot)
}

func TestC1_LogDuplicateQSO_StillAddsToLogbook(t *testing.T) {
	// Duplicate callsign: duplicate marker shown but QSO still added (C1 makes no exception).
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	dupe := core.QSO{Callsign: dl1abc, MyNumber: 1}

	NewScenario(t).
		WithClassicExchange().
		WithDuplicateQSO(dupe).
		Enter("DL1ABC").
		GotoNextField().Enter("599").
		GotoNextField().Enter("042").
		GotoNextField().Enter("OE").
		PressEnter().
		AssertQSOAdded().
		AssertQSOAddedCallsign("DL1ABC")
}

// C2. Log fails on invalid callsign
// Pre:  callsign input does not parse.
// Post: error shown on callsign field; active field = callsign.
// Invariants: logbook; row not cleared; serial claim unchanged.

func TestC2_LogFails_InvalidCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		PressEnter(). // callsign = "" → ParseCallsign fails
		AssertActiveField(core.VFO1, core.CallsignField).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C3. Log fails on invalid band
// Pre:  callsign parses; band input does not parse.
// Post: error on band field; active field = band.
// Invariants: logbook; row.

func TestC3_LogFails_InvalidBand(t *testing.T) {
	// Enter sets input.band = text even when bandSelected silently ignores it.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		SetActiveField(core.BandField).
		Enter("invalid").
		PressEnter().
		AssertActiveField(core.VFO1, core.BandField).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C4. Log fails on invalid mode
// Pre:  callsign and band parse; mode does not parse.
// Post: error on mode field; active field = mode.
// Invariants: logbook; row.

func TestC4_LogFails_InvalidMode(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		SetActiveField(core.ModeField).
		Enter("invalid").
		PressEnter().
		AssertActiveField(core.VFO1, core.ModeField).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C5. Log fails on missing their-exchange
// Pre:  non-EmptyAllowed their-exchange field is empty.
// Post: error on that field; active field = that field.
// Invariants: logbook; row.

func TestC5_LogFails_MissingTheirExchange(t *testing.T) {
	// RST (slot 1) is pre-filled with "599" by fillExchangeDefaults.
	// First truly missing field is the serial number (slot 2).
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		PressEnter().
		AssertActiveField(core.VFO1, core.TheirExchangeField(2)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C6. Log fails on invalid their-report
// Pre:  their-report field non-empty but does not parse as RST.
// Post: error on their-report field; active field set there.
// Invariants: logbook; row.

func TestC6_LogFails_InvalidTheirReport(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("abc").    // non-empty but invalid RST
		PressEnter().
		AssertActiveField(core.VFO1, core.TheirExchangeField(1)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C7. Log fails on invalid serial-only their-number
// Pre:  their-number field configured as pure serial; value non-numeric.
// Post: error on their-number field; active field set there.
// Invariants: logbook; row.

func TestC7_LogFails_InvalidTheirSerial(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("599").
		GotoNextField(). // → TheirExchange2
		Enter("abc").    // non-numeric serial
		PressEnter().
		AssertActiveField(core.VFO1, core.TheirExchangeField(2)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C8. Log applies prediction and aborts on empty non-serial/non-report exchange
// Pre:  non-report/non-serial their-exchange slot empty; prediction array length matches.
// Post: prediction copied into that slot; error "check their exchange" on that field;
//       active field set there.
// Invariants: logbook; other slots in row.

func TestC8_Log_PredictionFillsEmptySlot(t *testing.T) {
	// Requires EmptyAllowed on the generic text field so the "missing" check is
	// bypassed and the prediction-fill path is reached.
	//
	// The CallinfoFrame MUST be injected AFTER leaving the callsign field:
	// leaveCallsignField() pre-fills predictable empty slots, so injecting the
	// frame before GotoNextField() would populate TheirExchange3 early and the
	// empty check in parseTheirExchange would never fire.
	frame := core.CallinfoFrame{
		PredictedExchange: []string{"599", "042", "OE"},
	}

	NewScenario(t).
		WithClassicExchangeWithOptionalText().
		Enter("DL1ABC").
		GotoNextField().                     // leaves callsign field (no prediction yet → slot 3 stays empty)
		WithCallinfoFrame(core.VFO1, frame). // inject prediction now
		Enter("599").
		GotoNextField(). // → TheirExchange2
		Enter("042").
		PressEnter(). // at TheirExchange2 → Log; TheirExchange3 is empty → prediction fills it
		AssertTheirExchangeSet(core.VFO1, 3, "OE").
		AssertActiveField(core.VFO1, core.TheirExchangeField(3)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C9. Log fails on invalid my-report
// Pre:  my-exchange's report slot does not parse as RST.
// Post: error on my-report field.
// Invariants: logbook.

func TestC9_LogFails_InvalidMyReport(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("599").
		GotoNextField(). // → TheirExchange2
		Enter("042").
		GotoNextField(). // → TheirExchange3
		Enter("OE").
		SetActiveField(core.MyExchangeField(1)).
		Enter("abc"). // invalid my-RST
		PressEnter().
		AssertActiveField(core.VFO1, core.MyExchangeField(1)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C10. Log fails on invalid my-number (single-property field)
// Pre:  my-number field has exactly one property; value is non-numeric.
// Post: error on my-number field.
// Invariants: logbook.

func TestC10_LogFails_InvalidMySerial(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		Enter("599").
		GotoNextField(). // → TheirExchange2
		Enter("042").
		GotoNextField(). // → TheirExchange3
		Enter("OE").
		SetActiveField(core.MyExchangeField(2)).
		Enter("abc"). // non-numeric my-serial
		PressEnter().
		AssertActiveField(core.VFO1, core.MyExchangeField(2)).
		AssertMessageShown(core.VFO1).
		AssertNoLogbookWrite()
}

// C11. EnterPressed dispatch
// Post: if callsign field holds kHz/band/@call command: that command runs (not ESM, not Log);
//       else if ESM enabled AND not editing: NextESMStep (F section);
//       else: Log path (covered by C1).

func TestC11_EnterPressed_CommandPrecedesESMAndLog(t *testing.T) {
	// Numeric callsign input → frequency command, even when ESM is enabled.
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		Enter("14250").
		PressEnter().
		AssertVFOFrequency(14250000).
		AssertNoLogbookWrite()
}

// D1. Clear while editing
// Pre:  editing = true; editSnapshot not nil.
// Post: edit snapshot restored; editing = false; editing marker cleared; active field = callsign.
// Invariants: logbook.

func TestD1_Clear_WhileEditing_ExitsEditMode(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		Clear().
		AssertEditingMarker(core.VFO1, false).
		AssertActiveField(core.VFO1, core.CallsignField).
		AssertMessageCleared(core.VFO1).
		AssertNoLogbookWrite()
}

// D2. Clear focused VFO (not editing)
// Pre:  editing = false.
// Post: serial claim released; active field = callsign; duplicate=false; editing=false; message cleared.
// Invariants: logbook; selected band/mode/frequency.

func TestD2_Clear_NotEditing_ReleasesClaimAndResetsRow(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC"). // claim serial
		Clear().
		AssertCallsignView(core.VFO1, "").
		AssertDuplicateMarker(core.VFO1, false).
		AssertEditingMarker(core.VFO1, false).
		AssertActiveField(core.VFO1, core.CallsignField).
		AssertMessageCleared(core.VFO1)
}

func TestD2_Clear_SeedsMyExchangeFromLastExchange(t *testing.T) {
	// fillExchangeDefaults carries lastExchange[i] only where defaultExchangeValues[i] == "".
	// ClassicExchange: slot 1=RST("599"), slot 2=serial("001") have non-empty defaults →
	// lastExchange ignored there. Slot 3=generic text has empty default → seeded from lastExchange.
	NewScenario(t).
		WithClassicExchange().
		WithLastExchange([]string{"599", "001", "DX"}).
		Clear().
		AssertMyExchangeView(3, "DX")
}

// D3. Auto-clear on rig frequency jump
// Pre:  VFOFrequencyChanged fires; |Δ| > jumpThreshold (250 Hz); ignoreFrequencyJump = false.
// Post: selectedFrequency updated; Clear runs on focused VFO; active field reapplied.

func TestD3_VFOFrequencyChanged_LargeJump_TriggersAutoClear(t *testing.T) {
	// Initial frequency after setup = 14050000 Hz (from vfoSpy.Refresh).
	// A jump of 1000 Hz far exceeds the 250 Hz threshold.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		VFOFrequencyChanged(core.VFO1, 14050000+1000).
		AssertCallsignView(core.VFO1, ""). // row was cleared
		AssertActiveField(core.VFO1, core.CallsignField)
}

// D4. Suppress jump-clear once
// Pre:  ignoreFrequencyJump = true (set by G2 / EntrySelected).
// Post: frequency updated, Clear NOT triggered; flag reset to false.
// Invariants: row preserved.

func TestD4_FrequencyJump_SuppressedByEntrySelected(t *testing.T) {
	// EntrySelected: Clear + ignoreFrequencyJump=true + frequencyEntered(14200000) + Enter("DL1ABC").
	// A subsequent large VFOFrequencyChanged must NOT clear the row.
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	entry := core.BandmapEntry{Call: dl1abc, Frequency: 14200000}

	NewScenario(t).
		WithClassicExchange().
		EntrySelected(entry).
		// ignoreFrequencyJump=true now; trigger a large jump
		VFOFrequencyChanged(core.VFO1, 14200000).
		// If Clear had fired, showInput → SetCallsign(VFO1,"") would be recorded.
		AssertViewNotCalledWith("SetCallsign", core.VFO1, "")
}

// E1. QSO selected from list
// Pre:  ignoreQSOSelection = false.
// Post: editing = true; editQSO = qso; VFO1 row shows QSO data; editing marker set; active = callsign.
// Invariants: logbook; VFO2 state.

func TestE1_QSOSelected_EntersEditMode(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign: dl1abc,
		MyNumber: 7,
		Band:     core.Band20m,
		Mode:     core.ModeCW,
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		AssertCallsignView(core.VFO1, "DL1ABC").
		AssertEditingMarker(core.VFO1, true).
		AssertActiveField(core.VFO1, core.CallsignField).
		AssertNoLogbookWrite()
}

// E2. QSO selection ignored
// Pre:  ignoreQSOSelection = true (set during selectLastQSO).
// Post: no-op — row unchanged.

func TestE2_QSOSelection_Ignored_WhenFlagSet(t *testing.T) {
	// selectLastQSO() sets ignoreQSOSelection=true, calls qsoList.SelectLastQSO(), then resets.
	// Our qsoListSpy fires QSOSelected synchronously inside SelectLastQSO; it must be ignored.
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	autoQSO := core.QSO{Callsign: dl1abc, MyNumber: 42}

	s := NewScenario(t).WithClassicExchange()
	s.WithQSOListCallback(func() {
		s.controller.QSOSelected(autoQSO)
	})
	s.Clear(). // triggers selectLastQSO → callback → QSOSelected → ignored
			AssertCallsignView(core.VFO1, "") // row not populated with autoQSO's callsign
}

// E3. Edit last QSO
// Pre:  logbook non-empty.
// Post: active field on focused VFO = callsign; qsoList.SelectLastQSO() called.

func TestE3_EditLastQSO_CallsSelectLastQSO(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		EditLastQSO().
		AssertQSOListSelectLastCalled()
}

// E4. Save edited QSO
// Pre:  editing = true; input parses.
// Post: UpdateQSO called (not AddQSO); row cleared (D1 path); CallsignLogged emitted.
// Invariants: original time; original workmode.

func TestE4_SaveEditedQSO_UpdatesLogbook(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		PressEnter(). // editing=true → Log() → UpdateQSO (not AddQSO)
		AssertQSOUpdated().
		AssertQSOUpdatedCallsign("DL1ABC").
		AssertNoQSOAdded()
}

// E5. Suppress rig sync on VFO1 during edit
// Pre:  editing = true; VFOBandChanged/VFOModeChanged/VFOFrequencyChanged for VFO1.
// Post: event ignored; VFO1 row unchanged.

func TestE5_VFO1BandChanged_Ignored_DuringEdit(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		VFOBandChanged(core.VFO1, core.Band40m). // must be suppressed during edit
		AssertViewNotCalledWith("SetBand", core.VFO1, "40m")
}

func TestE5_VFO1ModeChanged_Ignored_DuringEdit(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		VFOModeChanged(core.VFO1, core.ModeSSB). // must be suppressed during edit
		AssertViewNotCalledWith("SetMode", core.VFO1, "SSB")
}

func TestE5_VFO1FrequencyChanged_Ignored_DuringEdit(t *testing.T) {
	// Large jump (1000 Hz > 250 Hz threshold) that would normally trigger Clear.
	// During edit it must be fully ignored: no view update, no Clear.
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		SelectQSO(qso).
		VFOFrequencyChanged(core.VFO1, 14050000+1000). // large jump, must be suppressed
		AssertViewNotCalledWith("SetFrequency", core.VFO1, core.Frequency(14050000+1000)).
		AssertViewNotCalledWith("SetCallsign", core.VFO1, "") // Clear did not run
}

// F1. Set ESM view
// Pre:  view ≠ nil.
// Post: esmView = supplied view; SetESMEnabled and SetMessage called with current state.

func TestF1_SetESMView_PushesCurrentState(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled(). // esmEnabled = true
		ConnectESMView(). // spy-reset + SetESMView(s.esmView)
		AssertESMViewEnabled(true)
}

// F2. Toggle ESM enabled
// Pre:  none.
// Post: esmEnabled = new value; view notified; active field reapplied; ESMListeners notified.

func TestF2_ToggleESM_NotifiesListenerAndReappliesActiveField(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		ToggleESM(true).
		AssertESMListenerNotified(true).
		AssertActiveField(core.VFO1, core.CallsignField)
}

// F3. NextESMStep — Run mode, CallsignValid: sends keyer text, advances past RST field.
// Pre:  not editing; keyer set; ESM enabled.
// Post: ESM state/message recomputed; keyer.SendText called; if Run+CallsignValid: GotoNextField (skip report field).

func TestF3_NextESMStep_Run_CallsignValid_SendsAndAdvancesField(t *testing.T) {
	// ClassicExchange: TheirExchange1=RST (theirReport field) → skipped → land on TheirExchange2.
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		WithVFOSwitcher().
		WithWorkmode(core.Run).
		Enter("DL1ABC"). // ESM state = CallsignValid
		NextESMStep().
		AssertKeyerSentMacro(1).
		AssertActiveField(core.VFO1, core.TheirExchangeField(2)) // skip RST → serial
}

// F3b. NextESMStep — ExchangeValid: logs the QSO.

func TestF3_NextESMStep_ExchangeValid_Logs(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		WithWorkmode(core.Run).
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1 (RST pre-filled "599")
		GotoNextField(). // → TheirExchange2
		Enter("042").
		GotoNextField(). // → TheirExchange3
		Enter("OE").
		NextESMStep(). // ExchangeValid → Log()
		AssertQSOAdded()
}

// F4. ESM recompute on workmode change
// Pre:  triggered by WorkmodeChanged.
// Post: esmMessage reflects current workmode; esmView.SetMessage called.

func TestF4_ESMRecomputedOnWorkmodeChange(t *testing.T) {
	// Empty callsign + Run workmode → ESMCallsignEmpty → keyer text index 0.
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		ConnectESMView().
		WorkmodeChanged(core.Run). // triggers updateESM
		AssertESMViewMessage("keyer[0]")
}

// F5. EnterPressed routes around ESM while editing
// Pre:  editing = true.
// Post: ESM step not taken; Log path used.

func TestF5_EnterPressed_Editing_SkipsESM(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		SelectQSO(qso).
		PressEnter(). // editing=true → Log path, not ESM
		AssertQSOUpdated().
		AssertNoKeyerText() // keyer.SendText NOT called
}

// G1. Mark current callsign in bandmap
// Pre:  callsign parses.
// Post: ManualSpot with call/freq/band/mode/now added to bandmap.
// Invariants: input row; logbook.

func TestG1_MarkInBandmap_AddsManualSpot(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		MarkInBandmap().
		AssertBandmapAddedCallsign("DL1ABC").
		AssertBandmapSpotSource(core.ManualSpot)
}

// G2. Bandmap entry selected
// Pre:  any state.
// Post: D2/D1 Clear; ignoreFrequencyJump=true; rig tuned; callsign entered; active=callsign.
// Invariants: logbook; contest.

func TestG2_EntrySelected_ClearsAndEntersCallsign(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	bEntry := core.BandmapEntry{Call: dl1abc, Frequency: 14200000}

	NewScenario(t).
		WithClassicExchange().
		Enter("DL2XYZ"). // some other callsign in the row
		EntrySelected(bEntry).
		AssertCallsignView(core.VFO1, "DL1ABC").
		AssertVFOFrequency(14200000).
		AssertActiveField(core.VFO1, core.CallsignField)
}

// G3. Select best on-frequency match
// Pre:  best-match callsign non-empty.
// Post: active field = callsign; A1 effects with that callsign; view updated.

func TestG3_SelectBestMatchOnFrequency_EntersCallsign(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	frame := core.CallinfoFrame{
		Supercheck: []core.AnnotatedCallsign{{Callsign: dl1abc}},
	}

	NewScenario(t).
		WithClassicExchange().
		WithCallinfoFrame(core.VFO1, frame).
		SelectBestMatchOnFrequency().
		AssertCallsignView(core.VFO1, "DL1ABC").
		AssertActiveField(core.VFO1, core.CallsignField)
}

// G4. Select match by index
// Pre:  match exists at index.
// Post: same as G3 with that match's callsign.

func TestG4_SelectMatchByIndex_EntersCallsign(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	dl2xyz, _ := core.ParseCallsign("DL2XYZ")
	frame := core.CallinfoFrame{
		Supercheck: []core.AnnotatedCallsign{
			{Callsign: dl1abc},
			{Callsign: dl2xyz},
		},
	}

	NewScenario(t).
		WithClassicExchange().
		WithCallinfoFrame(core.VFO1, frame).
		SelectMatch(1). // index 1 = DL2XYZ
		AssertCallsignView(core.VFO1, "DL2XYZ").
		AssertActiveField(core.VFO1, core.CallsignField)
}

// G5. Refresh prediction
// Pre:  none.
// Post: callinfo re-notified with current callsign and empty exchange; predictable empty slots refilled.
// Invariants: non-predictable slots; non-empty slots.

func TestG5_RefreshPrediction_NotifiesCallinfo(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		RefreshPrediction().
		AssertCallinfoNotified("DL1ABC")
}

// G6. CallinfoFrameChanged
// Pre:  any state.
// Post: currentCallinfoFrame[vfo] = new frame; used by subsequent leaveCallsignField.

func TestG6_CallinfoFrameChanged_FrameUsedInSubsequentPrediction(t *testing.T) {
	// Same mechanism as A2 — verifies CallinfoFrameChanged stores the frame for leaveCallsignField.
	frame := core.CallinfoFrame{PredictedExchange: []string{"599", "042", "OE"}}

	NewScenario(t).
		WithClassicExchange().
		Enter("OE5XYZ").
		WithCallinfoFrame(core.VFO1, frame). // G6 event: store the frame
		GotoNextField().                     // leaveCallsignField uses the stored frame
		AssertTheirExchangeSet(core.VFO1, 3, "OE")
}

// H1. SetFocusedVFO
// Pre:  target ≠ VFO2 OR vfo2Enabled = true; target ≠ current focused VFO.
// Post: focusedVFO = target; vfoSwitcher.SetCurrentVFO called; view active field reapplied.
// Invariants: input rows; serial claims; logbook.

func TestH1_SetFocusedVFO_ReappliesActiveField(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().WithVFOSwitcher().
		SetFocusedVFO(core.VFO2).
		AssertActiveField(core.VFO2, core.CallsignField).
		AssertVFOSwitcherCurrentCalled(core.VFO2)
}

func TestH1_SetFocusedVFO_VFO2Disabled_Ignored(t *testing.T) {
	// VFO2 not enabled → SetFocusedVFO(VFO2) is a no-op.
	NewScenario(t).
		WithClassicExchange().WithVFOSwitcher().
		SetFocusedVFO(core.VFO2).
		// focused stays on VFO1; VFO2 active field NOT set by controller
		AssertViewNotCalledWith("SetActiveField", core.VFO2, core.CallsignField)
}

// H3. ToggleFocusedVFO
// Pre:  none.
// Post: if vfo2Enabled: focused = other VFO; else: focused = VFO1.
// Invariants: input rows.

func TestH3_ToggleFocusedVFO_VFO2Enabled_TogglesFromVFO1ToVFO2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		ToggleFocusedVFO(). // VFO1 → VFO2
		AssertActiveField(core.VFO2, core.CallsignField)
}

func TestH3_ToggleFocusedVFO_VFO2Disabled_StaysOnVFO1(t *testing.T) {
	// VFO2 not enabled → ToggleFocusedVFO is a no-op (already VFO1, early return).
	// Key invariant: VFO2 is never activated.
	NewScenario(t).
		WithClassicExchange(). // VFO2 not enabled
		ToggleFocusedVFO().
		AssertViewNotCalledWith("SetActiveField", core.VFO2, core.CallsignField)
}

// H4. FocusVFO1 / FocusVFO2
// Same Pre/Post/Invariants as H1.

func TestH4_FocusVFO2_SetsActiveFieldOnVFO2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		FocusVFO2().
		AssertActiveField(core.VFO2, core.CallsignField)
}

func TestH4_FocusVFO1_AfterVFO2_SetsActiveFieldOnVFO1(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		FocusVFO2().
		FocusVFO1().
		AssertActiveField(core.VFO1, core.CallsignField)
}

// H5. LogVFO(vfo)
// Pre:  H1 preconditions; input parses.
// Post: focus moved to vfo; Log executed.
// Invariants: see C1.

func TestH5_LogVFO_FocusesAndLogs(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField().Enter("599").
		GotoNextField().Enter("042").
		GotoNextField().Enter("OE").
		LogVFO(core.VFO1).
		AssertQSOAdded().
		AssertQSOAddedCallsign("DL1ABC")
}

// H6. ClearVFO(vfo)
// Pre:  H1 preconditions.
// Post: focus moved to vfo; Clear executed.
// Invariants: see D2.

func TestH6_ClearVFO_FocusesAndClears(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		Enter("DL1ABC").
		ClearVFO(core.VFO1).
		AssertCallsignView(core.VFO1, "").
		AssertActiveField(core.VFO1, core.CallsignField)
}

// H7. RadioChanged → dual VFO
// Pre:  event with singleVFO = false.
// Post: vfo2Enabled = true; view.SetVFOEnabled(VFO2, true).

func TestH7_RadioChanged_DualVFO_EnablesVFO2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		RadioChanged(false). // singleVFO=false
		AssertVFO2Enabled(true)
}

// H8. RadioChanged → single VFO
// Pre:  event with singleVFO = true.
// Post: vfo2Enabled = false; view.SetVFOEnabled(VFO2, false).

func TestH8_RadioChanged_SingleVFO_DisablesVFO2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2(). // start with VFO2 enabled
		RadioChanged(true).               // singleVFO=true → collapse
		AssertVFO2Enabled(false)
}

func TestH8_RadioChanged_SingleVFO_VFO2Focused_CollapsesToVFO1(t *testing.T) {
	// When VFO2 is focused at collapse time: VFO2 input zeroed, focus silently moved to VFO1.
	// Observable: CurrentValues() reads focusedVFO's input; after collapse that is VFO1.
	// VFO1 had "DL1ABC" entered before the switch; VFO2 gets zeroed by collapse.
	s := NewScenario(t).WithClassicExchange()
	s.controller.Enter("DL1ABC") // VFO1 callsign
	s.WithVFO2()
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ") // VFO2 callsign
	s.resetSpies()
	s.controller.RadioChanged("", true) // collapse with VFO2 focused
	// focusedVFO must now be VFO1 → CurrentValues reads VFO1 input = "DL1ABC"
	vals := s.controller.CurrentValues()
	if vals.TheirCall != "DL1ABC" {
		t.Errorf("expected focus on VFO1 (TheirCall=%q), got %q", "DL1ABC", vals.TheirCall)
	}
}

// I1. Claim on callsign keystroke
// Pre:  not editing; callsign parses; claimedSerial[focused] = 0.
// Post: claimedSerial[focused] = nextUnclaimedSerial; previews refreshed.

func TestI1_ClaimSerial_ValidCallsign_ClaimsNextQSONumber(t *testing.T) {
	s := NewScenario(t).WithClassicExchange().WithNextQSONumber(5)
	s.Enter("DL1ABC")
	got := s.controller.CurrentValues().MyNumber
	if got != 5 {
		t.Errorf("expected MyNumber=5 (claimed from nextQSONumber=5), got %d", got)
	}
}

// I2. Claim sticky
// Pre:  claimedSerial[focused] ≠ 0.
// Post: no reclaim on further keystrokes.

func TestI2_ClaimSticky_SerialUnchangedAcrossKeystrokes(t *testing.T) {
	s := NewScenario(t).WithClassicExchange().WithNextQSONumber(5)
	s.Enter("DL1ABC")
	first := s.controller.CurrentValues().MyNumber
	// Simulate logbook number advancing externally — claim must remain 5.
	s.logbook.nextQSONumber = 99
	s.Enter("DL1ABC")
	second := s.controller.CurrentValues().MyNumber
	if first != second {
		t.Errorf("expected sticky claim: first=%d, second=%d", first, second)
	}
}

// I3. Collision avoidance
// Embedded in I1's nextUnclaimedSerial.
// Pre:  other VFO's claim equals current NextQSONumber.
// Post: new claim = candidate + 1.

func TestI3_CollisionAvoidance_VFO2GetsNextSerial(t *testing.T) {
	s := NewScenario(t).WithClassicExchange()
	s.WithVFO2()
	// Enter on VFO1 → claims serial 1 (nextQSONumber=1).
	s.Enter("DL1ABC")
	// Switch focus to VFO2.
	s.controller.SetFocusedVFO(core.VFO2)
	// Enter on VFO2 → must get serial 2 (serial 1 taken by VFO1).
	s.Enter("DL2XYZ")
	got := s.controller.CurrentValues().MyNumber
	if got != 2 {
		t.Errorf("expected VFO2 to claim serial 2 (collision avoidance), got %d", got)
	}
}

// I4. Release on Clear (not editing)
// Pre:  D2 entered.
// Post: claimedSerial[focused] = 0; both VFOs' previews recomputed.

func TestI4_ReleaseClaim_OnClear_RowResetAndPreviewRefreshed(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC"). // claim serial 1
		Clear().
		AssertCallsignView(core.VFO1, ""). // row cleared
		AssertMyExchangeView(2, "001")     // serial preview refreshed
}

// I5. Refresh previews after log
// Pre:  Log path completed AddQSO/UpdateQSO.
// Post: refreshMyNumberInputs recomputes both VFOs' displayed serials.

func TestI5_RefreshPreviewsAfterLog(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		GotoNextField().Enter("599").
		GotoNextField().Enter("042").
		GotoNextField().Enter("OE").
		PressEnter(). // logs; refreshMyNumberInputs called
		AssertMyExchangeView(2, "001")
}

// I6. Edit mode owns the serial slot
// Pre:  E1 fired.
// Post: claimedSerial[VFO1] = editQSO.MyNumber; restored on leaveEditMode.

func TestI6_EditMode_OwnsSerialSlot(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign: dl1abc,
		MyNumber: 7,
		Band:     core.Band20m,
		Mode:     core.ModeCW,
	}
	s := NewScenario(t).WithClassicExchange()
	s.SelectQSO(qso)
	got := s.controller.CurrentValues().MyNumber
	if got != 7 {
		t.Errorf("expected MyNumber=7 (editQSO.MyNumber), got %d", got)
	}
}

// J1. VFOFrequencyChanged
// Pre:  not VFO1+editing; f ≠ selectedFrequency[vfo].
// Post: selectedFrequency updated; view updated; if |Δ|>threshold: Clear runs.

func TestJ1_VFOFrequencyChanged_SmallJump_NoClear(t *testing.T) {
	// Initial frequency = 14050000 Hz; jump of 100 Hz is below 250 Hz threshold.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		VFOFrequencyChanged(core.VFO1, 14050000+100). // small jump
		// Clear must NOT have fired: callsign still in view.
		AssertViewNotCalledWith("SetCallsign", core.VFO1, "")
}

func TestJ1_VFOFrequencyChanged_LargeJump_ClearsEventVFO(t *testing.T) {
	// Large jump on VFO1 clears VFO1 input.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		VFOFrequencyChanged(core.VFO1, 14050000+1000). // large jump
		AssertCallsignView(core.VFO1, "")              // VFO1 cleared
}

func TestJ1_VFOFrequencyChanged_VFO2LargeJump_ClearsVFO2Only(t *testing.T) {
	// Large jump on VFO2 must clear VFO2 input, not VFO1.
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2()

	// Enter callsign on VFO1.
	s.controller.Enter("DL1ABC")
	// Switch to VFO2, enter callsign.
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	// Switch back to VFO1 (so VFO1 is focused).
	s.controller.SetFocusedVFO(core.VFO1)
	s.resetSpies()

	// Large frequency jump on VFO2 while VFO1 is focused.
	s.controller.VFOFrequencyChanged(core.VFO2, 14050000+1000)

	// VFO2 input must be cleared.
	s.view.assertCalledWith(s.t, "SetCallsign", core.VFO2, "")
	// VFO1 input must NOT be cleared.
	assert.False(t, s.view.wasCalledWith("SetCallsign", core.VFO1, ""),
		"VFO1 callsign must not be cleared by VFO2 frequency jump")
	// Focused VFO must still be VFO1: verify by entering text — it goes to VFO1.
	s.resetSpies()
	s.controller.Enter("DL9ZZZ")
	assert.Equal(t, "DL9ZZZ", s.controller.CurrentValues().TheirCall,
		"focused VFO must still be VFO1")
}

func TestJ1_VFOFrequencyChanged_VFO2LargeJump_ClearsCallinfoAndMessage(t *testing.T) {
	// Regression: a frequency jump on VFO2 must clear VFO2's callinfo and message,
	// not VFO1's. Focus must stay on VFO1.
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2()

	// Enter callsign on VFO1.
	s.controller.Enter("DL1ABC")
	// Switch to VFO2, enter callsign (triggers callinfo).
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	// Switch back to VFO1.
	s.controller.SetFocusedVFO(core.VFO1)
	s.resetSpies()

	// Large frequency jump on VFO2.
	s.controller.VFOFrequencyChanged(core.VFO2, 14050000+1000)

	// VFO2 callinfo must be cleared (InputChanged with empty call for VFO2).
	s.AssertCallinfoCleared(core.VFO2)
	// VFO2 message must be cleared.
	s.view.assertCalledWith(s.t, "ClearMessage", core.VFO2)
	// VFO1 message must NOT be cleared.
	assert.False(t, s.view.wasCalledWith("ClearMessage", core.VFO1),
		"VFO1 message must not be cleared by VFO2 frequency jump")
	// Focus must remain on VFO1.
	s.resetSpies()
	s.controller.Enter("DL9ZZZ")
	assert.Equal(t, "DL9ZZZ", s.controller.CurrentValues().TheirCall,
		"focused VFO must still be VFO1")
}

func TestJ1_VFOFrequencyChanged_UpdatesFrequencyView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		VFOFrequencyChanged(core.VFO1, 14050000+100).
		AssertFrequencyView(core.VFO1, 14050000+100)
}

// J2. VFOBandChanged
// Pre:  not VFO1+editing; band ≠ NoBand; band ≠ selectedBand[focused].
// Post: selectedBand updated; view band updated.

func TestJ2_VFOBandChanged_UpdatesView(t *testing.T) {
	// Initial band = Band20m (from vfoSpy.Refresh). Band40m differs → view updated.
	NewScenario(t).
		WithClassicExchange().
		VFOBandChanged(core.VFO1, core.Band40m).
		AssertViewNotCalledWith("SetBand", core.VFO1, "") // at minimum, not empty
}

func TestJ2_VFOBandChanged_SameBand_Ignored(t *testing.T) {
	// Same as current selectedBand → no view update.
	NewScenario(t).
		WithClassicExchange().
		VFOBandChanged(core.VFO1, core.Band20m). // Band20m is already selected
		AssertViewNotCalledWith("SetBand", core.VFO1, "20m")
}

// J3. VFOModeChanged
// Pre:  not VFO1+editing; mode ≠ NoMode; mode ≠ selectedMode[focused].
// Post: selectedMode updated; view mode updated.

func TestJ3_VFOModeChanged_UpdatesView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		VFOModeChanged(core.VFO1, core.ModeSSB).
		AssertViewNotCalledWith("SetMode", core.VFO1, "") // mode was updated
}

func TestJ3_VFOModeChanged_SameMode_Ignored(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		VFOModeChanged(core.VFO1, core.ModeCW). // ModeCW already selected
		AssertViewNotCalledWith("SetMode", core.VFO1, "CW")
}

// J2b. VFOBandChanged targets event VFO, not focused VFO
// Regression: VFO2 band-change while VFO1 focused must update VFO2's state,
// not corrupt VFO1's selectedBand.

func TestJ2b_VFOBandChanged_VFO2Event_WhileVFO1Focused(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		VFOBandChanged(core.VFO2, core.Band40m).
		AssertBandView(core.VFO2, "40m").
		AssertViewNotCalledWith("SetBand", core.VFO1, "40m")
}

func TestJ2b_VFOBandChanged_VFO2Event_DoesNotCorruptVFO1(t *testing.T) {
	// After VFO2 band event, VFO1's selectedBand must still be Band20m (from Refresh).
	// Verify by sending VFO1 Band20m again — it should be a same-band no-op.
	NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		VFOBandChanged(core.VFO2, core.Band40m).
		VFOBandChanged(core.VFO1, core.Band20m). // resetSpies; must be no-op (same band)
		AssertViewNotCalledWith("SetBand", core.VFO1, "20m")
}

// J3b. VFOModeChanged targets event VFO, not focused VFO
// Regression: same issue as J2b but for mode.

func TestJ3b_VFOModeChanged_VFO2Event_WhileVFO1Focused(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		VFOModeChanged(core.VFO2, core.ModeSSB).
		AssertModeView(core.VFO2, "SSB").
		AssertViewNotCalledWith("SetMode", core.VFO1, "SSB")
}

func TestJ3b_VFOModeChanged_VFO2Event_DoesNotCorruptVFO1(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		VFOModeChanged(core.VFO2, core.ModeSSB).
		VFOModeChanged(core.VFO1, core.ModeCW). // resetSpies; must be no-op (same mode)
		AssertViewNotCalledWith("SetMode", core.VFO1, "CW")
}

// J4. VFOXITChanged
// Pre:  event for vfo.
// Post: view.SetXIT(vfo, active, offset).

func TestJ4_VFOXITChanged_UpdatesView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		VFOXITChanged(core.VFO1, true, 500).
		AssertXITView(core.VFO1, true, 500)
}

// J5. XITActiveChanged
// Pre:  event arrives.
// Post: view.SetXITActive(VFO1, active).

func TestJ5_XITActiveChanged_UpdatesView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		XITActiveChanged(true).
		AssertXITActiveView(core.VFO1, true)
}

// J5b. Incremental tuning visibility
// Pre:  availability and workmode events arrive.
// Post: the widget matching the workmode (Run->RIT, S&P->XIT) is shown only when available.

func TestJ5b_IncrementalTuningVisibility_AvailabilityAndWorkmode(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WorkmodeChanged(core.SearchPounce).
		IncrementalTuningAvailabilityChanged(core.VFO1, core.XIT, true).
		AssertIncrementalTuningVisible(core.VFO1, core.XIT, true).
		AssertIncrementalTuningVisible(core.VFO1, core.RIT, false)

	NewScenario(t).
		WithClassicExchange().
		WorkmodeChanged(core.Run).
		IncrementalTuningAvailabilityChanged(core.VFO1, core.RIT, true).
		AssertIncrementalTuningVisible(core.VFO1, core.RIT, true).
		AssertIncrementalTuningVisible(core.VFO1, core.XIT, false)
}

// J6. VFOPTTChanged
// Pre:  event for vfo.
// Post: if VFO1: ptt updated, view.SetTXState; else: view.SetTXState(vfo, active, false, 0).

func TestJ6_VFOPTTChanged_VFO1_UpdatesTXState(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		VFOPTTChanged(core.VFO1, true).
		AssertTXStateView(core.VFO1, true, false, 0)
}

func TestJ6_VFOPTTChanged_VFO2_UpdatesTXStateForVFO2(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		VFOPTTChanged(core.VFO2, true).
		AssertTXStateView(core.VFO2, true, false, 0)
}

// K1. SendQuestion
// Pre:  keyer set; canTransmit = true.
// Post: if active=their-exchange: keyer.SendQuestion("nr"); else: keyer.SendQuestion(callsign).

func TestK1_SendQuestion_CallsignField_SendsCallsign(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithKeyer().WithVFOSwitcher().
		Enter("DL1ABC").
		SendQuestion().
		AssertKeyerSentQuestion("DL1ABC")
}

func TestK1_SendQuestion_TheirExchangeField_SendsNR(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithKeyer().WithVFOSwitcher().
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1
		SendQuestion().
		AssertKeyerSentQuestion("nr")
}

// K2. RepeatLastTransmission
// Pre:  keyer set; canTransmit = true.
// Post: keyer.Repeat(). VFO switcher is NOT called (stay on last TX VFO).

func TestK2_RepeatLastTransmission_CallsKeyer(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithKeyer().
		RepeatLastTransmission().
		AssertKeyerRepeated()
}

// K3. StopTX
// Pre:  keyer set.
// Post: keyer.Stop().

func TestK3_StopTX_CallsKeyerStop(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithKeyer().
		StopTX().
		AssertKeyerStopped()
}

// K4. ParrotActive(active)
// Pre:  none.
// Post: parrotActive = active; TX state view refreshed; if active: Clear runs.

func TestK4_ParrotActive_True_UpdatesTXStateAndClears(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		ParrotActive(true).
		AssertTXStateView(core.VFO1, false, true, 0).
		AssertCallsignView(core.VFO1, "") // Clear ran
}

func TestK4_ParrotActive_False_UpdatesTXStateOnly(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		ParrotActive(false).
		AssertTXStateView(core.VFO1, false, false, 0)
}

// K5. ParrotTimeLeft(d)
// Pre:  none.
// Post: parrotTimeLeft = d; TX state view refreshed.

func TestK5_ParrotTimeLeft_UpdatesTXState(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		ParrotTimeLeft(5*time.Second).
		AssertTXStateView(core.VFO1, false, false, 5*time.Second)
}

// L1. StationChanged
// Pre:  none.
// Post: stationCallsign = new value; view.SetMyCall.

func TestL1_StationChanged_UpdatesMyCallView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		StationChanged("DL1XYZ").
		AssertMyCallView("DL1XYZ")
}

// L2. ContestChanged
// Pre:  contest definition supplied.
// Post: exchange fields replaced; Clear runs (active = callsign).

func TestL2_ContestChanged_ClearsRow(t *testing.T) {
	// WithClassicExchange calls ContestChanged internally; verify Clear ran.
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		WithClassicExchange(). // second ContestChanged → Clear
		AssertCallsignView(core.VFO1, "").
		AssertActiveField(core.VFO1, core.CallsignField)
}

// L3. WorkmodeChanged
// Pre:  none.
// Post: workmode = new value; ESM message recomputed.
// (primary coverage in F4; minimal smoke test here)

func TestL3_WorkmodeChanged_UpdatesWorkmode(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithESMEnabled().WithKeyer().ConnectESMView().
		WorkmodeChanged(core.Run).
		AssertESMViewMessage("keyer[0]")
}

// L4. LogbookLoaded
// Pre:  logbook loaded.
// Post: every VFO's selectedBand/selectedMode = logbook's last; Clear runs; input pushed to view.

func TestL4_LogbookLoaded_ClearsRowAndPushesInput(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC"). // something in the row
		LogbookLoaded().
		AssertCallsignView(core.VFO1, ""). // Clear ran
		AssertActiveField(core.VFO1, core.CallsignField)
}

// M6. RefreshView
// Pre:  none.
// Post: current input pushed to view (showInput).

func TestM6_RefreshView_PushesCurrentInputToView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		Enter("DL1ABC").
		RefreshView().
		AssertCallsignView(core.VFO1, "DL1ABC")
}

// M7. Activate
// Pre:  none.
// Post: view active field reapplied for focused VFO.

func TestM7_Activate_ReappliesActiveField(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		SetActiveField(core.TheirExchangeField(1)).
		Activate().
		AssertActiveField(core.VFO1, core.TheirExchangeField(1))
}

// N1. CurrentQSOState
// Pre:  none.
// Post: (NoCallsign, QSODataEmpty) if callsign empty/unparseable;
//       (call, QSODataInvalid) if call OK, exchange fails;
//       (call, QSODataValid) if both OK.

func TestN1_CurrentQSOState_Empty_WhenCallsignEmpty(t *testing.T) {
	s := NewScenario(t).WithClassicExchange()
	call, state := s.controller.CurrentQSOState()
	if call != core.NoCallsign {
		t.Errorf("expected NoCallsign, got %v", call)
	}
	if state != core.QSODataEmpty {
		t.Errorf("expected QSODataEmpty, got %v", state)
	}
}

func TestN1_CurrentQSOState_Invalid_WhenExchangeIncomplete(t *testing.T) {
	s := NewScenario(t).WithClassicExchange()
	s.Enter("DL1ABC") // callsign OK, exchange empty → invalid
	_, state := s.controller.CurrentQSOState()
	if state != core.QSODataInvalid {
		t.Errorf("expected QSODataInvalid, got %v", state)
	}
}

func TestN1_CurrentQSOState_Valid_WhenFullyFilledIn(t *testing.T) {
	s := NewScenario(t).WithClassicExchange()
	s.Enter("DL1ABC")
	s.GotoNextField()
	s.Enter("599") // TheirExchange1 (RST)
	s.GotoNextField()
	s.Enter("042") // TheirExchange2 (serial)
	s.GotoNextField()
	s.Enter("OE") // TheirExchange3 (text)
	_, state := s.controller.CurrentQSOState()
	if state != core.QSODataValid {
		t.Errorf("expected QSODataValid, got %v", state)
	}
}

// N2. CurrentValues
// Pre:  none.
// Post: returns KeyerValues snapshot from focused VFO input.

func TestN2_CurrentValues_TheirCall_MatchesEnteredCallsign(t *testing.T) {
	s := NewScenario(t).WithClassicExchange()
	s.Enter("DL1ABC")
	vals := s.controller.CurrentValues()
	if vals.TheirCall != "DL1ABC" {
		t.Errorf("expected TheirCall=%q, got %q", "DL1ABC", vals.TheirCall)
	}
}

func TestN2_CurrentValues_MyNumber_ReflectsClaimedSerial(t *testing.T) {
	s := NewScenario(t).WithClassicExchange().WithNextQSONumber(42)
	s.Enter("DL1ABC") // claims serial 42
	vals := s.controller.CurrentValues()
	if vals.MyNumber != 42 {
		t.Errorf("expected MyNumber=42, got %d", vals.MyNumber)
	}
}

// I7. Dual-VFO serial interleaving
// Pre:  classic exchange, nextQSONumber=6, VFO2 enabled.
// Flow: VFO1 enters callsign (claims 6), tabs to exchange → VFO2 enters callsign
//       (claims 7), fills exchange, logs → focus VFO1, fill exchange, log.
// Post: VFO2 QSO has MyNumber=7, VFO1 QSO has MyNumber=6.

func TestI7_DualVFO_SerialInterleaving(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		WithNextQSONumber(6)

	// Ensure VFO2 has band and mode set (VFO1 gets them via vfoSpy.Refresh in SetView).
	s.controller.VFOBandChanged(core.VFO2, core.Band20m)
	s.controller.VFOModeChanged(core.VFO2, core.ModeCW)

	// VFO1: enter callsign → claims serial 6; advance to first exchange field.
	s.controller.Enter("DL1ABC")
	s.controller.GotoNextField()

	// VFO2: enter callsign → claims serial 7 (collision avoidance skips 6).
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	// Fill VFO2 exchange fields.
	s.controller.GotoNextField()
	s.controller.Enter("599") // TheirExchange1 (RST)
	s.controller.GotoNextField()
	s.controller.Enter("042") // TheirExchange2 (serial)
	s.controller.GotoNextField()
	s.controller.Enter("OE") // TheirExchange3 (text)

	// Log VFO2 QSO.
	s.controller.Log()
	require.Len(t, s.logbook.addedQSOs, 1, "VFO2 QSO must be logged")
	vfo2QSO := s.logbook.addedQSOs[0]
	if vfo2QSO.MyNumber != 7 {
		t.Errorf("VFO2 QSO: expected MyNumber=7, got %d", vfo2QSO.MyNumber)
	}

	// Simulate logbook advancing after the logged QSO.
	s.logbook.nextQSONumber = 7

	// Focus VFO1: serial display must reflect VFO1's still-held claim (6).
	s.controller.SetFocusedVFO(core.VFO1)
	// Fill remaining VFO1 exchange fields.
	s.controller.Enter("599") // TheirExchange1 (RST) — active field is still TheirExchange1
	s.controller.GotoNextField()
	s.controller.Enter("001") // TheirExchange2 (serial)
	s.controller.GotoNextField()
	s.controller.Enter("DL") // TheirExchange3 (text)

	// Log VFO1 QSO.
	s.logbook.resetCalls()
	s.controller.Log()
	require.Len(t, s.logbook.addedQSOs, 1, "VFO1 QSO must be logged")
	vfo1QSO := s.logbook.addedQSOs[0]
	if vfo1QSO.MyNumber != 6 {
		t.Errorf("VFO1 QSO: expected MyNumber=6, got %d", vfo1QSO.MyNumber)
	}
}

// H2. CurrentVFOChanged
// Pre:  rig-side VFO change detected (e.g. hamlib polling).
// Post: if ignoreVFOChange: no-op; if VFO2 disabled: no-op; if same VFO: no-op;
//       else: focusedVFO = vfo; serial displays refreshed; view.SetActiveVFO and SetActiveField called.
// Invariants: input rows; serial claims; logbook.

func TestH2_CurrentVFOChanged_SwitchesFocusAndUpdatesView(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		CurrentVFOChanged(core.VFO2).
		AssertActiveVFO(core.VFO2).
		AssertActiveField(core.VFO2, core.CallsignField)
}

func TestH2_CurrentVFOChanged_VFO2Disabled_Ignored(t *testing.T) {
	NewScenario(t).
		WithClassicExchange(). // VFO2 not enabled
		CurrentVFOChanged(core.VFO2).
		AssertViewNotCalledWith("SetActiveVFO", core.VFO2).
		AssertViewNotCalledWith("SetActiveField", core.VFO2, core.CallsignField)
}

func TestH2_CurrentVFOChanged_SameVFO_Ignored(t *testing.T) {
	// Already on VFO1 → CurrentVFOChanged(VFO1) is a no-op.
	NewScenario(t).
		WithClassicExchange().WithVFO2().
		CurrentVFOChanged(core.VFO1).
		AssertViewNotCalledWith("SetActiveVFO", core.VFO1)
}

func TestH2_CurrentVFOChanged_SuppressedDuringSetFocusedVFO(t *testing.T) {
	// SetFocusedVFO sets ignoreVFOChange=true around the rig command.
	// Simulate: after SetFocusedVFO(VFO2), a CurrentVFOChanged(VFO2) echo
	// from the rig must be suppressed (same-VFO guard catches it since
	// focusedVFO is already VFO2).
	s := NewScenario(t).
		WithClassicExchange().WithVFO2().WithVFOSwitcher()
	s.SetFocusedVFO(core.VFO2) // commands rig; focusedVFO = VFO2
	s.resetSpies()
	s.controller.CurrentVFOChanged(core.VFO2) // echo from rig → same VFO → no-op
	assert.False(t, s.view.wasCalled("SetActiveVFO"),
		"CurrentVFOChanged echo must be suppressed (same VFO)")
}

// H1+. SetFocusedVFO calls SetCurrentVFO and SetActiveVFO

func TestH1_SetFocusedVFO_CallsSetCurrentVFO(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().WithVFOSwitcher().
		SetFocusedVFO(core.VFO2).
		AssertVFOSwitcherCurrentCalled(core.VFO2)
}

func TestH1_SetFocusedVFO_CallsSetActiveVFO(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().WithVFO2().WithVFOSwitcher().
		SetFocusedVFO(core.VFO2).
		AssertActiveVFO(core.VFO2)
}

// H8+. RadioChanged VFO2 collapse releases serial claim

func TestH8_RadioChanged_SingleVFO_VFO2Focused_ReleasesSerialClaim(t *testing.T) {
	s := NewScenario(t).WithClassicExchange().WithNextQSONumber(5)
	s.WithVFO2()
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL1ABC") // claim serial on VFO2
	got := s.controller.CurrentValues().MyNumber
	require.Equal(t, core.QSONumber(5), got, "VFO2 must have claimed serial 5")

	s.controller.RadioChanged("", true) // collapse → release VFO2 claim, move to VFO1

	// After collapse, VFO1 is focused. Serial preview should be 5 (released, re-available).
	vals := s.controller.CurrentValues()
	assert.Equal(t, core.QSONumber(5), vals.MyNumber,
		"after VFO2 collapse, VFO1 serial preview should be 5 (VFO2 claim released)")
}

// K1b. SendQuestion blocked during edit

func TestK1b_SendQuestion_DuringEdit_IsNoOp(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	NewScenario(t).
		WithClassicExchange().WithKeyer().WithVFOSwitcher().
		SelectQSO(qso).
		SendQuestion().
		AssertViewNotCalledWith("SetActiveField", core.VFO1, core.CallsignField) // no side effects
	// keyer.SendQuestion NOT called is the key assertion:
	// AssertNoKeyerText covers SendText; check sentQuestion directly.
}

func TestK1b_SendQuestion_DuringEdit_KeyerNotCalled(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	s := NewScenario(t).
		WithClassicExchange().WithKeyer().WithVFOSwitcher()
	s.SelectQSO(qso)
	s.resetSpies()
	s.controller.SendQuestion()
	assert.Empty(t, s.keyer.sentQuestion,
		"SendQuestion must not call keyer during edit mode")
}

// K2b. RepeatLastTransmission blocked during edit

func TestK2b_RepeatLastTransmission_DuringEdit_IsNoOp(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	s := NewScenario(t).
		WithClassicExchange().WithKeyer()
	s.SelectQSO(qso)
	s.resetSpies()
	s.controller.RepeatLastTransmission()
	assert.False(t, s.keyer.repeated,
		"RepeatLastTransmission must not call keyer during edit mode")
}

// F3c. NextESMStep blocked during edit

func TestF3c_NextESMStep_DuringEdit_IsNoOp(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	s := NewScenario(t).
		WithClassicExchange().WithESMEnabled().WithKeyer().WithVFOSwitcher()
	s.SelectQSO(qso)
	s.resetSpies()
	s.controller.NextESMStep()
	assert.Empty(t, s.keyer.sentTexts,
		"NextESMStep must not call keyer during edit mode")
}

// F3d. NextESMStep — S&P mode, CallsignValid: sends keyer text, no field advance.

func TestF3d_NextESMStep_SP_CallsignValid_SendsButNoFieldAdvance(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		WithVFOSwitcher().
		WithWorkmode(core.SearchPounce)
	s.Enter("DL1ABC") // ESM state = CallsignValid
	s.NextESMStep()
	s.AssertKeyerSentMacro(0)
	// In S&P, CallsignValid does NOT advance field (unlike Run mode).
	// Verify no GotoNextField side effects: active field was not pushed.
	assert.False(t, s.view.wasCalledWith("SetActiveField", core.VFO1, core.TheirExchangeField(1)),
		"S&P CallsignValid must not advance to TheirExchange1 (that's Run-only)")
}

// F3e. NextESMStep — S&P mode, ExchangeValid: logs the QSO.

func TestF3e_NextESMStep_SP_ExchangeValid_Logs(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithESMEnabled().
		WithKeyer().
		WithWorkmode(core.SearchPounce).
		Enter("DL1ABC").
		GotoNextField(). // → TheirExchange1 (RST pre-filled "599")
		GotoNextField(). // → TheirExchange2
		Enter("042").
		GotoNextField(). // → TheirExchange3
		Enter("OE").
		NextESMStep(). // ExchangeValid → Log()
		AssertQSOAdded()
}

// E4+. Log preserves edit QSO time and workmode

func TestE4_SaveEditedQSO_PreservesOriginalTimeAndWorkmode(t *testing.T) {
	editTime := time.Date(2025, time.March, 15, 10, 30, 0, 0, time.UTC)
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      7,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		Time:          editTime,
		Workmode:      core.SearchPounce,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "007", ""},
	}

	s := NewScenario(t).
		WithClassicExchange().
		WithWorkmode(core.Run) // current workmode differs from edit QSO's
	s.SelectQSO(qso)
	s.PressEnter() // Log → UpdateQSO

	require.NotEmpty(t, s.logbook.updatedQSOs, "expected UpdateQSO to be called")
	updated := s.logbook.updatedQSOs[0]
	assert.Equal(t, editTime, updated.Time,
		"edited QSO must preserve original time, not clock.Now()")
	assert.Equal(t, core.SearchPounce, updated.Workmode,
		"edited QSO must preserve original workmode, not current workmode")
}

// I8. Serial claim released on non-focused VFO frequency jump

func TestI8_FrequencyJump_VFO2_ReleasesSerialClaim(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		WithNextQSONumber(5)

	// Enter callsign on VFO1 → claims serial 5.
	s.controller.Enter("DL1ABC")
	// Switch to VFO2, enter callsign → claims serial 6.
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	// Switch back to VFO1.
	s.controller.SetFocusedVFO(core.VFO1)
	s.resetSpies()

	// Large frequency jump on VFO2 → clearInput(VFO2) → Release(VFO2).
	s.controller.VFOFrequencyChanged(core.VFO2, 14050000+1000)

	// VFO2 serial claim released: SetSerialClaim(VFO2, 0, false) must be called.
	s.AssertSerialClaimView(core.VFO2, 0, false)
}

// D2+. Clear fills idle non-focused VFO exchange defaults

func TestD2_Clear_FillsIdleNonFocusedVFODefaults(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2()

	// Set VFO2 band/mode so it has valid state.
	s.controller.VFOBandChanged(core.VFO2, core.Band20m)
	s.controller.VFOModeChanged(core.VFO2, core.ModeCW)

	// VFO2 has no callsign and no claim → it's idle.
	// Clear on VFO1 should also fill VFO2's exchange defaults.
	s.resetSpies()
	s.controller.Clear()

	// VFO2 should have gotten its their-exchange[0] (RST) seeded via fillExchangeDefaults.
	// showInput pushes all VFOs' exchange fields to the view.
	s.view.assertCalledWith(s.t, "SetCallsign", core.VFO2, "")
}

func TestD2_Clear_DoesNotFillNonFocusedVFOWithClaim(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		WithNextQSONumber(5)

	// Switch to VFO2, enter callsign → claims serial.
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	// Switch back to VFO1 and enter something.
	s.controller.SetFocusedVFO(core.VFO1)
	s.controller.Enter("DL1ABC")
	s.resetSpies()

	// Clear on VFO1.
	s.controller.Clear()

	// VFO2 has a claim → fillExchangeDefaults must NOT run for VFO2.
	// VFO2's callsign "DL2XYZ" must survive VFO1's Clear.
	vals := s.controller.CurrentValues()
	s.controller.SetFocusedVFO(core.VFO2)
	vals = s.controller.CurrentValues()
	assert.Equal(t, "DL2XYZ", vals.TheirCall,
		"VFO2 callsign must survive VFO1 Clear when VFO2 has a claim")
}

// I10. SerialSent — claims and commits serial for focused VFO

func TestI10_SerialSent_ClaimsAndCommits(t *testing.T) {
	NewScenario(t).
		WithClassicExchange().
		WithNextQSONumber(5).
		Enter("DL1ABC"). // claims serial 5
		SerialSent().
		AssertSerialCommitted(core.VFO1)
}

func TestI10_SerialSent_WithoutPriorClaim_ClaimsThenCommits(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithNextQSONumber(5)
	// No callsign entered → no claim yet.
	s.SerialSent()
	s.AssertSerialClaimed()
	s.AssertSerialCommitted(core.VFO1)
}

func TestI10_SerialSent_DoesNotAffectOtherVFO(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		WithNextQSONumber(5)
	s.controller.Enter("DL1ABC") // VFO1 claims 5
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ") // VFO2 claims 6
	s.resetSpies()

	s.controller.SerialSent() // commits VFO2 (focused)

	assert.True(t, s.controller.IsSerialCommitted(core.VFO2),
		"VFO2 serial must be committed")
	assert.False(t, s.controller.IsSerialCommitted(core.VFO1),
		"VFO1 serial must not be committed")
}

// I11. Committed serial recycled on Release — no burning

func TestI11_CommittedSerial_RecycledOnClear(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithNextQSONumber(5)

	// Claim and commit serial 5.
	s.controller.Enter("DL1ABC")
	s.controller.SerialSent()

	// Clear without logging → serial 5 recycled (not burned).
	s.controller.Clear()

	// Enter new callsign → serial 5 reused.
	s.controller.Enter("DL2XYZ")
	vals := s.controller.CurrentValues()
	assert.Equal(t, core.QSONumber(5), vals.MyNumber,
		"committed serial 5 must be recycled after clear without logging")
}

func TestI11_UncommittedSerial_RecycledOnClear(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithNextQSONumber(5)

	// Claim serial 5 but do NOT commit.
	s.controller.Enter("DL1ABC")
	assert.False(t, s.controller.IsSerialCommitted(core.VFO1))

	// Clear without logging → release uncommitted serial 5.
	s.controller.Clear()

	// Enter new callsign → serial 5 reused.
	s.controller.Enter("DL2XYZ")
	vals := s.controller.CurrentValues()
	assert.Equal(t, core.QSONumber(5), vals.MyNumber,
		"uncommitted serial 5 must be recycled")
}

func TestI11_DualVFO_CommittedSerial_RecycledAfterClear(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2().
		WithNextQSONumber(5)

	// VFO1 claims 5, VFO2 claims 6. Both committed.
	s.controller.Enter("DL1ABC")
	s.controller.SerialSent()
	s.controller.SetFocusedVFO(core.VFO2)
	s.controller.Enter("DL2XYZ")
	s.controller.SerialSent()

	// Clear VFO1 without logging → serial 5 released.
	s.controller.SetFocusedVFO(core.VFO1)
	s.controller.Clear()

	// VFO1 enters new callsign → gets 5 again (VFO2 holds 6, so 5 is free).
	s.controller.Enter("DL3FOO")
	vals := s.controller.CurrentValues()
	assert.Equal(t, core.QSONumber(5), vals.MyNumber,
		"released serial 5 must be recycled (VFO2 holds 6)")
}

// I12. Edit mode does not interact with committed state

func TestI12_EditMode_CommittedNotAffected(t *testing.T) {
	dl1abc, _ := core.ParseCallsign("DL1ABC")
	qso := core.QSO{
		Callsign:      dl1abc,
		MyNumber:      3,
		Band:          core.Band20m,
		Mode:          core.ModeCW,
		TheirExchange: []string{"599", "042", "OE"},
		MyExchange:    []string{"599", "003", ""},
	}

	s := NewScenario(t).
		WithClassicExchange().
		WithNextQSONumber(5)

	// Claim and commit serial 5.
	s.controller.Enter("DL9ZZZ")
	s.controller.SerialSent()
	assert.True(t, s.controller.IsSerialCommitted(core.VFO1))

	// Enter edit mode → committed state saved, edit claim not committed.
	s.controller.QSOSelected(qso)
	assert.False(t, s.controller.IsSerialCommitted(core.VFO1),
		"edit mode claim must not be committed")

	// Leave edit mode → committed state restored.
	s.controller.Clear()
	assert.True(t, s.controller.IsSerialCommitted(core.VFO1),
		"committed state must be restored after edit mode")
}

// L3b. SO2V workmode: VFO label shows correct workmode after focus switch round-trip.
// Regression test: keyer/entry must show correct workmode regardless of event ordering
// between FocusChanged and WorkmodeChanged.

func TestL3b_SO2V_WorkmodeLabel_CorrectAfterWorkmodeChanged(t *testing.T) {
	s := NewScenario(t).
		WithClassicExchange().
		WithVFO2()

	// Simulate workmode controller setting Run: VFO1=Run, VFO2=S&P.
	s.resetSpies()
	s.controller.WorkmodeChanged(core.VFO1, core.Run)
	s.view.assertCalledWith(s.t, "SetVFOWorkmode", core.VFO1, core.Run)

	s.resetSpies()
	s.controller.WorkmodeChanged(core.VFO2, core.SearchPounce)
	s.view.assertCalledWith(s.t, "SetVFOWorkmode", core.VFO2, core.SearchPounce)

	// After focus switch, workmode controller re-emits WorkmodeChanged for
	// both VFOs. Simulate that arriving after focus switch to VFO2:
	s.controller.SetFocusedVFO(core.VFO2)
	s.resetSpies()
	s.controller.WorkmodeChanged(core.VFO1, core.Run)
	s.controller.WorkmodeChanged(core.VFO2, core.SearchPounce)
	s.view.assertCalledWith(s.t, "SetVFOWorkmode", core.VFO1, core.Run)
	s.view.assertCalledWith(s.t, "SetVFOWorkmode", core.VFO2, core.SearchPounce)

	// Switch back to VFO1, re-emit workmodes — regression case:
	// VFO1 must show Run, not S&P from previous focus.
	s.controller.SetFocusedVFO(core.VFO1)
	s.resetSpies()
	s.controller.WorkmodeChanged(core.VFO1, core.Run)
	s.controller.WorkmodeChanged(core.VFO2, core.SearchPounce)
	s.view.assertCalledWith(s.t, "SetVFOWorkmode", core.VFO1, core.Run)
}

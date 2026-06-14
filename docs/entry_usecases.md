# Entry Controller — Use Cases

All use cases implemented by entry `Controller` (`core/entry/entry.go`,
`core/entry/esm.go`). Each: short description, **Pre** (must hold before),
**Post** (holds after), **Invariants** (unchanged across the action).

Conventions:
- "focused VFO" = `c.focusedVFO`; "other VFO" = the non-focused one.
- "row" = `c.input[vfo]` plus its derived view state.
- Unless stated, the other VFO's input, serial claim, ESM state and view are
  invariants.

---

## A. Callsign / field navigation

### A1. Enter callsign
Type characters into the focused VFO's callsign field.
- **Pre:** active field = `CallsignField` on focused VFO.
- **Post:** `input[focused].callsign` = typed text; `CallsignEntered` listeners notified; callinfo input-changed notified; if text parses as callsign AND not editing: sticky serial claim held for focused VFO; if duplicate exists: error on callsign field, else message cleared.
- **Invariants:** other VFO's row; editing flag; selected band/mode/frequency; logbook.

### A2. Leave callsign field
Triggered by Tab/next-field when leaving callsign.
- **Pre:** active field = `CallsignField`; callsign parses; predicted exchange present iff its length matches `theirExchangeFields`.
- **Post:** empty predictable their-exchange slots filled from prediction; duplicate marker = (duplicate exists AND (not editing OR `editQSO.Callsign != current`)).
- **Invariants:** callsign input text; serial claim; logbook.

### A3. Goto next field (Tab)
Advance active field per transition map.
- **Pre:** any active field.
- **Post:** if leaving callsign: A2 ran; active field = next per map (callsign → first their-exchange; last their-exchange and any my-exchange → callsign; band/mode → callsign); view active field set; ESM state recomputed.
- **Invariants:** input values; per-VFO data other than active field; focused VFO.

### A4. Goto next placeholder
- **Pre:** none.
- **Post:** active field = `CallsignField`; view selects `FilterPlaceholder` text on callsign field.
- **Invariants:** all input values; focused VFO.

### A5. Set active field
- **Pre:** none.
- **Post:** `activeField[focused]` = given field; ESM recomputed.
- **Invariants:** view focus marker not directly updated (caller's job); input values.

### A6. Enter text into a field
Generic text entry routed by active field.
- **Pre:** active field set; for exchange fields, `exchange[i]` slot exists.
- **Post:** corresponding input slot updated; per active field: callsign → A1 effects; band → B1 effects; mode → B2 effects; their-exchange → callinfo notified, error on that field cleared; ESM recomputed.
- **Invariants:** other VFO; editing flag.

---

## B. Band / mode / frequency

### B1. Select band (band field)
- **Pre:** input parses as band.
- **Post:** `selectedBand[focused]` = parsed band; VFO rig commanded to that band; callsign re-entered (A1 effects).
- **Invariants:** other VFO band; mode; frequency.

### B2. Select mode (mode field)
- **Pre:** input parses as mode.
- **Post:** `selectedMode[focused]` = parsed mode; VFO rig commanded; if `generateReport`: my/their report regenerated for focused VFO; callsign re-entered.
- **Invariants:** other VFO mode; band; frequency.

### B3. Enter frequency in kHz via callsign field
Special `EnterPressed` path: numeric callsign input.
- **Pre:** active field = `CallsignField`; input parses as integer.
- **Post:** focused VFO rig commanded with `kHz * 1000`; callsign input cleared; A1 effects with empty callsign.
- **Invariants:** band; mode; other VFO.

### B4. Enter band via callsign field
- **Pre:** active field = `CallsignField`; input parses as band.
- **Post:** B1 effects (via `bandEntered` → `vfos[focused].SetBand`); callsign input cleared; A1 effects with empty callsign.
- **Invariants:** other VFO.

### B5. Jump to bandmap call via callsign field
- **Pre:** active field = `CallsignField`; input starts with `@` and remainder parses as callsign.
- **Post:** bandmap `SelectByCallsign` invoked; callsign input cleared; A1 effects with empty callsign.
- **Invariants:** logbook; rig; other VFO.

### B6. XIT active toggled by UI
- **Pre:** none.
- **Post:** VFO1 rig commanded with new XIT-active flag; view active field reapplied.
- **Invariants:** input values; VFO2 XIT.

---

## C. Logging

### C1. Log valid QSO
- **Pre:** input parses fully (callsign, band, mode, their-exchange complete and valid).
- **Post:** if editing: `UpdateQSO` with original time/workmode; else: `AddQSO` with `clock.Now()` and current workmode; serial previews refreshed on both VFOs; `CallsignLogged` listeners notified; if workmode = S&P: `WorkedSpot` added to bandmap; row cleared (D2 effects on focused VFO).
- **Invariants:** other VFO's input and claim (claim may be re-previewed but not released); contest definition.

### C2. Log fails on invalid callsign
- **Pre:** callsign input does not parse.
- **Post:** error shown on callsign field; active field = callsign.
- **Invariants:** logbook; row not cleared; serial claim unchanged.

### C3. Log fails on invalid band
- **Pre:** callsign parses; band input does not parse.
- **Post:** error on band field; active field = band.
- **Invariants:** logbook; row.

### C4. Log fails on invalid mode
- **Pre:** callsign and band parse; mode does not parse.
- **Post:** error on mode field; active field = mode.
- **Invariants:** logbook; row.

### C5. Log fails on missing their-exchange
- **Pre:** non-`EmptyAllowed` their-exchange field is empty.
- **Post:** error on that field; active field = that field.
- **Invariants:** logbook; row.

### C6. Log fails on invalid their-report
- **Pre:** their-report field non-empty but does not parse as RST.
- **Post:** error on their-report field; active field set there.
- **Invariants:** logbook; row.

### C7. Log fails on invalid serial-only their-number
- **Pre:** their-number field configured as pure serial (only `SerialNumberProperty`/`EmptyProperty`); value non-numeric.
- **Post:** error on their-number field; active field set there.
- **Invariants:** logbook; row.

### C8. Log applies prediction and aborts on empty non-serial/non-report exchange
- **Pre:** non-report/non-serial their-exchange slot empty; prediction array length matches.
- **Post:** prediction copied into that slot; error "check their exchange" on that field; active field set there.
- **Invariants:** logbook; other slots in row.

### C9. Log fails on invalid my-report
- **Pre:** my-exchange's report slot does not parse as RST.
- **Post:** error on my-report field.
- **Invariants:** logbook.

### C10. Log fails on invalid my-number (single-property field)
- **Pre:** my-number field has exactly one property and value is non-numeric.
- **Post:** error on my-number field.
- **Invariants:** logbook.

### C11. EnterPressed dispatch
- **Pre:** any field active.
- **Post:** if callsign field holds a kHz/band/`@call` command (B3–B5): that command run; else if ESM enabled AND not editing: `NextESMStep` (F2); else: Log (C1 and friends).
- **Invariants:** depend on selected branch.

---

## D. Clearing

### D1. Clear while editing
Exit edit modal.
- **Pre:** `editing` = true; `editSnapshot` not nil.
- **Post:** edit snapshot restored (input, claim, claim snapshot, active field, error field, callinfo frame, ESM state/message, pre-edit focused VFO via silent set); `editing` = false; editing marker cleared on VFO1; my-call/frequency/active-field views reapplied; message cleared; last QSO re-selected (with `ignoreQSOSelection` true); callinfo notified with empty input.
- **Invariants:** logbook; other VFO's pre-edit input (was untouched while editing).

### D2. Clear focused VFO (not editing)
- **Pre:** `editing` = false.
- **Post:** serial claim released on focused VFO; focused VFO rig refreshed; active field = callsign; default exchange values seeded on focused VFO and any other VFO with no claim and empty callsign; serial previews refreshed on both VFOs; ESM recomputed; view fully resynced (my-call, frequency, active field, duplicate=false, editing=false on VFO1, message cleared); last QSO selected ignoring re-entry; callinfo notified empty.
- **Invariants:** logbook; selected band/mode/frequency; contest definition; other VFO's claimed serial when non-zero or callsign non-empty.

### D3. Auto-clear on rig frequency jump
- **Pre:** `VFOFrequencyChanged` fires; new frequency differs from stored by > `jumpThreshold`; `ignoreFrequencyJump` = false; not (vfo=VFO1 AND editing).
- **Post:** `selectedFrequency[vfo]` updated; view frequency updated; `clearInput(vfo)` runs — releases serial claim for the event's VFO, resets that VFO's input (exchange defaults seeded), refreshes serial displays, clears callinfo/message for that VFO; if event's VFO = focused VFO: view active field reapplied; `ignoreFrequencyJump` reset to false.
- **Invariants:** focused VFO identity (does not change even if event VFO differs); other VFO's input when not the event target.

### D4. Suppress jump-clear once
- **Pre:** `ignoreFrequencyJump` = true (set by G2).
- **Post:** D3 runs without clear; flag reset to false.
- **Invariants:** row preserved.

---

## E. Edit mode

### E1. QSO selected from list
- **Pre:** `ignoreQSOSelection` = false.
- **Post:** snapshot of VFO1 state captured; `editing` = true; `editQSO` = qso; focused VFO silently = VFO1; `claimedSerial[VFO1]` = `qso.MyNumber`; `claimSnapshot[VFO1]` = current `NextQSONumber`; row shows QSO data on VFO1; active field = callsign; editing marker on VFO1 set; callinfo notified with QSO's call/band/mode/exchange.
- **Invariants:** logbook; VFO2 state.

### E2. QSO selection ignored
- **Pre:** `ignoreQSOSelection` = true.
- **Post:** no-op.
- **Invariants:** everything.

### E3. Edit last QSO
- **Pre:** logbook non-empty (for the list to select).
- **Post:** active field on focused VFO = callsign; `qsoList.SelectLastQSO()` called → routes through E1.
- **Invariants:** until selection event fires: editing flag.

### E4. Save edited QSO
- **Pre:** `editing` = true; current input parses (C1 prereqs).
- **Post:** `UpdateQSO` called with `editQSO.Time` and `editQSO.Workmode`; previews refreshed; `CallsignLogged` emitted; (no bandmap add — bandmap path only fires in non-edit S&P); row cleared via D1 (since `editing` still true at Clear time).
- **Invariants:** original time; original workmode; logbook size.

### E5. Suppress rig sync on VFO1 during edit
- **Pre:** `editing` = true; incoming `VFOFrequencyChanged`/`VFOBandChanged`/`VFOModeChanged` for VFO1.
- **Post:** event ignored.
- **Invariants:** VFO1 row.

---

## F. ESM

### F1. Set ESM view
- **Pre:** none (nil replaced by null view).
- **Post:** `esmView` = supplied view; view's `SetESMEnabled` and `SetMessage` called with current state.
- **Invariants:** `esmEnabled`; ESM state/message.

### F2. Toggle ESM enabled
- **Pre:** none.
- **Post:** `esmEnabled` = new value; view notified; active field reapplied on focused VFO; `ESMListener`s notified.
- **Invariants:** input values; ESM message (until next ESM update).

### F3. NextESMStep
- **Pre:** not editing (`canTransmit` true); keyer set.
- **Post:** `SetTXVFO(focused)` called (with `ignoreVFOChange` guard); ESM state/message recomputed and message displayed; `keyer.SendText(esmMessage[focused])`.
  - **Run + CallsignValid:** GotoNextField; if then on their-report: GotoNextField again (skip RST).
  - **S&P + CallsignValid:** keyer text sent, **no** field advance.
  - **ExchangeValid (both modes):** Log (C1 path).
  - **Editing:** no-op (`canTransmit` returns false).
- **Invariants:** workmode; focused VFO (modulo Log's Clear effects).

### F4. ESM recompute on input/field/workmode change
- **Pre:** triggered by SetActiveField, Enter, WorkmodeChanged, Clear.
- **Post:** `esmState[focused]` and `esmMessage[focused]` reflect current input and workmode; view message updated.
- **Invariants:** other VFO's ESM slot.

### F5. EnterPressed routes around ESM while editing
- **Pre:** `editing` = true.
- **Post:** ESM step not taken; Log path used.
- **Invariants:** ESM state/message.

---

## G. Bandmap / callinfo

### G1. Mark current callsign in bandmap
- **Pre:** focused VFO callsign input parses.
- **Post:** `ManualSpot` with call/freq/band/mode/now added to bandmap.
- **Invariants:** input row; logbook.

### G2. Bandmap entry selected
- **Pre:** any state.
- **Post:** D2 effects (Clear, not editing) — or D1 if editing; `ignoreFrequencyJump` = true; rig tuned to entry frequency; active field = callsign; callsign entered (A1) with entry call; view callsign updated.
- **Invariants:** logbook; contest.

### G3. Select best on-frequency match
- **Pre:** best-match's callsign non-empty.
- **Post:** active field = callsign; A1 effects with that callsign; view callsign/active-field updated.
- **Invariants:** band/mode/frequency.

### G4. Select match by index
- **Pre:** match exists at index.
- **Post:** same as G3 with that match's callsign.
- **Invariants:** same as G3.

### G5. Refresh prediction
- **Pre:** none.
- **Post:** callinfo re-notified with current callsign/band/mode and empty exchange array; if prediction length matches: predictable empty their-exchange slots refilled.
- **Invariants:** non-predictable slots; non-empty slots.

### G6. CallinfoFrameChanged
- **Pre:** any state.
- **Post:** `currentCallinfoFrame[vfo]` = new frame.
- **Invariants:** input; view.

---

## H. VFO focus / dual-VFO

### H1. SetFocusedVFO
User-initiated VFO focus change.
- **Pre:** target ≠ VFO2 OR `vfo2Enabled` = true; target ≠ current focused VFO.
- **Post:** `focusedVFO` = target; `ignoreVFOChange` set true around rig command; `vfoSwitcher.SetCurrentVFO(target)`; serial displays refreshed; `view.SetActiveVFO(target)`; view active field reapplied for target.
- **Invariants:** input rows; serial claims; logbook.

### H2. CurrentVFOChanged
Rig-side VFO change detected (e.g. hamlib polling sees operator changed VFO on radio). Implements `core.CurrentVFOListener`.
- **Pre:** `ignoreVFOChange` = false; target ≠ VFO2 OR `vfo2Enabled` = true; target ≠ current focused VFO.
- **Post:** `focusedVFO` = vfo; serial displays refreshed; `view.SetActiveVFO(vfo)`; `view.SetActiveField(vfo, activeField[vfo])`.
- **Guards (no-op):** `ignoreVFOChange` true (prevents loop when `SetFocusedVFO` commands rig and rig echoes); VFO2 when disabled; same VFO.
- **Invariants:** input rows; serial claims; logbook; rig not commanded.

### H2i. setFocusedVFOSilent
Used internally (edit mode, radio single-VFO collapse).
- **Pre:** target ≠ current.
- **Post:** `focusedVFO` = target; rig **not** commanded.
- **Invariants:** input; view active marker.

### H3. ToggleFocusedVFO
- **Pre:** none.
- **Post:** if `vfo2Enabled` = false: focused = VFO1 (no-op if already); else: focused = other VFO via H1.
- **Invariants:** input rows.

### H4. FocusVFO1 / FocusVFO2
Thin wrappers over H1.
- Same Pre/Post/Invariants as H1.

### H5. LogVFO(vfo)
- **Pre:** H1 preconditions for vfo.
- **Post:** focus moved to vfo (H1); Log executed (C1 path).
- **Invariants:** see C1.

### H6. ClearVFO(vfo)
- **Pre:** H1 preconditions for vfo.
- **Post:** focus moved to vfo (H1); Clear executed (D1/D2).
- **Invariants:** see D1/D2.

### H7. RadioChanged → dual VFO
- **Pre:** event arrives with `singleVFO` = false.
- **Post:** `vfo2Enabled` = true; view `SetVFOEnabled(VFO2, true)`.
- **Invariants:** input; focused VFO.

### H8. RadioChanged → single VFO
- **Pre:** event arrives with `singleVFO` = true.
- **Post:** `vfo2Enabled` = false; if focused was VFO2: VFO2 serial claim released, `input[VFO2]` zeroed, focused silently moved to VFO1; view `SetVFOEnabled(VFO2, false)`.
- **Invariants:** VFO1 input unless focus collapse path touched it (it does not).

---

## I. Serial claim

Claims transition through three states: **unclaimed** → **claimed** (reserved) → **committed** (sent over air). Both claimed and committed serials are recyclable on release. The committed flag is used only for UI visualization (bold serial label). Collision avoidance between VFOs (`nextUnclaimed` skipping other VFO's active claim) is the only reuse prevention.

### I1. Claim on callsign keystroke
- **Pre:** not editing; current callsign parses; `claimedSerial[focused]` = 0.
- **Post:** `claimedSerial[focused]` = `nextUnclaimedSerial(focused)`; `claimSnapshot[focused]` = current `NextQSONumber`; `committed[focused]` = false; my-number inputs refreshed on both VFOs.
- **Invariants:** other VFO's claim; logbook.

### I2. Claim sticky
- **Pre:** `claimedSerial[focused]` ≠ 0.
- **Post:** no-op on further keystrokes (no reclaim).
- **Invariants:** claim value; previews.

### I3. Collision avoidance
Embedded in I1's `nextUnclaimedSerial`.
- **Pre:** other VFO's claim equals current `NextQSONumber`.
- **Post:** new claim = candidate + 1.
- **Invariants:** other VFO's claim.

### I4. Release on Clear (not editing)
- **Pre:** D2 entered.
- **Post:** `claimedSerial[focused]` = 0; `claimSnapshot[focused]` = 0; `committed[focused]` = false; both VFOs' my-number previews recomputed. Released serial is recyclable regardless of committed state.
- **Invariants:** other VFO's claim slot.

### I5. Refresh previews after log
- **Pre:** Log path completed AddQSO/UpdateQSO.
- **Post:** `refreshMyNumberInputs` recomputes both VFOs' displayed serials based on the new `NextQSONumber` and any remaining claim.
- **Invariants:** claim slots.

### I6. Edit mode owns the serial slot
- **Pre:** E1 fired.
- **Post:** `claimedSerial[VFO1]` = `editQSO.MyNumber`; `committed[VFO1]` = false (edit claim is temporary, not a real claim); previous committed state saved in snapshot; restored on leaveEditMode.
- **Invariants:** other VFO's claim.

### I8. Release on frequency jump (non-focused VFO)
- **Pre:** D3 fires on a non-focused VFO (e.g. VFO2 frequency jump while VFO1 focused).
- **Post:** `claimedSerial[eventVFO]` = 0 (via `clearInput` → `claims.Release`); serial previews refreshed on both VFOs; `view.SetSerialClaim(eventVFO, 0, false)`.
- **Invariants:** focused VFO's claim; focused VFO identity.

### I9. Release on RadioChanged VFO2 collapse
- **Pre:** H8 fires with VFO2 focused.
- **Post:** `claimedSerial[VFO2]` = 0 (via `releaseSerialClaimFor`); `input[VFO2]` zeroed; focused silently moved to VFO1.
- **Invariants:** VFO1's claim.

### I10. SerialSent — commit serial on transmission
- **Pre:** keyer emits `SerialSent` (pattern contained `MyNumber` or `MyExchange`).
- **Post:** if no claim on focused VFO: claim created first (I1); `committed[focused]` = true.
- **Invariants:** other VFO's claim and committed state; logbook.

### I11. Committed serial recycled on Release
- **Pre:** `committed[vfo]` = true; Release called (Clear, frequency jump, collapse).
- **Post:** serial released and recyclable. No burning — committed state is UI-only. Next `ClaimNext` can reissue the same serial if logbook and other VFO's claim allow it.
- **Invariants:** logbook; other VFO's claim.

### I12. Edit mode does not interact with committed state
- **Pre:** operator selects QSO for editing.
- **Post:** `committed[VFO1]` set false for duration of edit; restored from snapshot on leave. Committed state of pre-edit claim preserved across edit.
- **Invariants:** highWaterMark; other VFO.

---

## J. Incoming rig events

### J1. VFOFrequencyChanged
- **Pre:** event for vfo with frequency f; not (vfo=VFO1 AND editing); f ≠ `selectedFrequency[vfo]`.
- **Post:** `selectedFrequency[vfo]` = f; view frequency updated for vfo; if `|Δ| > jumpThreshold` AND `ignoreFrequencyJump` = false: Clear runs (focused VFO) and active field reapplied; `ignoreFrequencyJump` reset to false.
- **Invariants:** band; mode; other VFO input when not affected.

### J2. VFOBandChanged
- **Pre:** event for vfo; not (vfo=VFO1 AND editing); band ≠ `NoBand`; band ≠ `selectedBand[focused]`.
- **Post:** `selectedBand[focused]` = band; `input[vfo].band` = band string; view band updated for vfo.
- **Invariants:** mode; frequency; other input slots.

### J3. VFOModeChanged
- **Pre:** event for vfo; not (vfo=VFO1 AND editing); mode ≠ `NoMode`; mode ≠ `selectedMode[focused]`.
- **Post:** `selectedMode[focused]` = mode; `input[vfo].mode` = mode string; view mode updated for vfo.
- **Invariants:** band; frequency.

### J4. VFOXITChanged
- **Pre:** event for vfo.
- **Post:** view `SetXIT(vfo, active, offset)`.
- **Invariants:** input; selected band/mode/freq.

### J5. XITActiveChanged
- **Pre:** event arrives.
- **Post:** view `SetXITActive(VFO1, active)`.
- **Invariants:** all input; VFO2 XIT view.

### J6. VFOPTTChanged
- **Pre:** event for vfo.
- **Post:** if vfo = VFO1: `ptt` = active and TX state view refreshed; else: view `SetTXState(vfo, active, false, 0)`.
- **Invariants:** parrot state when vfo ≠ VFO1.

---

## K. TX, keyer, parrot

### K1. SendQuestion
- **Pre:** keyer set; `canTransmit` = true (not editing).
- **Post:** `SetTXVFO(focused)` called (with `ignoreVFOChange` guard); if active field is their-exchange: `keyer.SendQuestion("nr")`; else: `keyer.SendQuestion(callsign)`.
- **Editing:** no-op (`canTransmit` false).
- **Invariants:** input; logbook.

### K2. RepeatLastTransmission
- **Pre:** keyer set; `canTransmit` = true.
- **Post:** `keyer.Repeat()`. VFO switcher is **not** called (stay on last TX VFO).
- **Editing:** no-op (`canTransmit` false).
- **Invariants:** input.

### K3. StopTX
- **Pre:** keyer set.
- **Post:** `keyer.Stop()`.
- **Invariants:** input; editing flag.

### K4. ParrotActive(active)
- **Pre:** none.
- **Post:** `parrotActive` = active; TX state view refreshed; if active: Clear runs (D1/D2).
- **Invariants:** logbook.

### K5. ParrotTimeLeft(d)
- **Pre:** none.
- **Post:** `parrotTimeLeft` = d; TX state view refreshed.
- **Invariants:** input; parrotActive.

---

## L. Settings / contest / workmode

### L1. StationChanged
- **Pre:** none.
- **Post:** `stationCallsign` = new value; view `SetMyCall`.
- **Invariants:** input; contest.

### L2. ContestChanged
- **Pre:** contest definition supplied.
- **Post:** exchange field definitions replaced; per-VFO `myExchange`/`theirExchange` slices reallocated to new length; Clear runs (D2 since editing not changed; if editing, D1 — but contest changes are not expected mid-edit).
- **Invariants:** station callsign; selected band/mode/freq.

### L3. WorkmodeChanged
- **Pre:** none.
- **Post:** `workmode` = new value; ESM message recomputed.
- **Invariants:** input; contest.

### L4. LogbookLoaded
- **Pre:** logbook loaded.
- **Post:** every VFO's `selectedBand`/`selectedMode` = logbook's last band/mode; Clear runs; input pushed to view.
- **Invariants:** station callsign; contest; focused VFO id.

---

## M. Wiring / refresh / listeners

### M1. SetView
- **Pre:** view ≠ nil; `c.view` is still the initial nullView.
- **Post:** `c.view` = view; `SetVFOEnabled(VFO2, vfo2Enabled)`; Clear; UTC refreshed.
- **Invariants:** input; contest.

### M2. SetKeyer / SetCallinfo
- **Pre:** none.
- **Post:** corresponding field replaced with supplied instance.
- **Invariants:** everything else.

### M3. SetVFO(id, vfo)
- **Pre:** none.
- **Post:** `vfos[id]` = supplied (or null if nil); `vfo.Notify(c)` called (panics if nil supplied because the call happens before the nil-check guard — known caller responsibility).
- **Invariants:** other VFO slot.

### M4. SetVFOSwitcher
- **Pre:** none.
- **Post:** `vfoSwitcher` = supplied (null if nil).
- **Invariants:** focused VFO; rig.

### M5. StartAutoRefresh
- **Pre:** none.
- **Post:** UTC ticker started; periodic `SetUTC` updates begin.
- **Invariants:** input; view fields other than UTC.

### M6. RefreshView
- **Pre:** none.
- **Post:** current input pushed to view (`showInput`).
- **Invariants:** model state.

### M7. Activate
- **Pre:** none.
- **Post:** view active field reapplied for focused VFO.
- **Invariants:** model state.

### M8. Notify(listener)
- **Pre:** none.
- **Post:** listener appended to `listeners`; will receive matching callbacks (`CallsignEntered`, `CallsignLogged`, `ESMEnabled`).
- **Invariants:** existing listeners.

---

## N. Queries

### N1. CurrentQSOState
- **Pre:** none.
- **Post:** returns `(NoCallsign, QSODataEmpty)` if callsign empty or unparseable; `(call, QSODataInvalid)` if call OK and their-exchange parse fails; `(call, QSODataValid)` if both OK.
- **Invariants:** no side effects on input, view, ESM, claims, logbook.

### N2. CurrentValues
- **Pre:** none.
- **Post:** returns `KeyerValues` snapshot from focused VFO input (my report, my number, my-exchange full, my-exchange minus report/number, their callsign).
- **Invariants:** same as N1.

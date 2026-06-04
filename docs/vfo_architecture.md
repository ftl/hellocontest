# VFO Architecture — Current State (Dual-VFO Work in Progress)

## Context

Reference document describing how `app`, `entry`, `vfo`, `radio`, `callinfo`, and `ui` collaborate inside hellocontest. The single-VFO baseline lives in git history; this version captures the state after completing the initial dual-VFO implementation (steps 1–8 of the plan), so that the remaining work can be planned against an accurate baseline.

Last revised: 2026-06-04.

---

## 1. Layered roles

| Package | Role |
|---------|------|
| `core/app` | Composition root. Constructs `core.VFOCount` (=2) VFOs in a loop; wires `RadioChangedListener`, `VFOSwitcher`, and all `Notify` chains. |
| `core/entry` | Owns per-VFO state slices indexed by `VFOID`. A `focusedVFO` cursor routes all user-driven commands. Both VFOs hold independent pending QSOs simultaneously (aka. SO2V). |
| `core/vfo` | One `VFO` struct per physical VFO. Emits events tagged with `id`; filters inbound events that don't match its `id`. |
| `core/radio` | Single backend, single `activeRadio`. The `vfo.Client` interface and all backends are VFO-ID-aware. Emits `RadioChanged(name, singleVFO)` to notify downstream of backend capability. |
| `core/callinfo` | Threaded with `VFOID`; maintains per-VFO `frames` and emits `CallinfoFrameChanged(vfo, frame)`. |
| `ui` | `entryView` has a full second row of widgets for VFO2 (callsign, exchange fields, Log/Clear buttons, status labels). Row visibility driven by `SetVFOEnabled`. |

---

## 2. Core types (`core/core.go:1679–1725`)

```go
type VFOID int
const ( VFO1 VFOID = iota; VFO2; VFOCount )

type VFO interface {
    XITControl
    Name() string
    Notify(any)
    Refresh()
    SetFrequency(Frequency)
    SetBand(Band)
    SetMode(Mode)
    SetXIT(bool, Frequency)
}

type RadioChangedListener interface {
    RadioChanged(name string, singleVFO bool)
}

type CurrentVFOListener interface { CurrentVFOChanged(VFOID) }

// All five VFO event interfaces now carry a leading VFOID:
type VFOFrequencyListener interface { VFOFrequencyChanged(VFOID, Frequency) }
type VFOBandListener      interface { VFOBandChanged(VFOID, Band) }
type VFOModeListener      interface { VFOModeChanged(VFOID, Mode) }
type VFOXITListener       interface { VFOXITChanged(VFOID, bool, Frequency) }
type VFOPTTListener       interface { VFOPTTChanged(VFOID, bool) }
```

Note: `core.VFO` is the per-instance interface (setters take plain values, not `VFOID`). Each `vfo.VFO` struct tags its events with its own `id` on emission. `RadioChangedListener` carries the radio name and a `singleVFO` flag that drives VFO2 visibility.

---

## 3. `core/vfo` — VFO struct and Client interface

**VFO struct** (`vfo.go:28–42`):

```go
type VFO struct {
    XITControl
    id            core.VFOID
    name          string
    bandplan      bandplan.Bandplan
    logbook       Logbook
    client        Client
    offlineClient *offlineClient
    refreshing    bool
    asyncRunner   core.AsyncRunner
    listeners     []any
}
```

Constructor: `NewVFO(id core.VFOID, name string, bandplan, logbook, asyncRunner)`.

**`vfo.Client` interface** (`vfo.go:12–21`) — per-VFO setters take `VFOID`:

```go
type Client interface {
    Notify(any)
    Active() bool
    Refresh()
    SetCurrentVFO(core.VFOID)             // stub in all backends
    SetFrequency(core.VFOID, core.Frequency)
    SetBand(core.VFOID, core.Band)
    SetMode(core.VFOID, core.Mode)
    SetXIT(bool, core.Frequency)          // not yet VFO-scoped
}
```

Inbound filter (`vfo.go:142–176`): each `VFO*Changed` callback starts with `if vfo != v.id { return }`.

Emit semantics (`vfo.go:178–216`): each `emit*Changed` method wraps listeners via `asyncRunner` before calling the listener method, ensuring callers land on the UI thread even when the backend emits from a goroutine.

---

## 4. `core/radio` — backend bridge

`Controller` struct (`radio.go:31–46`) holds one `activeRadio radio` plus `listeners []any`. The internal `radio` interface (`radio.go:48–60`) now requires `SingleVFO() bool` and `SetCurrentVFO(core.VFOID)`.

**`emitRadioChanged`** (`radio.go:137–146`): called from `SelectRadio` (success and error paths) and `Stop`. Fires `RadioChangedListener.RadioChanged(name, singleVFO)` to all `c.listeners`, then updates the view via `c.view.SetRadioSelected(name)`.

**`SelectRadio`** (`radio.go:168–219`): after connecting, forwards `c.listeners` wholesale to `activeRadio.Notify` (`radio.go:204–206`), then registers a `ConnectionChangedFunc` for radio status tracking, giving VFO objects and other listeners direct access to hamlib/TCI events.

**Entry subscribed via `c.Radio.Notify(c.Entry)`** (app.go:211). Because the listener list is forwarded to the hamlib client, Entry's VFO event handlers can be invoked from the hamlib polling goroutine; all such handlers are wrapped in `c.asyncRunner` (see §5).

**Backend capability:**

| Backend | `SingleVFO()` | `SetCurrentVFO()` | Dual-VFO polling | Dual-VFO setting |
|---------|--------------|------------------|-----------------|--------------------|
| hamlib | returns `c.singleVFO` (set when `vfo1` option absent — `sanitizeVFOs` falls back to `CurrVFO`) | stub — logs, no-op | `pollDualVFO` / `pollSingleVFO` fallback | `SetFrequency/Band/Mode(vfo, …)` ✓ |
| TCI | reads `single_vfo` config key (default false) | stub — logs, no-op | per-TRX routing | `toTCIVFO()` mapping ✓ |

**`SetXIT`** (`hamlib.go:359+`): still hardcodes `core.VFO1` internally. `// TODO: add the VFOID to all VFO-related Setters`.

---

## 5. `core/entry` — per-VFO state and routing

### 5.1 Controller struct fields (`entry.go:151–203`)

Per-VFO slices (length `core.VFOCount`, index = `VFOID`):

| Field | Type |
|-------|------|
| `vfos` | `[]core.VFO` |
| `input` | `[]input` |
| `selectedFrequency` | `[]core.Frequency` |
| `selectedBand` | `[]core.Band` |
| `selectedMode` | `[]core.Mode` |
| `activeField` | `[]core.EntryField` |
| `errorField` | `[]core.EntryField` |
| `currentCallinfoFrame` | `[]core.CallinfoFrame` |
| `esmState` | `[]core.ESMState` |
| `esmMessage` | `[]string` |

Serial claims are managed by the `claims SerialClaims` struct (see §5.3).

Shared/single fields: `focusedVFO core.VFOID`, `vfo2Enabled bool`, `editing bool`, `editQSO core.QSO`, `editSnapshot *editSnapshot`, `ptt bool`, `parrotActive bool`, `parrotTimeLeft`, `esmEnabled bool`.

All per-VFO slices initialised in `NewController` (`entry.go:95–128`). `activeField` elements seeded to `core.CallsignField` so that a VFO focus switch always lands on a valid field.

### 5.2 Focus model (`entry.go:466–519`)

`SetFocusedVFO(vfo VFOID)` is the single funnel:
1. Guards: no-op if `vfo == VFO2 && !c.vfo2Enabled`; no-op if already focused.
2. Updates `c.focusedVFO`.
3. Calls `c.vfoSwitcher.SetCurrentVFO(vfo)` (commands rig, currently stub).
4. Calls `c.view.SetActiveField(c.focusedVFO, c.activeField[c.focusedVFO])` — moves UI cursor.

`setFocusedVFOSilent` skips steps 3–4; used by edit-mode enter/leave to avoid rig command and unnecessary view churn.

Public focus actions (wired to F8/F9/F10 in `ui/actions.go` and remote server):

| Method | Hotkey | Behaviour |
|--------|--------|-----------|
| `ToggleFocusedVFO()` | F8 | Flips VFO1↔VFO2; no-op if VFO2 disabled |
| `FocusVFO1()` | F9 | `SetFocusedVFO(VFO1)` |
| `FocusVFO2()` | F10 | `SetFocusedVFO(VFO2)` |

Log/Clear helpers:

```go
func (c *Controller) LogVFO(vfo VFOID)   { c.SetFocusedVFO(vfo); c.Log() }
func (c *Controller) ClearVFO(vfo VFOID) { c.SetFocusedVFO(vfo); c.Clear() }
```

`LogVFO`/`ClearVFO` have keybinding slots in `cfg.go` (`entry.log_vfo1`, `entry.log_vfo2`, `entry.clear_vfo1`, `entry.clear_vfo2`) but no default hotkeys assigned.

### 5.3 Serial claim model (`core/entry/serial_claims.go`)

Serial claim logic is encapsulated in the `SerialClaims` struct:

```go
type SerialClaims struct {
    claimed  []core.QSONumber // per-VFO reserved serial; 0 = unclaimed
    snapshot []core.QSONumber // logbook.NextQSONumber() at claim time
}
```

A serial number can be in one of three states:

- **Claimed**: assigned to a VFO's pending (unlogged) QSO the moment a valid callsign is first parsed. Stored in `claimed[vfo]`. Sticky — re-parsing the same field with the same VFO does not re-claim.
- **Taken**: logged into the logbook. Permanently unavailable.
- **Unclaimed / preview**: next available serial, skipping the other VFO's claimed serial.

Key methods on `SerialClaims`:

```go
nextUnclaimed(forVFO, base)        // walks from base, skips other VFO's claim
DisplayedSerial(vfo, base)         // claimed || nextUnclaimed preview
AllDisplayed(base)                 // displayed serial for every VFO
ClaimNext(vfo, base)               // reserve next unclaimed serial; no-op if already claimed
Release(vfo)                       // zero the claim and snapshot slots
```

Entry controller wrappers:

```go
claimSerialFor(vfo)                // delegates to claims.ClaimNext, then refreshMyNumberInputs
releaseSerialClaimFor(vfo)         // delegates to claims.Release, then refreshMyNumberInputs
refreshMyNumberInputs()            // reads NextQSONumber once, writes both VFOs' input.myNumber
refreshMyNumberInput(vfo)          // single-VFO variant
writeMyNumberInput(vfo, serial)    // writes serial into input[vfo].myNumber and exchange slot
```

### 5.4 Edit-mode modal (`entry.go:578–619`)

`enterEditMode(qso core.QSO)`:
1. Snapshots VFO1 state into `editSnapshot` (focusedVFO, full `input[VFO1]`, `claims.claimed[VFO1]`, `claims.snapshot[VFO1]`, activeField, errorField, callinfoFrame, esmState[], esmMessage[]).
2. `setFocusedVFOSilent(VFO1)` — no rig command, no view cursor move.
3. Sets `editing = true`, `editQSO = qso`.
4. Claims `qso.MyNumber` for VFO1 for the duration (writes directly to `claims.claimed[VFO1]` and `claims.snapshot[VFO1]`).

`leaveEditMode()` (called by `Clear()` when `editing == true`):
1. Restores all snapshotted fields for VFO1.
2. `setFocusedVFOSilent(snap.focusedVFO)` — returns focus to where it was.
3. Clears `editing`, `editQSO`, `editSnapshot`.
4. **TODO**: call `c.view.SetVFOEnabled(core.VFO2, true)` after edit exit (VFO2 disable during edit not yet wired, see §11).

`canTransmit() bool` (`entry.go:534`): returns `!c.editing`. All keyer entry points (`SendQuestion`, `RepeatLastTransmission`, `NextESMStep`, ESM auto-send in `EnterPressed`) guard on this.

### 5.5 VFO event handlers (`entry.go:684–805`)

All five handlers are fully wrapped in `c.asyncRunner` so that the entire body (state mutation + view call) runs on the UI thread, regardless of whether the caller is a hamlib goroutine or a vfo-marshalled call:

```go
func (c *Controller) VFOFrequencyChanged(vfo VFOID, frequency Frequency) {
    c.asyncRunner(func() {
        // ... guard edits, compute jump, update selectedFrequency[vfo], call view
    })
}
```

Same pattern for `VFOBandChanged`, `VFOModeChanged`, `VFOXITChanged`, `VFOPTTChanged`.

`VFOFrequencyChanged`: detects jumps (`> jumpThreshold`); on jump triggers `Clear()` and resets `ignoreFrequencyJump`.

`VFOBandChanged` / `VFOModeChanged`: **note** — these compare and write `selectedBand/selectedMode` using `c.focusedVFO` (not the event's `vfo` parameter), but write `input[vfo].band/mode` using the event parameter. This means the comparison guard and state update target different VFO slots when the event comes from a non-focused VFO.

`VFOPTTChanged`: VFO1 PTT drives `c.ptt` + `updateTXState()`; other VFOs update their TX indicator directly.

### 5.6 ESM (`entry/esm.go`)

Fully per-VFO: `esmState[focusedVFO]`, `esmMessage[focusedVFO]`. `NextESMStep` guards via `canTransmit()`. `updateESM` reads and writes the focused VFO's slots.

### 5.7 `Log()` (`entry.go:934–1027`)

Routes from `c.focusedVFO` for callsign, frequency, band, mode, exchange input. No longer locked to VFO1. On success, calls `refreshMyNumberInputs()` so the other VFO's serial preview updates immediately.

### 5.8 `Clear()` (`entry.go:1140–1194`)

Two paths:
- **Edit exit** (`c.editing == true`): calls `leaveEditMode()`, then redraws VFO1 state from snapshot.
- **Normal clear**: calls `c.claims.Release(c.focusedVFO)`, refreshes VFO, resets `input[focusedVFO]`, calls `fillExchangeDefaults(focusedVFO, lastExchange)`. Also fills idle non-focused VFOs (those with no callsign and no claim — checked via `c.claims.claimed[v] != 0`) so default reports stay current after mode/band changes. Finally calls `refreshMyNumberInputs()` to sync serial displays for both VFOs.

---

## 6. `core/callinfo` — VFOID threading

`Callinfo.InputChanged(vfo core.VFOID, call string, band, mode, exchange)` updates `frames[vfo]` and emits `CallinfoFrameChanged(vfo, frames[vfo])`.

`Entry.CallinfoFrameChanged(vfo, frame)` stores into `currentCallinfoFrame[vfo]`.

`EntryOnFrequency` (bandmap spot overlay) is VFO1-only by design — bandmap is tied to VFO1.

---

## 7. App wiring (`core/app/app.go:188–280`)

Construction and notification order:

```
Entry := entry.NewController(...)
Radio := radio.NewController(...)

Radio.Notify(ServiceStatus)
Radio.Notify(Entry)         // Entry receives RadioChanged + is forwarded to hamlib listeners
Entry.SetVFOSwitcher(Radio) // Entry can command rig VFO switch

for vfoID in 0..VFOCount:
    v := vfo.NewVFO(vfoID, ...)
    v.SetClient(Radio)
    Entry.SetVFO(vfoID, v)
    Logbook.Notify(v)

Bandmap.SetVFO(VFOs[VFO1])       // VFO1-only
Workmode.Notify(VFOs[VFO1])      // VFO1-only

Radio.SelectRadio(session.Radio1())
Radio.SelectKeyer(session.Keyer1())

Keyer := keyer.New(...)
Entry.SetKeyer(Keyer)

Callinfo.Notify(Entry)  // CallinfoFrameChanged flows to entry
Entry.SetCallinfo(Callinfo)

VFOs[VFO1].Notify(QTCController) // VFO1-only
```

**`DoAction` VFO cases** (`app.go:962–975`):

```go
case ActionEntryToggleFocusedVFO: c.Entry.ToggleFocusedVFO()
case ActionEntryFocusVFO1:        c.Entry.FocusVFO1()
case ActionEntryFocusVFO2:        c.Entry.FocusVFO2()
case ActionEntryLogVFO1:          c.Entry.LogVFO(core.VFO1)
case ActionEntryLogVFO2:          c.Entry.LogVFO(core.VFO2)
case ActionEntryClearVFO1:        c.Entry.ClearVFO(core.VFO1)
case ActionEntryClearVFO2:        c.Entry.ClearVFO(core.VFO2)
```

All seven actions also reachable via remote server (`POST /do?action=<id>`).

---

## 8. UI surface

### `ui/entryView.go`

VFO2 widgets (`entryView` struct, lines 50–62):
- `vfo2Label`, `vfo2FrequencyLabel`, `vfo2BandModeContainer`
- `vfo2XITIndicator`, `vfo2TXIndicator`
- `vfo2TheirLabel`, `vfo2Callsign`, `vfo2TheirExchangeFields`
- `vfo2LogButton`, `vfo2ClearButton`

`vfo2Enabled bool` stored on `entryView` (line 78) for visibility bookkeeping.

**`SetVFOEnabled(vfo, enabled)`** (`entryView.go:646–681`): VFO1 is always enabled (no-op). For VFO2, shows/hides the entire widget cluster. Records `v.vfo2Enabled = enabled`.

**`setExchangeFields`** (`entryView.go`): when building new VFO2 exchange fields, checks `v.vfo2Enabled` and hides/disables newly created fields if the flag is false. This prevents a startup race where `SetVFOEnabled(VFO2, false)` fires before the exchange fields exist.

All per-row view methods take a `VFOID` and dispatch to the correct widget set: `SetCallsign`, `SetTheirExchange`, `SetActiveField`, `SetDuplicateMarker`, `SetEditingMarker`, `ShowMessage`, `ClearMessage`, `SetFrequency`, `SetBand`, `SetMode`, `SetXIT`, `SetXITActive`, `SetTXState`.

Hotkey actions (`ui/actions.go`):

```
F8  → toggleVFOAction  → Entry.ToggleFocusedVFO()
F9  → focusVFO1Action  → Entry.FocusVFO1()
F10 → focusVFO2Action  → Entry.FocusVFO2()
```

Hotkey defaults also recorded in `cfg.Default.Keybindings` (`cfg/cfg.go`).

### `ui/centralArea.go`

VFO2 widgets occupy rows 8–9 of the entry grid layout. Row 8 holds VFO2 frequency/band/mode/XIT/TX indicators; row 9 holds their-data (callsign, exchange, Log, Clear). `removeWidgetsFromLayout` handles both rows.

---

## 9. Event flow diagram

```
hamlib goroutine                vfo.VFO (UI thread via asyncRunner)
      │                                │
      │ emitFrequencyChanged ──────────► VFOFrequencyChanged(id, f)
      │ emitBandChanged      ──────────► VFOBandChanged(id, b)
      │ emitModeChanged      ──────────► VFOModeChanged(id, m)
      │ emitXITChanged       ──────────► VFOXITChanged(id, …)
      │ emitPTTChanged       ──────────► VFOPTTChanged(id, p)
      │                                │
      │ (also forwarded directly to Entry via Radio.Notify → wrapped in asyncRunner inside handler)
      │                                │
      ▼                                ▼
                        entry.Controller
                        ┌─────────────────────────────────┐
                        │ focusedVFO cursor                │
                        │ per-VFO: input[], band[], mode[] │
                        │          activeField[], esmState[]│
                        │          claims (SerialClaims)    │
                        └──────────────┬──────────────────┘
                                       │ view calls (UI thread)
                                       ▼
                               entryView (Qt widgets)
                               VFO1 row  │  VFO2 row
                               callsign  │  callsign
                               exchange  │  exchange
                               Log/Clear │  Log/Clear

radio.Controller (main thread)
  emitRadioChanged(name, singleVFO)
        │
        ├──► Entry.RadioChanged  → vfo2Enabled, view.SetVFOEnabled
        └──► view.SetRadioSelected (inside emitRadioChanged)
```

---

## 10. Invariants

- **Single event source per VFO**: `vfo.VFO.emit*` is the canonical egress. Backends never bypass it. (Entry also receives raw hamlib events via `Radio.Notify` forwarding, but those handlers are fully wrapped in `asyncRunner`.)
- **asyncRunner contract**: all UI-touching code in Entry event handlers runs inside `c.asyncRunner`. VFO objects marshal outward via their own `asyncRunner`. No Qt call made from a non-UI goroutine.
- **Construction precedes wiring**: `Radio` constructed before VFOs; VFOs constructed before `SelectRadio`; VFOs registered on Entry before events can fire.
- **focusedVFO as single routing cursor**: every user input path reads `c.focusedVFO`; `SetFocusedVFO` is the sole mutator from outside the controller.
- **vfo2Enabled gates VFO2**: `SetFocusedVFO(VFO2)`, `FocusVFO2`, `LogVFO(VFO2)`, `ClearVFO(VFO2)` all no-op when `!c.vfo2Enabled`. View mirrors state via `SetVFOEnabled`.

---

## 11. Outstanding work

| Location | Item |
|----------|------|
| `core/hamlib/hamlib.go:318`, `core/tci/tci.go:78` | `SetCurrentVFO(VFOID)` is a stub (log + no-op). Real implementation must command the rig to switch TX VFO. |
| `core/hamlib/hamlib.go:359` `SetXIT` | Hardcodes `core.VFO1`. Needs `VFOID` parameter like the other setters. Comment: `// TODO: add the VFOID to all VFO-related Setters`. |
| `core/entry/entry.go:789` `XITActiveChanged` | `// TODO: add VFO parameter to XITActiveChanged` — handler and `vfo.XITControl` interface lack `VFOID`. |
| `core/entry/entry.go:597` `enterEditMode` | `// TODO step 6: c.view.SetVFOEnabled(core.VFO2, false)` — edit mode does not yet hide VFO2 on enter. |
| `core/entry/entry.go:618` `leaveEditMode` | `// TODO step 6: c.view.SetVFOEnabled(core.VFO2, true)` — complement of above; VFO2 stays visible and interactive during editing. |
| `core/entry/entry.go:714,768` `VFOBandChanged`/`VFOModeChanged` | These handlers compare and update `selectedBand[focusedVFO]` / `selectedMode[focusedVFO]` but write `input[vfo]` using the event's VFO param. When the event arrives for a non-focused VFO the comparison guard and state update target different slots. Needs review. |
| `core/app/app.go:223–224` | `Bandmap.SetVFO`, `Workmode.Notify` intentionally VFO1-only. Decide if bandmap routing should follow `focusedVFO` or stay VFO1-only. |
| `core/app/app.go:251` | `VFOs[VFO1].Notify(QTCController)` — QTC stays VFO1-only by design; confirm or expand. |
| `ui/entryView.go` | Tab order cycles within each VFO's row independently. Confirm correct for SO2R workflow. |
| `ui/actions.go` | `LogVFO`/`ClearVFO` actions not wired as UI actions (keybinding slots exist in `cfg.go` with empty defaults but no `makeTriggerAction` in `actions.go`). |
| Remote server | `LogVFO`/`ClearVFO` currently pass `VFOID` as a Go call only; verify remote HTTP surface encodes VFO identity if needed beyond action IDs. |

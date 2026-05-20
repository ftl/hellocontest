# VFO Architecture — Current State (Dual-VFO Work in Progress)

## Context

Reference document describing how `app`, `entry`, `vfo`, `radio`, `callinfo`, and `ui` collaborate inside hellocontest. The single-VFO baseline lives in git history; this version captures the state after completing the initial dual-VFO implementation (steps 1–8 of the plan), so that the remaining work can be planned against an accurate baseline.

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

`RadioChangedListener` is new. It carries the radio name and a `singleVFO` flag that drives VFO2 visibility.

---

## 3. `core/vfo` — VFO struct and Client interface

**VFO struct** (`vfo.go:27–59`):

```go
type VFO struct {
    XITControl
    id          core.VFOID   // identifies this instance
    name        string
    bandplan    bandplan.Bandplan
    logbook     Logbook
    client      Client
    offlineClient *offlineClient
    asyncRunner core.AsyncRunner
    listeners   []any
}
```

Constructor: `NewVFO(id core.VFOID, name string, bandplan, logbook, asyncRunner)`.

**`vfo.Client` interface** (`vfo.go:12–20`) — per-VFO setters take `VFOID`:

```go
type Client interface {
    Notify(any)
    Active() bool
    Refresh()
    SetCurrentVFO(core.VFOID)             // NEW; stub in all backends
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

**`emitRadioChanged`** (`radio.go:137–146`): called from `SelectRadio` (success and error paths) and `Stop`. Fires `RadioChangedListener.RadioChanged(name, singleVFO)` to all `c.listeners`. `c.listeners` is also forwarded wholesale to `activeRadio.Notify` in `SelectRadio` (`radio.go:204–205`), giving VFO objects and other listeners direct access to hamlib/TCI events.

**Entry subscribed via `c.Radio.Notify(c.Entry)`** (app.go:211). Because the listener list is forwarded to the hamlib client, Entry's VFO event handlers can be invoked from the hamlib polling goroutine; all such handlers are wrapped in `c.asyncRunner` (see §5).

**Backend capability:**

| Backend | `SingleVFO()` | `SetCurrentVFO()` | Dual-VFO polling | Dual-VFO setting |
|---------|--------------|------------------|-----------------|-----------------|
| hamlib | returns `c.singleVFO` (set when `vfo2` option absent) | stub — logs, no-op | `pollDualVFO` / `pollSingleVFO` fallback | `SetFrequency/Band/Mode(vfo, …)` ✓ |
| TCI | reads `single_vfo` config key (default false) | stub — logs, no-op | per-TRX routing | `toTCIVFO()` mapping ✓ |

**`SetXIT`** (`hamlib.go:359+`): still hardcodes `core.VFO1` internally. `// TODO: add VFOID to SetXIT`.

---

## 5. `core/entry` — per-VFO state and routing

### 5.1 Controller struct fields (`entry.go:152–205`)

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
| `claimedSerial` | `[]core.QSONumber` |
| `claimSnapshot` | `[]core.QSONumber` |
| `esmState` | `[]core.ESMState` |
| `esmMessage` | `[]string` |

Shared/single fields: `focusedVFO core.VFOID`, `vfo2Enabled bool`, `editing bool`, `editQSO core.QSO`, `editSnapshot *editSnapshot`, `ptt bool`, `parrotActive bool`, `parrotTimeLeft`, `esmEnabled bool`.

All per-VFO slices initialised in `NewController` (`entry.go:95–128`). `activeField` elements seeded to `core.CallsignField` so that a VFO focus switch always lands on a valid field.

### 5.2 Focus model (`entry.go:466–520`)

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

### 5.3 Serial claim model (`entry.go:540–620`)

A serial number can be in one of three states:

- **Claimed**: assigned to a VFO's pending (unlogged) QSO the moment a valid callsign is first parsed. Stored in `claimedSerial[vfo]`. Sticky — re-parsing the same field with the same VFO does not re-claim.
- **Taken**: logged into the logbook. Permanently unavailable.
- **Unclaimed / preview**: next available serial, skipping the other VFO's claimed serial.

Key methods:

```go
claimSerialFor(vfo)          // reserve next unclaimed serial; no-op if already claimed
releaseSerialClaimFor(vfo)   // zero the claim slot
nextUnclaimedSerial(forVFO)  // walks from NextQSONumber(), skips other VFO's claim
displayedSerialFor(vfo)      // claimed || nextUnclaimedSerial preview
refreshMyNumberInputs()      // reads NextQSONumber once, writes both VFOs' input.myNumber
```

`claimSnapshot[vfo]` records `NextQSONumber()` at claim time. On release via `Clear()`, if the logbook number has advanced (i.e. a QSO was logged on the other VFO in between), the serial is permanently skipped.

### 5.4 Edit-mode modal (`entry.go:624–663`)

`enterEditMode(qso core.QSO)`:
1. Snapshots VFO1 state into `editSnapshot` (focusedVFO, full `input[VFO1]`, claimedSerial, claimSnapshot, activeField, errorField, callinfoFrame, esmState[], esmMessage[]).
2. `setFocusedVFOSilent(VFO1)` — no rig command, no view cursor move.
3. Sets `editing = true`, `editQSO = qso`.
4. Claims `qso.MyNumber` for VFO1 for the duration.

`leaveEditMode()` (called by `Clear()` when `editing == true`):
1. Restores all snapshotted fields for VFO1.
2. `setFocusedVFOSilent(snap.focusedVFO)` — returns focus to where it was.
3. Clears `editing`, `editQSO`, `editSnapshot`.
4. **TODO**: call `c.view.SetVFOEnabled(core.VFO2, true)` after edit exit (VFO2 disable during edit not yet wired, see §8).

`canTransmit() bool` (`entry.go:536`): returns `!c.editing`. All keyer entry points (`SendQuestion`, `RepeatLastTransmission`, `NextESMStep`, ESM auto-send in `EnterPressed`) guard on this.

### 5.5 VFO event handlers (`entry.go:728–868`)

All five handlers are fully wrapped in `c.asyncRunner` so that the entire body (state mutation + view call) runs on the UI thread, regardless of whether the caller is a hamlib goroutine or a vfo-marshalled call:

```go
func (c *Controller) VFOFrequencyChanged(vfo VFOID, frequency Frequency) {
    c.asyncRunner(func() {
        // ... guard edits, compute jump, update selectedFrequency, call view
    })
}
```

Same pattern for `VFOBandChanged`, `VFOModeChanged`, `VFOXITChanged`, `VFOPTTChanged`.

`VFOFrequencyChanged`: detects jumps (`> jumpThreshold`); on jump triggers `Clear()` and resets `ignoreFrequencyJump`.

`VFOPTTChanged`: VFO1 PTT drives `c.ptt` + `updateTXState()`; other VFOs update their TX indicator directly.

### 5.6 ESM (`entry/esm.go`)

Fully per-VFO: `esmState[focusedVFO]`, `esmMessage[focusedVFO]`. `NextESMStep` guards via `canTransmit()`. `updateESM` reads and writes the focused VFO's slots.

### 5.7 `Log()` (`entry.go:980–1071`)

Now routes from `c.focusedVFO` for callsign, frequency, band, mode, exchange input. No longer locked to VFO1. On success, calls `refreshMyNumberInputs()` so the other VFO's serial preview updates immediately.

### 5.8 `Clear()` (`entry.go:1184–1239`)

Two paths:
- **Edit exit** (`c.editing == true`): calls `leaveEditMode()`, then redraws VFO1 state from snapshot.
- **Normal clear**: releases `claimedSerial[focusedVFO]`, resets `input[focusedVFO]`, calls `fillExchangeDefaults(focusedVFO, lastExchange)`. Also fills idle non-focused VFOs (those with no callsign and no claim) so default reports stay current after mode/band changes.

---

## 6. `core/callinfo` — VFOID threading

`Callinfo.InputChanged(vfo core.VFOID, call string, band, mode, exchange)` updates `frames[vfo]` and emits `CallinfoFrameChanged(vfo, frames[vfo])`.

`Entry.CallinfoFrameChanged(vfo, frame)` stores into `currentCallinfoFrame[vfo]`.

`EntryOnFrequency` (bandmap spot overlay) is VFO1-only by design — bandmap is tied to VFO1.

---

## 7. App wiring (`core/app/app.go:188–275`)

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
VFOs[VFO1].Notify(QTCController) // VFO1-only

Radio.SelectRadio(session.Radio1())

Callinfo.Notify(Entry)  // CallinfoFrameChanged flows to entry
Entry.SetCallinfo(Callinfo)
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

`vfo2Enabled bool` stored on `entryView` for visibility bookkeeping (see below).

**`SetVFOEnabled(vfo, enabled)`** (`entryView.go:640–676`): VFO1 is always enabled (no-op). For VFO2, shows/hides the entire widget cluster. Records `v.vfo2Enabled = enabled`.

**`setExchangeFields`** (`entryView.go:485–511`): when building new VFO2 exchange fields, checks `v.vfo2Enabled` and hides/disables newly created fields if the flag is false. This prevents a startup race where `SetVFOEnabled(VFO2, false)` fires before the exchange fields exist.

All per-row view methods take a `VFOID` and dispatch to the correct widget set: `SetCallsign`, `SetTheirExchange`, `SetActiveField`, `SetDuplicateMarker`, `SetEditingMarker`, `ShowMessage`, `ClearMessage`, `SetFrequency`, `SetBand`, `SetMode`, `SetXIT`, `SetXITActive`, `SetTXState`.

Hotkey actions (`ui/actions.go`):

```
F8  → toggleVFOAction  → Entry.ToggleFocusedVFO()
F9  → focusVFO1Action  → Entry.FocusVFO1()
F10 → focusVFO2Action  → Entry.FocusVFO2()
```

Hotkey defaults also recorded in `cfg.Default.Keybindings` (`cfg/cfg.go`).

### `ui/centralArea.go`

VFO2 widgets occupy row 9 of the entry grid layout (below VFO1's their-exchange row). `removeWidgetsFromLayout` handles both rows 8 and 9.

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
                        │          claimedSerial[]          │
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
        └──► view (if implements RadioChangedListener)
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
| `core/hamlib/hamlib.go`, `core/tci/tci.go` | `SetCurrentVFO(VFOID)` is a stub (log + no-op). Real implementation must command the rig to switch TX VFO. |
| `core/hamlib/hamlib.go` `SetXIT` | Hardcodes `core.VFO1`. Needs `VFOID` parameter like the other setters. |
| `core/entry/entry.go:577` `XITActiveChanged` | `// TODO: add VFO parameter to XITActiveChanged` — handler and `vfo.XITControl` interface lack `VFOID`. |
| `core/entry/entry.go` `leaveEditMode` line 662 | `// TODO step 6: c.view.SetVFOEnabled(core.VFO2, true)` — edit mode does not yet call `SetVFOEnabled(VFO2, false)` on enter or `true` on leave. VFO2 stays visible and interactive during editing. |
| `core/entry/entry.go` `enterEditMode` | Complement of above: `c.view.SetVFOEnabled(core.VFO2, false)` missing at edit entry. |
| `core/app/app.go:223–224` | `Bandmap.SetVFO`, `Workmode.Notify` intentionally VFO1-only. Decide if bandmap routing should follow `focusedVFO` or stay VFO1-only. |
| `core/app/app.go:249` | `VFOs[VFO1].Notify(QTCController)` — QTC stays VFO1-only by design; confirm or expand. |
| `ui/entryView.go` | Tab order cycles within each VFO's row independently. Confirm correct for SO2R workflow. |
| Remote server | `LogVFO`/`ClearVFO` currently pass `VFOID` as a Go call only; verify remote HTTP surface encodes VFO identity if needed beyond action IDs. |

# VFO Architecture — Current State (Dual-VFO Work in Progress)

## Context

Reference document describing how `app`, `entry`, `vfo`, and `radio` collaborate inside hellocontest. The single-VFO baseline of this document lives in git history; the present version captures the state on the in-progress `2vfo` branch, which extends the wiring to support a second VFO. Each section calls out what changed against the original single-VFO design so the remaining work can be planned against an accurate baseline.

---

## 1. Layered roles — unchanged in principle, multiplied in count

| Package | Role then | Role now |
|---------|-----------|----------|
| `core/app` | Composition root, wires single VFO | Same role, but constructs **`core.VFOCount` VFOs** in a loop (currently 2: VFO1, VFO2) |
| `core/entry` | Held one VFO via `core.VFO` interface | Holds a **slice** `vfos []core.VFO` indexed by `VFOID`, plus a `focusedVFO core.VFOID` cursor that routes user input |
| `core/vfo` | One `VFO` struct, owned listener list | Each `VFO` carries an `id core.VFOID` and a `name`; events emitted are tagged with that ID. Multiple `VFO` instances coexist over the same `Client` |
| `core/radio` | Single backend, single `vfo.Client` | Still a single backend, but the `vfo.Client` interface is **VFO-ID-aware** and the hamlib/TCI backends know how to talk to two physical VFOs on one rig |

No new top-level packages. No cyclic imports. All multi-VFO contracts live in `core/core.go` alongside the existing listener interfaces.

---

## 2. New core types (`core/core.go:1679–1721`)

```go
type VFOID int

const (
    VFO1 VFOID = iota
    VFO2

    VFOCount
)

type VFO interface {
    XITControl
    Name() string         // NEW
    Notify(any)
    Refresh()
    SetFrequency(Frequency)
    SetBand(Band)
    SetMode(Mode)
    SetXIT(bool, Frequency)
}

type CurrentVFOListener interface {       // NEW
    CurrentVFOChanged(VFOID)
}
```

All five existing VFO event interfaces gained a leading `VFOID` parameter:

| Then | Now |
|------|-----|
| `VFOFrequencyChanged(Frequency)` | `VFOFrequencyChanged(VFOID, Frequency)` |
| `VFOBandChanged(Band)` | `VFOBandChanged(VFOID, Band)` |
| `VFOModeChanged(Mode)` | `VFOModeChanged(VFOID, Mode)` |
| `VFOXITChanged(bool, Frequency)` | `VFOXITChanged(VFOID, bool, Frequency)` |
| `VFOPTTChanged(bool)` | `VFOPTTChanged(VFOID, bool)` |

`CurrentVFOListener` is the new "focus moved" event channel; in the original single-VFO snapshot there was no notion of an active/focused VFO.

---

## 3. `core/vfo` package — VFO struct and Client interface

**VFO struct** now (`core/vfo/vfo.go:27–41`):

```go
type VFO struct {
    XITControl
    id   core.VFOID          // NEW
    name string               // NEW
    bandplan      bandplan.Bandplan
    logbook       Logbook
    client        Client
    offlineClient *offlineClient
    refreshing    bool
    asyncRunner   core.AsyncRunner
    listeners []any
}
```

Constructor signature (`vfo.go:43`):

```go
func NewVFO(id core.VFOID, name string, bandplan bandplan.Bandplan,
            logbook Logbook, asyncRunner core.AsyncRunner) *VFO
```

(was `NewVFO("VFO 1", bandplan, Logbook, asyncRunner)`).

**`vfo.Client` interface** — gained `VFOID` on the per-VFO setters:

```go
type Client interface {
    Notify(any)
    Active() bool
    Refresh()
    SetFrequency(core.VFOID, core.Frequency)   // CHANGED
    SetBand(core.VFOID, core.Band)             // CHANGED
    SetMode(core.VFOID, core.Mode)             // CHANGED
    SetXIT(bool, core.Frequency)               // unchanged (single XIT)
}
```

XIT deliberately remains unscoped — there is currently a single XIT control surface in the UI, tied to VFO1 (see TODOs).

**Online/offline contract is unchanged**: each `VFO` still uses an `offlineClient` until `SetClient(client)` connects it to the radio. Per-band cache lives in the `offlineClient` exactly as before. The `online()` check (`vfo.go:75`) is identical.

**Emit semantics**: each emit method now propagates `v.id` as the first argument so listeners can filter. The inbound listener methods on the VFO itself (`VFOFrequencyChanged(vfo, f)`, etc.) include an early-return guard `if vfo != v.id { return }` so cross-VFO chatter from a shared client is ignored.

`asyncRunner` UI-thread marshalling is unchanged.

---

## 4. `core/radio` package — bridge stays single, backends become dual-VFO

`Controller` still has exactly **one** `activeRadio` / `activeKeyer`. It is still constructed before VFOs (`app.go:209`).

The `vfo.Client` implementation methods take `VFOID` and forward it (`core/radio/radio.go:269–309`):

```go
func (c *Controller) SetFrequency(vfo core.VFOID, frequency core.Frequency)
func (c *Controller) SetBand(vfo core.VFOID, band core.Band)
func (c *Controller) SetMode(vfo core.VFOID, mode core.Mode)
func (c *Controller) SetXIT(active bool, offset core.Frequency)
```

New config options for hamlib (`radio.go:17–18`): `hamlibVFO1Option`, `hamlibVFO2Option`. `newHamlibClient()` reads both and hands them to `hamlib.New(...)`.

`SelectRadio` listener re-registration logic is unchanged: when the backend switches, every cached listener is re-`Notify`'d into the new backend (`radio.go:203–204`). VFO1 and VFO2 both ride this same wire.

**Backend status:**

- **TCI** has dual-VFO routing via `toTCIVFO()` (`core/tci/tci.go:331–332`).
- **Hamlib** is dual-VFO on both directions:
  - Polling: `pollDualVFO` reads `GetRigInfo` and fills `lastState[VFO1]`/`lastState[VFO2]`, with single-VFO fallback via `pollSingleVFO` when `vfo2` is unconfigured (`hamlib.go:139–224`).
  - Setting: `SetFrequency(vfo core.VFOID, …)`, `SetBand(vfo core.VFOID, …)`, `SetMode(vfo core.VFOID, …)` all take a `VFOID` and forward to `c.client.SetX(c.vfos[vfo], …)` (`hamlib.go:314, 323, 332`).

The remaining gap is **`SetXIT`** (`hamlib.go:344`): signature `SetXIT(active bool, offset core.Frequency)` with `// TODO: add the VFOID to all VFO-related Setters` and `vfo := core.VFO1` hardcoded inside, so XIT writes always target VFO1. (`Speed`/`Send`/`Abort` still use `hl.CurrVFO`, but those are keyer concerns, not per-VFO ones.)

---

## 5. `core/entry` — biggest behavioural surface change

The original Entry held one `core.VFO`. Today (`core/entry/entry.go`):

- `vfos []core.VFO` (line 128), initialised to a slice of `nullVFO` of length `core.VFOCount`.
- `focusedVFO core.VFOID` (line 151) tracks which VFO the typing user is currently driving.
- Per-VFO scratch state: `input[VFOID]`, `selectedBand[VFOID]`, `selectedMode[VFOID]`, etc. (constructor lines 105–115).

**New / changed methods:**

| Method | Then | Now |
|--------|------|-----|
| `SetVFO(VFO)` | One slot | `SetVFO(id core.VFOID, vfo core.VFO)` (line 231) — fills a slot in the slice, then `vfo.Notify(c)` |
| `SetFocusedVFO(VFOID)` | did not exist | New entrypoint (`entry.go:411`) — currently only updates the cursor, with a `// TODO: whatever else is necessary when the focused VFO changes` |

**Routing of user-driven commands:**

- `SetFrequency` (line 469) → **hardcoded** `c.vfos[core.VFO1].SetFrequency(...)` (only one frequency input field exists today).
- `SetBand` (lines 474, 505) → `c.vfos[c.focusedVFO].SetBand(band)` — band combo follows focus.
- `SetMode` (line 526) → `c.vfos[c.focusedVFO].SetMode(mode)`.
- `SetXITActive` (line 478) → **hardcoded** `core.VFO1` (single XIT switch in UI).
- `Refresh()` (line 921) → `c.vfos[c.focusedVFO].Refresh()`.

**Callbacks from VFOs:** all five (`VFOFrequencyChanged`, `VFOBandChanged`, `VFOModeChanged`, `VFOXITChanged`, `VFOPTTChanged`) now carry the originating `VFOID`. Entry routes them per-VFO (caching band/mode/frequency into the right slot of its scratch state). A few paths are still VFO1-only by intent — e.g. the parrot/TX-state handling in `VFOPTTChanged` (around line 605).

**`Log()` gated to VFO1**: `entry.go:709` — early-return when `focusedVFO != VFO1`. Comment: `// TODO: remove once we have real edit fields for both VFOs.` Logging from VFO2 is intentionally disabled until per-VFO call/exchange input lands.

---

## 6. App wiring — current `Startup()` block (`core/app/app.go:188–249`)

Construction order:

1. `entry.NewController(...)` (line 188) — same as before.
2. `radio.NewController(...)` (line 209) — must precede VFO construction now: each VFO calls `SetClient(c.Radio)` inline during the loop below.
3. **Loop** (lines 213–220):
   ```go
   c.VFOs = make([]*vfo.VFO, core.VFOCount)
   for vfoID := range len(c.VFOs) {
       v := vfo.NewVFO(core.VFOID(vfoID),
                       fmt.Sprintf("VFO %d", vfoID+1),
                       c.bandplan, c.Logbook, c.asyncRunner)
       v.SetClient(c.Radio)
       c.VFOs[vfoID] = v
       c.Entry.SetVFO(core.VFOID(vfoID), v)
       c.Logbook.Notify(v)
   }
   ```
4. `c.Bandmap.SetVFO(c.VFOs[core.VFO1])` — **VFO1 only**.
5. `c.Workmode.Notify(c.VFOs[core.VFO1])` — **VFO1 only**.
6. `c.Radio.SelectRadio(c.session.Radio1())` — still one radio selection.
7. `c.VFOs[core.VFO1].Notify(c.QTCController)` (line 249) — **VFO1 only**.

Both VFOs exist and both push events into Entry, but several downstream consumers (Bandmap, Workmode, QTC) intentionally listen only to VFO1 right now. That is consistent with the UI: there is one bandmap, one workmode panel, one QTC controller.

**Shutdown** (`app.go:391`): unchanged — `c.Radio.Stop()` and the remote server if present. No per-VFO teardown.

---

## 7. UI surface (`ui/entryView.go`, `ui/centralArea.go`)

The entry view has gained a parallel row of VFO2 widgets (`entryView.go:47–53, 137–169`): `vfo2Label`, `vfo2FrequencyLabel`, `vfo2Band`, `vfo2Mode`, `vfo2XITIndicator`, `vfo2TXIndicator`. `centralArea.go:127–132` places them in the layout.

All `SetFrequency / SetBand / SetMode / SetXIT / SetTXState` setters on the view now take a `VFOID` and dispatch to the right widget row (`entryView.go:302–406`).

There is **no second call/exchange edit field** — that remains the gating reason `Entry.Log()` is locked to VFO1.

---

## 8. Updated diagram

```
                  ┌────────────┐
                  │   app      │  Startup builds N=VFOCount VFOs
                  └─────┬──────┘
                        │ for each VFOID: NewVFO + SetClient + SetVFO
       ┌────────────────┼──────────────────────────┐
       ▼                ▼                          ▼
   ┌───────┐  SetVFO(id, v)   ┌─────────────┐  SetClient   ┌─────────┐
   │ entry │ ◄────────────── ┤ vfo[VFO1]    │ ───────────► │  radio  │
   │ vfos[]│  VFO1 events    │ vfo[VFO2]    │ ◄────────── (Notify both)
   │ focus │ ◄────────────── │              │  freq/band/  └────┬────┘
   └───────┘  VFO2 events    └──────┬───────┘  mode events       │
                                    │                            │
                                    ▼                            ▼
                            offlineClient                hamlib (dual-VFO
                            (per-band cache,             polling + setting)
                             per VFO instance)           / tci
```

Bandmap, Workmode, QTC remain attached but **only to `vfo[VFO1]`** at present.

---

## 9. Invariants — what held, what shifted

Held:
- **Single VFO event source**: still true per VFO instance. `vfo.emit*` is still the only egress; backends never bypass.
- **Late binding via type assertion**: still the only registration mechanism; producer side iterates with type assertions.
- **`asyncRunner` UI-thread contract**: unchanged in vfo. Radio still does not marshal.
- **Construction precedes wiring**: still respected.

Shifted:
- **Events are now tagged with `VFOID`**. Listeners must dispatch on the ID; the old "the VFO" assumption no longer holds.
- **Multiple VFO instances share one Client**. Each VFO filters inbound events from the shared client by its own `id`. The client must address commands to a specific VFO (via the `VFOID` argument on `SetFrequency`/`SetBand`/`SetMode`).
- **A user-input focus concept now exists in Entry** (`focusedVFO`) that did not exist before. Several downstream paths still hardcode VFO1.

---

## 10. Outstanding work (in-flight)

| File:line | TODO |
|-----------|------|
| `core/hamlib/hamlib.go:344` | `// TODO: add the VFOID to all VFO-related Setters` — `SetXIT` lacks `VFOID` and hardcodes `core.VFO1`. `SetFrequency`/`SetBand`/`SetMode` already take `VFOID`. |
| `core/entry/entry.go:411` `SetFocusedVFO` | `// TODO: whatever else is necessary when the focused VFO changes` — currently only updates the cursor; band/mode display refresh on focus change not implemented |
| `core/entry/entry.go:577` | `// TODO: add VFO parameter to XITActiveChanged` — XIT change handler hardcodes VFO1 |
| `core/entry/entry.go:709` | `// TODO: remove once we have real edit fields for both VFOs` — `Log()` early-returns if `focusedVFO != VFO1` |
| `core/entry/entry.go:979` | `// TODO: myNumber, myReport, myExchange are independent of the currently focused VFO` — design ambiguity: per-VFO vs global serial/report |
| `core/entry/entry.go:1082` | `// TODO: check if the entry's band is currently selected in one of the two VFOs` — bandmap → VFO routing needs to pick a VFO when a spot is taken |
| `core/app/app.go:221–249` | Several listeners attached to VFO1 only (Bandmap, Workmode, QTC). Decide which of these should become per-VFO and which stay VFO1-only |
| UI | No second call/exchange input fields exist; without them, VFO2 cannot be the source of a logged QSO |

---

## Verification (read-only)

- `core/core.go:1679–1721` — new `VFOID` type, updated listener interfaces, `CurrentVFOListener`.
- `core/vfo/vfo.go:12–20, 27–63, 135–215` — VFO struct, `Client` interface, emit/filter logic.
- `core/radio/radio.go:17–19, 167–249, 269–309` — VFO2 option, `SelectRadio` re-registration, `vfo.Client` impl with `VFOID`.
- `core/app/app.go:188–249` — current wiring block including VFO loop.
- `core/entry/entry.go:105–115, 128, 151, 231–238, 411–417, 469–605, 709–712, 1082` — slice-of-VFO storage, `focusedVFO`, routing.
- `core/hamlib/hamlib.go:139–351` and `core/tci/tci.go:331–332` — backend dual-VFO state.
- `ui/entryView.go:47–53, 137–169, 302–406` and `ui/centralArea.go:127–132` — VFO2 widgets.
- `git log --oneline master..2vfo` — commits of the second-VFO work.

## Outlook: upcoming implementation tasks

- show separate supercheck, callinfo, and message rows for the VFO2
- handle the TX VFO

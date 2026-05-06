# Dependencies & Collaboration: `app` ↔ `entry` ↔ `vfo` ↔ `radio`

## Context

Reference document describing how the four packages collaborate inside hellocontest. Useful before refactoring entry/VFO/radio wiring (e.g. Qt6 port work, or any change to listener/notifier topology).

---

## 1. Layered roles

| Package | Role |
|---------|------|
| `core/app` | Top-level composition root. Constructs every subsystem in `Controller.Startup()` and wires the listener graph. |
| `core/entry` | QSO data-entry state machine. Holds the current callsign/exchange, drives ESM macros, talks to keyer/callinfo/bandmap. Owns no radio state itself. |
| `core/vfo` | Logical VFO model. Caches band/mode/freq per band when offline; when online, delegates to a `Client` (the radio Controller). Single point of truth for VFO state events. |
| `core/radio` | Backend manager. Owns the active hamlib/TCI radio + cwdaemon keyer. Provides `vfo.Client` implementation; bridges listener subscriptions to whichever backend is currently selected. |

Direction of import: `app` → {`entry`, `vfo`, `radio`}; `entry` → `core` (uses `core.VFO` interface, no direct vfo import); `vfo` → `core` only; `radio` → `core` + `core/hamlib`, `core/tci`, `core/cwdaemon`. **No cyclic imports.** All cross-package contracts live in `core/core.go` (listener interfaces, `core.AsyncRunner`, `core.Frequency/Band/Mode`).

---

## 2. Construction order in `app.Controller.Startup()`

Critical lines in `core/app/app.go`:

| Line | Call | Notes |
|------|------|-------|
| 188 | `entry.NewController(Settings, clock, Logbook, QSOList, Bandmap, asyncRunner)` | Built before VFO/Radio — Entry exists in a "no VFO yet" state via `nullVFO`. |
| 209 | `vfo.NewVFO("VFO 1", bandplan, Logbook, asyncRunner)` | Constructs VFO with `offlineClient` already active (`SetClient(nil)` at vfo.go:53). |
| 216 | `radio.NewController(config.Radios(), config.Keyers(), bandplan)` | No active backend yet. |

**Wiring block (~lines 196–223):**

```
Entry.SetESMEnabled(session.ESM())
Entry.Notify(session)            // session learns ESM toggles
Entry.Notify(hamDXMap)           // hamDXMap learns logged callsigns
Bandmap.Notify(Entry)            // Entry receives bandmap selection events
Logbook.Notify(Entry)            // Entry resets after logbook load
QSOList.Notify(Entry)            // Entry hears QSO edits

Entry.SetVFO(VFO)                // Entry now drives + listens to VFO
VFO.Notify(Bandmap)              // Bandmap follows VFO band/freq
Bandmap.SetVFO(VFO)              // Bandmap can command VFO (goto spot)
Logbook.Notify(VFO)              // VFO.LogbookLoaded restores last band/mode
Workmode.Notify(VFO)             // mode switches forwarded

Radio.Notify(ServiceStatus)      // status icons
Bandmap.Notify(Radio)            // bandmap entries → tci spots
VFO.SetClient(Radio)             // Radio becomes vfo.Client
Radio.SetSendSpotsToTci(session.SendSpotsToTci())
Radio.SelectRadio(session.Radio1())
Radio.SelectKeyer(session.Keyer1())
```

The order matters: `VFO.SetClient(Radio)` must precede `Radio.SelectRadio(...)` so that when a backend connects, the vfo's listener chain is already attached.

---

## 3. Listener interfaces (defined in `core/core.go`)

| Interface | Method | Emitted by | Consumed by (typical) |
|-----------|--------|-----------|-----------------------|
| `VFOFrequencyListener` | `VFOFrequencyChanged(Frequency)` | vfo | entry, bandmap, VFO itself (offline cache) |
| `VFOBandListener` | `VFOBandChanged(Band)` | vfo | entry, bandmap, callinfo |
| `VFOModeListener` | `VFOModeChanged(Mode)` | vfo | entry, workmode |
| `VFOXITListener` | `VFOXITChanged(bool, Frequency)` | vfo | entry, ui |
| `VFOPTTListener` | `VFOPTTChanged(bool)` | vfo | entry (PTT-aware logic) |
| `ServiceStatusListener` | `StatusChanged(Service, bool)` | radio | service status indicator |
| (anonymous) `RadioSelected(string)` / `KeyerSelected(string)` | radio | session, ui |

The notifier pattern is uniform: each producer holds `listeners []any`; `Notify(listener any)` appends; `emitX` iterates and uses type assertion to filter (vfo.go:156–204, radio.go:115–161). VFO wraps every emit through `asyncRunner` (UI-thread marshalling); Radio does **not** — radio events arrive on UI thread already because backends route through their own dispatchers.

---

## 4. The `vfo.Client` bridge — how Radio plugs into VFO

`vfo.Client` interface (vfo.go:12–20):

```go
type Client interface {
    Notify(any)
    Active() bool
    Refresh()
    SetFrequency(core.Frequency)
    SetBand(core.Band)
    SetMode(core.Mode)
    SetXIT(bool, core.Frequency)
}
```

`radio.Controller` satisfies it (radio.go:111, 265, 272–298, 300). On `VFO.SetClient(Radio)` (vfo.go:58):

1. VFO calls `Radio.Notify(v)` → Radio appends VFO to its listener list.
2. When `Radio.SelectRadio(name)` runs (radio.go:169), Radio iterates **all** registered listeners and calls `activeRadio.Notify(listener)` (radio.go:205–207). The actual hamlib or TCI client thereby receives the VFO as a `VFOFrequencyListener / VFOBandListener / …`.
3. Radio backend now pushes frequency/band/mode/xit/PTT changes directly to VFO (`v.VFOFrequencyChanged(...)` etc., vfo.go:135–154).
4. VFO re-emits those to its own subscribers (entry, bandmap, …) and mirrors them into `offlineClient` so the cache stays warm.

### Online vs offline data flow

**Setting frequency from UI/Entry, online:**

```
Entry.SetFrequency  →  VFO.SetFrequency  →  client (Radio).SetFrequency
                                          →  activeRadio.SetFrequency  (hamlib/tci)
                                          →  backend roundtrip
                                          →  backend.VFOFrequencyChanged → VFO
                                          →  VFO.emitFrequencyChanged → all listeners
                                          →  offlineClient cache write (for next offline switch)
```

**Setting frequency, offline (no backend):**

```
Entry.SetFrequency  →  VFO.SetFrequency  →  offlineClient.SetFrequency
                                          →  VFO.emitFrequencyChanged → all listeners
```

`VFO.online()` (vfo.go:69) decides which path runs. Radio's `Active()` returns false until a backend is connected; until then, VFO uses its `offlineClient`.

---

## 5. Entry's collaboration surface

Entry (entry/entry.go) holds VFO via the `core.VFO` interface (no direct dependency on the vfo package). Setter at `entry.go:220` wires both directions: `vfo.Notify(c)` registers Entry as a VFO listener.

**Entry → VFO commands:**
- `SetFrequency` (entry.go:450, 892, 1052) — when user types a frequency or selects a bandmap entry.
- `SetBand` (entry.go:455, 486), `SetMode` (entry.go:508), `SetXIT` (entry.go:459).
- `Refresh()` (entry.go:892) — re-pull current state.

**Entry ← VFO callbacks:**
- `VFOFrequencyChanged` (entry.go:463), `VFOBandChanged` (entry.go:491), `VFOModeChanged` (entry.go:543), `VFOXITChanged` (entry.go:555), `VFOPTTChanged` (entry.go:563).

These keep Entry's "current QSO" context (band/mode/frequency stored on the in-progress QSO) synced with the radio.

Entry never imports `radio` directly. Radio is invisible to Entry; the only surface is the `core.VFO` abstraction.

---

## 6. App-level runtime calls (post-Startup)

The Controller funnels UI commands into the right subsystem:

- `SelectRadio` (app.go:902) → `Radio.SelectRadio` (and persists to session). Triggers full chain: backend connect → status events → VFO listener registration → frequency events → Entry/Bandmap/Callinfo refresh.
- `SelectKeyer` (app.go:918) → `Radio.SelectKeyer`.
- `Stop` / `DoubleStop` (app.go:931, 935) → mostly Entry actions (`Entry.Clear`, etc.).
- `XITActive` / `SetXITActive` (app.go:854, 858) → VFO XIT control.
- `Shutdown` (app.go:388) → `Radio.Stop()` to cleanly disconnect backends.

---

## 7. Diagram (compact)

```
                    ┌────────────┐
                    │   app      │ Startup wires graph
                    └─────┬──────┘
                          │ constructs + Notify/Set
       ┌──────────────────┼──────────────────────┐
       ▼                  ▼                      ▼
   ┌───────┐  Set/Notify  ┌─────┐  SetClient   ┌─────────┐
   │ entry │ ───────────► │ vfo │ ───────────► │  radio  │
   │       │ ◄─────────── │     │ ◄────────── (Notify)   │
   └───────┘   VFO events └──┬──┘  freq events   └────┬────┘
                             │                       │
                             ▼                       ▼
                       offlineClient            hamlib / tci /
                       (band cache)             cwdaemon backends
```

`bandmap`, `logbook`, `workmode`, `callinfo`, `session`, `hamDXMap` all attach to the same listener buses (VFO and Entry) but are out of scope for this document.

---

## 8. Key invariants

- **Single VFO event source:** Whether online or offline, `VFO.emit*` is the only place subscribers are notified — backends never bypass VFO.
- **Late binding via type assertion:** `Notify(any)` accepts anything; producers select via type assertion. Adding a new listener interface requires only (a) defining it in `core`, (b) calling its method in the relevant emit loop. No changes to `Notify` signatures.
- **`asyncRunner` contract:** VFO marshals every emit through `asyncRunner` (UI thread). Listeners can touch UI without locking. Radio does not — its callers are expected to be on the right thread already.
- **Construction precedes wiring:** All `NewX` calls happen before any `Notify` / `Set*` calls, so listeners can never miss the construction phase.

---

## Verification (read-only)

- Re-read `core/app/app.go:147–323` to confirm wiring order matches above.
- `core/vfo/vfo.go:12–20, 58–63, 135–204` for Client interface + emit logic.
- `core/radio/radio.go:111, 169–218, 265–298` for the bridge.
- `core/entry/entry.go:220–226, 450–563` for Entry's VFO-facing methods.
- `core/core.go:1689–1705` for listener interface definitions.

# Entry Test Harness — Architecture

## Goal

A test harness for `core/entry` that lets authors write tests mirroring the use case descriptions in `entry_usecases.md`. Each test should read like a slightly abbreviated version of the corresponding use case: Pre conditions established by setup or action chains, the triggering action, then Post assertions.

## Package and visibility

The harness lives in **`package entry_test`** (black-box). Tests assert only through:
- Public methods on `*Controller`
- Spy recordings from all output channels

Internal struct fields are never inspected. This keeps tests honest about observable contracts and lets them survive internal refactors.

## Core abstraction: `Scenario`

```go
type Scenario struct { ... }

func NewScenario(t *testing.T) *Scenario
```

`Scenario` owns the controller and all its dependencies as spies. Every chainable method returns `*Scenario` for fluent chaining.

### Method categories

Methods fall into three categories. Only **action** methods reset the spy log before executing.

| Category | Examples | Spy reset? |
|----------|----------|-----------|
| Setup | `WithClassicExchange()`, `WithLogbookQSOs(...)`, `WithBand("40m")`, `WithESMEnabled(true)` | No |
| Action | `Enter(text)`, `PressEnter()`, `Clear()`, `GotoNextField()`, `SelectQSO(qso)`, `VFOFrequencyChanged(vfo, f)` | **Yes** |
| Assertion | `AssertCallsign(vfo, text)`, `AssertQSOLogged(qso)`, `AssertActiveField(vfo, field)`, `AssertNoQSOLogged()`, `AssertDuplicateMarker(vfo, bool)`, `AssertMessage(vfo, ...)`, `AssertNoMessage(vfo)` | No |

The spy reset on action methods means: every `Assert...()` call checks what happened *as a direct result of the immediately preceding action*. The test author never manages spy lifecycle.

### Example: use case A1 (Enter callsign)

```go
func TestA1_EnterCallsign(t *testing.T) {
    NewScenario(t).
        WithClassicExchange().
        Enter("DL1ABC").
        AssertCallsign(core.VFO1, "DL1ABC").
        AssertSerialClaimed().
        AssertNoMessage(core.VFO1)
}
```

### Example: use case C1 (Log valid QSO)

```go
func TestC1_LogValidQSO(t *testing.T) {
    NewScenario(t).
        WithClassicExchange().
        WithBand("40m").WithMode("CW").
        Enter("DL1ABC").GotoNextField().
        Enter("559").GotoNextField().
        Enter("042").GotoNextField().
        Enter("thx").
        PressEnter().
        AssertQSOLogged(core.QSO{Callsign: dl1abc, ...}).
        AssertActiveField(core.VFO1, core.CallsignField)
}
```

### Example: invariant as manual assertion

```go
// Use case C1 invariant: other VFO's input unchanged
scenario.
    PressEnter().
    AssertQSOLogged(expected).
    AssertCallsign(core.VFO2, "")  // VFO2 callsign not touched
```

## Spy design

Each output channel has a dedicated spy struct that records every call.

### `viewSpy`

Implements `entry.View`. Records every call as a typed entry in a log slice. Action reset clears the log.

```go
type viewCall struct {
    Method string
    Args   []any
}

type viewSpy struct {
    log []viewCall
}
```

Assertion helpers search the log:
- `AssertCalled(method, args...)` — fails if no matching entry found
- `AssertNotCalled(method)` — fails if any entry with that method found
- Higher-level helpers (e.g. `AssertDuplicateMarker(vfo, val)`) call these internally

### `logbookSpy`

Implements `entry.Logbook`. Records `AddQSO` and `UpdateQSO` calls. Configurable return values for `NextQSONumber`, `LastBand`, `LastMode`, `LastExchange` (set via setup methods).

```go
type logbookSpy struct {
    addedQSOs   []core.QSO
    updatedQSOs []core.QSO
    nextNumber  core.QSONumber
    lastExchange []string
    // ...
}
```

### `bandmapSpy`

Implements `entry.Bandmap`. Records `Add` and `SelectByCallsign` calls.

### `callinfoSpy`

Implements `entry.Callinfo`. Records `InputChanged` calls. Also provides a way to inject a `CallinfoFrame` (for prediction tests).

### `listenerSpy`

Implements `core.CallsignEnteredListener` and `core.CallsignLoggedListener`. Records emissions. Registered via `controller.Notify(listenerSpy)`.

## Setup presets

The full preset list is intentionally small. Add a preset only when the action chain to reach the state would be long boilerplate irrelevant to the test.

| Preset | Purpose |
|--------|---------|
| `WithClassicExchange()` | RST + serial + generic text fields, serial generated |
| `WithExchangeFields(n)` | N generic text fields |
| `WithLogbookQSOs(qsos...)` | Pre-populate logbook spy's duplicate/number state |
| `WithBand(s)` | Set initial band without going through band-field action |
| `WithMode(s)` | Set initial mode |
| `WithESMEnabled(bool)` | Toggle ESM |
| `WithVFO2Enabled()` | Enable dual-VFO mode |
| `WithWorkmode(w)` | Set Run or S&P |
| `WithCallinfoFrame(vfo, frame)` | Inject a prediction frame |

Pre conditions that *can* be reached via short action chains should use action chains, not presets.

## Spy reset mechanics

`Scenario` holds a reference to every spy. Every action method calls `s.resetSpies()` as its first line before forwarding to the controller. Setup and assertion methods do not call reset.

```go
func (s *Scenario) Enter(text string) *Scenario {
    s.resetSpies()
    s.controller.Enter(text)
    return s
}

func (s *Scenario) resetSpies() {
    s.view.reset()
    s.logbook.resetCalls()
    s.bandmap.resetCalls()
    s.callinfo.resetCalls()
    s.listener.resetCalls()
}
```

## Async runner

The controller accepts a `core.AsyncRunner`. The harness always uses synchronous inline execution (`func(f func()) { f() }`), so spy recordings are always complete before the action method returns. No goroutines, no sleeps, no races.

## Relationship to existing tests

Existing tests in `entry_test.go` use testify mocks and remain untouched. The harness is additive. New use-case tests live in a separate file (e.g., `usecases_test.go`). Migration of existing tests to the harness is deferred to a later pass.

## File layout

```
core/entry/
  entry.go
  esm.go
  null.go
  entry_test.go          # existing tests, unchanged
  esm_test.go            # existing tests, unchanged
  scenario_test.go       # Scenario type + spies + helpers (package entry_test)
  usecases_test.go       # use-case tests using Scenario (package entry_test)
```

## Naming convention

Test functions reference the use case ID:

```go
func TestA1_EnterCallsign(t *testing.T)
func TestC1_LogValidQSO(t *testing.T)
func TestD2_ClearFocusedVFO(t *testing.T)
```

This makes it trivial to cross-reference a failing test with the corresponding use case spec.

# The Spot Window

The spot list lives in its own top-level window (`ui/spotWindow.go`), not in a dock of the
main window. The spot window is a second `QMainWindow`, so it provides dock areas of its own
and can host the dockable views (QSO list, rate, score, clock, ...) next to the spot list.

```mermaid
graph TD
    subgraph main["QMainWindow (main)"]
        mainCentral["central widget:<br/>centralArea"]
        mainDocks["docks:<br/>qsoTable, qtcTable, rate,<br/>scoreGraph, scoreTable, clock"]
    end
    subgraph spots["QMainWindow (spots)"]
        spotsCentral["central widget:<br/>spotsView.widget"]
        spotsDocks["docks:<br/>(none yet)"]
    end
    main -. "no drag between the windows,<br/>AddDockWidget only" .-> spots
```

## Qt constraint: docks cannot be dragged between main windows

Qt6 offers **no** way to let the user drag a `QDockWidget` from one `QMainWindow` into
another one. The drag target is resolved through the dock's *parent chain*, in
`qtbase/src/widgets/widgets/qdockwidget.cpp`:

```cpp
static const QMainWindow *mainwindow_from_dock(const QDockWidget *dock)
{
    for (const QWidget *p = dock->parentWidget(); p; p = p->parentWidget())
        if (const QMainWindow *window = qobject_cast<const QMainWindow*>(p))
            return window;
    return nullptr;
}
```

`QDockWidgetPrivate::mouseMoveEvent()` only hovers the layout of that one main window. Other
top-level main windows are never hit-tested, no matter how the dock is configured. The
third-party frameworks that do support this (KDDockWidgets, Qt Advanced Docking System) are
C++-only and have no miqt bindings.

What *does* work is moving a dock programmatically:

```go
spotWindow.window.AddDockWidget(qtlib.RightDockWidgetArea, view.Dock())
```

`AddDockWidget` reparents the dock, and afterwards the dock drags normally *inside* its new
main window. The planned UI for this is a Window menu entry per dockable view ("Move to Main
Window" / "Move to Spot Window"), not drag & drop.

## Persisted state

Both windows store geometry and dock layout separately via `QSettings`
(`ui/windowState.go`):

| Key | Content |
| --- | --- |
| `ui/mainWindowGeometry`, `ui/mainWindowState` | main window, unchanged keys from before the split |
| `ui/spotWindowGeometry`, `ui/spotWindowState` | spot window |
| `ui/spotWindowVisible` | whether the spot window was open on exit |

`QMainWindow::restoreState()` only restores docks that are children of that window at the
time of the call. Once docks can be assigned to either window, the assignment has to be
stored separately and applied with `AddDockWidget` **before** `restoreState()` runs.
`RestoreDockWidget()` covers docks that are created later.

## Lifecycle

- The spot window is unparented, so it gets its own taskbar entry and independent stacking.
- Closing it only hides it: `QWidget::close()` hides the widget unless `WA_DeleteOnClose`
  is set, so the dock layout survives. Window ▸ Spots brings it back via
  `Controller.ShowSpots()` (`core/app/app.go`).
- Do **not** veto the spot window's close event to keep it open: quitting the application
  closes all top-level windows first, and a vetoed close also vetoes the quit. The
  application then keeps running with all windows hidden.
- Closing the main window closes the spot window, too (`ui/app.go`), otherwise the
  application would keep running without a visible main window.
- The window state of both windows is stored by `storeWindowStateOnce()`, called from the
  main window's close event and from `Quit()`. It only stores on the first call, because
  quitting closes all windows and the spot window may already be closed on later calls.

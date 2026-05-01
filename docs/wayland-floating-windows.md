# Wayland-Compatible Floating Windows (QDockWidget, QDialog, QFileDialog, QMessageBox)

Every Qt6 window that may appear detached from the main window — floating docks,
modal and modeless dialogs, file/message dialogs, message boxes — must be
implemented so the user can

1. see it as a real top-level window (not pinned above the main window),
2. drag it by its title bar to anywhere on any connected monitor,
3. close/redock it normally.

Gnome on Wayland (our primary test environment, tested on Ubuntu 25.10 +
Gnome 49 + Wayland) exposes two independent defects that affect Qt windows
created with the default flags:

- **`Qt::Tool` subsurfaces are pinned to the parent.** The Wayland
  compositor treats `Qt::Tool` windows as subsurfaces anchored to the
  main window. They are visually stuck in the center of the parent and
  cannot be dragged, even though the window decoration accepts clicks.
- **Qt cannot grab the pointer outside popup windows.** QtWayland prints
  `This plugin supports grabbing the mouse only for popup windows`
  whenever Qt code calls `grabMouse()` on a non-popup. This breaks any
  widget that implements its own drag via mouse grab — notably the
  built-in `QDockWidget` titlebar.

The fixes below are required; do not omit them "because it works on my
machine" — on X11 or macOS the bug is invisible.

## QDockWidget — floating docks

**Problem.** When the user undocks a `QDockWidget` on Wayland, Qt gives
the floating window `Qt::Tool` flags, which the compositor pins to the
parent. Additionally, the dock's internal titlebar uses `grabMouse()`,
which Wayland refuses outside popups — the user can grab the titlebar
but the window never moves.

**Fix.** Hook `OnTopLevelChanged` and, when the dock becomes top-level,
replace its window flags with a proper `Qt::Window` set. Keep the dock's
built-in titlebar so the user can still double-click or drag to redock;
the compositor's own window decoration handles moving.

```go
a.scoreDock.OnTopLevelChanged(func(topLevel bool) {
    if !topLevel {
        return
    }
    a.scoreDock.SetWindowFlags(
        qtlib.Window |
            qtlib.CustomizeWindowHint |
            qtlib.WindowTitleHint |
            qtlib.WindowCloseButtonHint |
            qtlib.WindowMinimizeButtonHint |
            qtlib.WindowMaximizeButtonHint,
    )
    a.scoreDock.Show()
})
```

Notes:
- Use `SetWindowFlags` (plural, full replacement). `SetWindowFlag2(Window, true)`
  does **not** work — it only toggles the `Window` bit, leaving `Tool`
  (value 11, which overlaps `Window` = 1) in place.
- Do **not** replace the dock's titlebar with an empty widget — that
  breaks the drag-to-redock and double-click-to-redock paths. Two title
  bars (native compositor on top, Qt dock titlebar below) is the
  intended result.
- Reference implementation: `ui/qt/app.go` in Step 4 (Score Dock).

## QDialog — modal and modeless dialogs

**Problem.** `QDialog`'s default flags (`Qt::Dialog`) can be mapped by
QtWayland to a subsurface of the parent main window, leaving the dialog
pinned above the center of the main window with no way to move it across
monitors.

**Fix.** When creating any `QDialog` instance (including subclasses),
explicitly set `Qt::Window`-based flags instead of relying on the
default `Qt::Dialog`:

```go
dialog := qtlib.NewQDialog(parent.QWidget)
dialog.SetWindowTitle("…")
dialog.SetWindowFlags(
    qtlib.Window |
        qtlib.CustomizeWindowHint |
        qtlib.WindowTitleHint |
        qtlib.WindowCloseButtonHint,
)
// Modal dialogs still call dialog.SetModal(true) and dialog.Exec();
// modeless dialogs call dialog.SetModal(false) and dialog.Show().
```

Apply this to **every** QDialog-based window created in Step 7:
settings, keyer settings, new contest, summary, export cabrillo, and
QTC. The same applies to any additional dialogs introduced in later
steps (polish, about box, sponsors, etc.).

## QFileDialog and QMessageBox — static helpers

**Problem.** The static helpers `QFileDialog_GetOpenFileName4`,
`QFileDialog_GetSaveFileName4`, `QMessageBox_Information`,
`QMessageBox_Question`, and `QMessageBox_Critical` construct an
internal QDialog with default flags and parent it to the main window.
On Wayland the dialog is stuck above the main window and cannot be
moved.

**Fix — preferred.** Use the native platform dialog via the
xdg-desktop-portal. These are separate top-level windows managed by the
compositor and do not have the subsurface problem. For `QFileDialog`
this is the default as long as `DontUseNativeDialog` is not set — keep
it that way. Gnome provides a portal-backed file chooser that works
correctly.

**Fix — fallback (when the non-native Qt dialog must be used, or for
`QMessageBox` which has no native Wayland backend).** Construct an
instance with explicit `Qt::Window` flags instead of using the static
helper:

```go
// QFileDialog instance form:
dlg := qtlib.NewQFileDialog(parent.QWidget)
dlg.SetWindowTitle(title)
dlg.SetDirectory(dir)
dlg.SetNameFilter(filter)
dlg.SetWindowFlags(
    qtlib.Window |
        qtlib.CustomizeWindowHint |
        qtlib.WindowTitleHint |
        qtlib.WindowCloseButtonHint,
)
if dlg.Exec() == int(qtlib.QDialog__Accepted) {
    files := dlg.SelectedFiles()
    // …
}

// QMessageBox instance form (for Info/Question/Critical):
mb := qtlib.NewQMessageBox(parent.QWidget)
mb.SetWindowTitle(title)
mb.SetText(msg)
mb.SetIcon(qtlib.QMessageBox__Information) // or Warning/Critical/Question
mb.SetWindowFlags(
    qtlib.Window |
        qtlib.CustomizeWindowHint |
        qtlib.WindowTitleHint |
        qtlib.WindowCloseButtonHint,
)
mb.Exec()
```

If the current static-helper-based implementations in `ui/qt/app.go`
(`SelectOpenFile`, `SelectSaveFile`, `ShowInfoDialog`,
`ShowQuestionDialog`, `ShowErrorDialog`) exhibit the stuck-dialog bug on
Wayland, replace them with the instance form shown above.

## Verification checklist

When you add or modify any window that can appear detached from the
main window, verify on Wayland (Gnome 49 / Ubuntu 25.10) that:

- [ ] The window appears as a real top-level window with native
      decorations.
- [ ] The window can be dragged anywhere on the current monitor.
- [ ] The window can be dragged to a different monitor (if available).
- [ ] No `This plugin supports grabbing the mouse only for popup
      windows` warning is printed to stderr during interaction.
- [ ] For docks: dragging the Qt dock titlebar to the main window edge
      redocks. Double-clicking the Qt dock titlebar redocks.
- [ ] For dialogs: closing the dialog via the close button, the Cancel
      button, or pressing Escape all work.

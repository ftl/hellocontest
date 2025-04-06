# Concurrent Programming Guide

This guide defines some conventions for how concurrent programming is done in Hellocontest.

## Thread Confinement

The main thread of the application is the UI thread. Whenever a component uses a separate
goroutine, it must make sure to communicate with other components through the main thread
by using the `asyncRunner` provided by the `core/app.Controller'.

## Thread Safety

The following structs are thread-safe and can be used from any goroutine:
- `core/callhistory.Finder`
- `core/dxcc.Finder`
- `core/logbook.QSOList`
- `core/score.Counter`
- `core/scp.Finder`

Thread-safe structs are marked with the comment `// xxx is thread-safe`.

## UI Components

Code in the `ui` package must not take care of thread confinement. Any synchronization between
goroutines and the UI thread must happen in the `core` package, before code from the `ui`
package is called.

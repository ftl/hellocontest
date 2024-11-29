# Bandmap Thread Confinement

The following is a detailed analysis of the thread confinement strategy used in the bandmap feature. I use a bunch of mermaid diagrams to visualize a detailed view on most stimuli using sequence diagrams.

In context of the bandmap feature, keeping track of which thread is doing what is rather complicated because of the complex event system of the tree view widget. This analysis helped me a lot to keep things nice and working.

## Select the visible band

```mermaid
sequenceDiagram
actor ui

ui ->> spotsView: selectBand

spotsView ->> SpotsController: SetVisibleBand

SpotsController ->> SpotsController: do
activate SpotsController
critical [confine to bandmap, sync]
    note right of SpotsController: set visible frequency for next frame
    SpotsController ->> SpotsController:  update
    SpotsController ->> spotsView:  ShowFrame
end
deactivate SpotsController

spotsView ->> spotsView: runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsView: setupBands
    spotsView ->> spotsView: updateBands
    spotsView ->> spotsTableView: showInitialFrameInTable
    spotsView ->> spotsTableView: refreshTable
    spotsView ->> spotsTableView: updateHighlightedColumns
    spotsView ->> spotsTableView: revealTableEntry
    spotsTableView ->> SpotsController: EntryVisible
    SpotsController ->> SpotsController: do
    activate SpotsController
    critical [confine to bandmap, sync]
        SpotsController ->> SpotsController: entryVisible
    end
    deactivate SpotsController
end
deactivate spotsView

```

## Select the active band

```mermaid
sequenceDiagram
actor ui

ui ->> spotsView:selectBand

spotsView ->> SpotsController:SetActiveBand
SpotsController ->> vfo:SetBand

activate vfo
vfo ->> SpotsController:VFOFrequencyChanged
SpotsController ->> SpotsController:do
note right of SpotsController: set activeFrequency for next frame
vfo ->> SpotsController:VFOBandChanged
deactivate vfo

SpotsController ->> SpotsController:do
activate SpotsController
critical [confine to bandmap, sync]
    note right of SpotsController: set activeBand, visibleBand for the next frame
    SpotsController ->> SpotsController:update
    SpotsController ->> spotsView:ShowFrame
end
deactivate SpotsController

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsView:setupBands
    spotsView ->> spotsView:updateBands
    spotsView ->> spotsTableView:showInitialFrameInTable
    spotsView ->> spotsTableView:refreshTable
    spotsView ->> spotsTableView:updateHighlightedColumns
    spotsView ->> spotsTableView:revealTableEntry
    spotsTableView ->> SpotsController:EntryVisible
    SpotsController ->> SpotsController:do
    activate SpotsController
    critical [confine to bandmap, sync]
        SpotsController ->> SpotsController:entryVisible
    end
    deactivate SpotsController
end
deactivate spotsView
```

## Select a table entry

```mermaid
sequenceDiagram
actor ui
ui ->> spotsTableView:onTableSelectionChanged

spotsTableView ->> spotsTableView:getSelectedIndex
spotsTableView ->> gtk:call
spotsTableView ->> SpotsController:SelectEntry

SpotsController ->> SpotsController:do
activate SpotsController
critical [confine to bandmap, sync]
    SpotsController ->> entries:Select
    entries ->> spotsView:EntrySelected
end
deactivate SpotsController

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsTableView:revealTableEntry
    spotsTableView ->> SpotsController:EntryVisible
    SpotsController ->> SpotsController:do
    activate SpotsController
    critical [confine to bandmap, sync]
        SpotsController ->> SpotsController:entryVisible
    end
    deactivate SpotsController
end
deactivate spotsView
```

## Add an entry coming from the DX cluster

```mermaid
sequenceDiagram
SpotsController ->> spotsView:EntryAdded

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsTableView:addTableEntry

    spotsTableView ->> gtk.ListStore:Insert
    spotsTableView ->> spotsTableView:fillEntryToTableRow
    spotsTableView ->> gtk.ListStore:Set
end
deactivate spotsView
```

## Update an existing table entry due to changes coming from the DX cluster

```mermaid
sequenceDiagram
SpotsController ->> spotsView:EntryUpdated

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsTableView:updateTableEntry

    spotsTableView ->> spotsTableView:tableRowByIndex
    spotsTableView ->> gtk.ListStore:GetIterFromString
    spotsTableView ->> spotsTableView:fillEntryToTableRow
    spotsTableView ->> gtk.ListStore:Set
end
deactivate spotsView
```

## Remove a stale table entry

```mermaid
sequenceDiagram
SpotsController ->> spotsView:EntryRemoved

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsTableView:removeTableEntry

    spotsTableView ->> spotsTableView:tableRowByIndex
    spotsTableView ->> gtk.ListStore:GetIterFromString
    spotsTableView ->> gtk.ListStore:Remove
end
deactivate spotsView
```

## Reveal an entry

```mermaid
sequenceDiagram
SpotsController ->> spotsView:RevealEntry

spotsView ->> spotsView:runAsync
activate spotsView
critical [later in the UI]
    spotsView ->> spotsTableView:revealTableEntry

    spotsTableView ->> SpotsController:EntryVisible
    SpotsController ->> SpotsController:do
    activate SpotsController
    critical [confine to bandmap, sync]
        SpotsController ->> SpotsController:entryVisible
    end
    deactivate SpotsController

    spotsTableView ->> gtk.ListStore:GetIterFromString
    spotsTableView ->> gtk.ListStore:GetPath
    spotsTableView ->> gtk.TreeView:GetColumn
end
deactivate spotsView
```
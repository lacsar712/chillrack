# chillrack

Wastewater cooling biorack defrost coordination service written in Go.

## Overview

`chillrack` coordinates multi-cell biorack plants: coolant level monitoring, cooling zones, compressor air scour, manifold valve actuation, and sequenced defrost phases. The service exposes a JSON HTTP API (no bundled frontend).

## Build and test

```bash
cd chillrack
go build ./...
go test ./...
```

## Run

```bash
go run ./cmd/chillrack
```

## Key flows

1. `app.RequestDefrost` acquires an interlock lease (`defer unlock`), then `defrost.Coordinator.Run` opens valves via `manifold.ValveBank.Open`.
2. `fsm.WashMachine.Transition` calls `defrost.Emitter.Emit`.
3. `config.WashCloseWindow` is enforced via `clock.ProcessClock` in `defrost.Window`.
4. `store.RackStore.Snapshot` returns deep-copied slices.

# Idiomatic test architecture

`contract/` contains implementation-neutral helpers. Thin tests in each
`starter/monitor` and `solution/monitor` root adapt their local packages to
those helpers.

Add deterministic fixtures under `fixtures/` and progressive shared contracts
under `m1/` through `m5/`. Tests should accept factories or small interfaces so
the same assertions can run against either implementation tree.

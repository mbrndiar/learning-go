# Comparative test architecture

`contract/` contains implementation-neutral helpers. Thin tests in each
`starter/kvstore` and `solution/kvstore` root adapt their local packages to
those helpers. Future `m1` through `m5` suites should follow the same factory
pattern instead of copying assertions between targets.

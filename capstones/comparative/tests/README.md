# Comparative test architecture

`contract/` contains the fixture reader, exact envelope assertions, SQLite
setup/assertion support, process barrier, lock helper, timeout handling, and
cleanup checks. `m1` through `m5` expose the frozen learner milestones without
copying assertions into an implementation tree.

These checks extend Module 11 from focused temporary databases to a versioned
SQLite file, migration rules, transaction races, and independent verification.

The solution adapter runs:

1. domain keys, expectations, restricted JSON, and typed errors;
2. exact CLI grammar, validation precedence, envelopes, and exit codes;
3. initialization, v0 migration, v1 invariants, and ordinary persistence;
4. boundaries, binary ordering, revisions, CAS, and deletes; and
5. real child-process initialization/migration races, contention, and busy
   waiting/timeouts.

Every scenario uses a repository-local temporary directory. The runner closes
SQLite resources, terminates only recorded child PIDs/process groups, verifies
integrity, and removes database/WAL sidecars before deleting the directory.

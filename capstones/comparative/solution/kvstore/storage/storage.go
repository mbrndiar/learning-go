// Package storage defines persistence capabilities consumed by the application.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/capstones/comparative/solution/kvstore/domain"
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	busyTimeoutMS  = 10000
	createMetadata = `CREATE TABLE store_metadata (
    singleton       INTEGER PRIMARY KEY CHECK (singleton = 1),
    schema_version  INTEGER NOT NULL CHECK (schema_version = 1),
    global_revision INTEGER NOT NULL
                    CHECK (global_revision BETWEEN 0 AND 9007199254740991)
)`
	createEntries = `CREATE TABLE entries (
    key        TEXT PRIMARY KEY COLLATE BINARY,
    value_json TEXT NOT NULL CHECK (json_valid(value_json)),
    revision   INTEGER NOT NULL
               CHECK (revision BETWEEN 1 AND 9007199254740991)
)`
	insertMetadata = `INSERT INTO store_metadata(singleton, schema_version, global_revision)
VALUES (1, 1, 0)`

	// Canonical fingerprints recognize legacy v0/v1 schemas independent of
	// formatting; see canonicalSQL for the normalization they rely on.
	v0EntriesCanonical  = "createtableentries(keytextprimarykeycollatebinary,value_jsontextnotnull)"
	v1EntriesCanonical  = "createtableentries(keytextprimarykeycollatebinary,value_jsontextnotnullcheck(json_valid(value_json)),revisionintegernotnullcheck(revisionbetween1and9007199254740991))"
	v1MetadataCanonical = "createtablestore_metadata(singletonintegerprimarykeycheck(singleton=1),schema_versionintegernotnullcheck(schema_version=1),global_revisionintegernotnullcheck(global_revisionbetween0and9007199254740991))"
)

// Store is the persistence boundary used by the comparative application.
type Store interface {
	Set(context.Context, string, domain.Value, domain.Expectation) (domain.SetResult, error)
	Get(context.Context, string) (domain.Entry, error)
	Delete(context.Context, string, domain.Expectation) (domain.DeleteResult, error)
	List(context.Context) (domain.ListResult, error)
	Close() error
}

// Opener creates a store for one literal database path.
type Opener interface {
	Open(context.Context, string) (Store, error)
}

// SQLiteOpener opens stores through database/sql and the pinned pure-Go driver.
type SQLiteOpener struct{}

// NewSQLiteOpener constructs the SQLite opener boundary.
func NewSQLiteOpener() *SQLiteOpener {
	return &SQLiteOpener{}
}

// Open configures, initializes or migrates, and validates one SQLite store.
func (*SQLiteOpener) Open(ctx context.Context, path string) (Store, error) {
	parent := filepath.Dir(path)
	if info, err := os.Stat(parent); err != nil || !info.IsDir() {
		if err == nil {
			err = fmt.Errorf("%s is not a directory", parent)
		}
		return nil, storageError("open", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, storageError("open", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	connection, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, mapSQLiteError(err, "open")
	}
	store := &sqliteStore{db: db, connection: connection}
	if err := store.configure(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	if err := store.prepareSchema(ctx); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

type sqliteStore struct {
	db         *sql.DB
	connection *sql.Conn
}

func (s *sqliteStore) configure(ctx context.Context) error {
	// busy_timeout lets SQLite block and retry internally on SQLITE_BUSY
	// write-lock contention before returning an error, so ordinary writer
	// contention resolves without surfacing a spurious failure.
	if _, err := s.connection.ExecContext(ctx, `PRAGMA busy_timeout = 10000`); err != nil {
		return mapSQLiteError(err, "configure")
	}
	// Switching to WAL can transiently fail while another connection holds
	// the file lock, so poll for confirmation within the busy_timeout budget
	// instead of trusting a single attempt.
	deadline := time.Now().Add(busyTimeoutMS * time.Millisecond)
	for {
		var journalMode string
		err := s.connection.QueryRowContext(ctx, `PRAGMA journal_mode = WAL`).Scan(&journalMode)
		if err == nil && strings.EqualFold(journalMode, "wal") {
			break
		}
		if err == nil {
			return storageError("configure", fmt.Errorf("journal mode is %q", journalMode))
		}
		if !isBusyError(err) || time.Now().After(deadline) {
			return mapSQLiteError(err, "configure")
		}
		select {
		case <-ctx.Done():
			return storageError("configure", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if _, err := s.connection.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return mapSQLiteError(err, "configure")
	}
	return nil
}

func (s *sqliteStore) Close() error {
	var result error
	if s.connection != nil {
		if err := s.connection.Close(); err != nil {
			result = fmt.Errorf("close SQLite connection: %w", err)
		}
		s.connection = nil
	}
	if s.db != nil {
		if err := s.db.Close(); err != nil && result == nil {
			result = fmt.Errorf("close SQLite database: %w", err)
		}
		s.db = nil
	}
	return result
}

func (s *sqliteStore) Set(
	ctx context.Context,
	key string,
	value domain.Value,
	expectation domain.Expectation,
) (domain.SetResult, error) {
	transaction, err := beginImmediate(ctx, s.connection, "write")
	if err != nil {
		return domain.SetResult{}, err
	}
	defer transaction.rollback()

	// Reading the current revision and validating the expectation inside
	// this same BEGIN IMMEDIATE transaction is what makes the check atomic:
	// the write lock is already held, so no concurrent writer can change or
	// remove the key between the check and the write below.
	currentRevision, err := queryEntryRevision(ctx, transaction, key, "write")
	if err != nil {
		return domain.SetResult{}, err
	}
	switch expectation.Kind {
	case domain.ExpectAny:
	case domain.ExpectAbsent:
		if currentRevision != nil {
			return domain.SetResult{}, conflict(
				key,
				"absent",
				domain.Revision(*currentRevision),
			)
		}
	case domain.ExpectExact:
		if currentRevision == nil || domain.Revision(*currentRevision) != expectation.Revision {
			var actual any
			if currentRevision != nil {
				actual = domain.Revision(*currentRevision)
			}
			return domain.SetResult{}, conflict(key, expectation.Revision, actual)
		}
	default:
		return domain.SetResult{}, storageError("write", errors.New("unknown expectation"))
	}

	revision, err := nextRevision(ctx, transaction)
	if err != nil {
		return domain.SetResult{}, err
	}
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return domain.SetResult{}, storageError("write", err)
	}
	if _, err := transaction.ExecContext(
		ctx,
		`INSERT INTO entries(key, value_json, revision) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE
		 SET value_json = excluded.value_json, revision = excluded.revision`,
		key,
		string(valueJSON),
		revision,
	); err != nil {
		return domain.SetResult{}, mapSQLiteError(err, "write")
	}
	if err := updateGlobalRevision(ctx, transaction, revision); err != nil {
		return domain.SetResult{}, err
	}
	if err := transaction.commit(ctx); err != nil {
		return domain.SetResult{}, err
	}
	return domain.SetResult{
		Key:      key,
		Value:    value,
		Revision: domain.Revision(revision),
		Created:  currentRevision == nil,
	}, nil
}

func (s *sqliteStore) Get(ctx context.Context, key string) (domain.Entry, error) {
	var valueJSON string
	var revision int64
	err := s.connection.QueryRowContext(
		ctx,
		`SELECT value_json, revision FROM entries WHERE key = ?`,
		key,
	).Scan(&valueJSON, &revision)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Entry{}, notFound(key)
	}
	if err != nil {
		return domain.Entry{}, mapSQLiteError(err, "read")
	}
	value, err := domain.ParseStoredValue(valueJSON)
	if err != nil {
		return domain.Entry{}, invalidStoredValue(key, err)
	}
	if revision < 1 || revision > int64(domain.MaxRevision) {
		return domain.Entry{}, revisionInvariant()
	}
	return domain.Entry{
		Key:      key,
		Value:    value,
		Revision: domain.Revision(revision),
	}, nil
}

func (s *sqliteStore) Delete(
	ctx context.Context,
	key string,
	expectation domain.Expectation,
) (domain.DeleteResult, error) {
	transaction, err := beginImmediate(ctx, s.connection, "write")
	if err != nil {
		return domain.DeleteResult{}, err
	}
	defer transaction.rollback()

	// Same compare-and-swap guarantee as Set: existence and expectation are
	// checked under the write lock this transaction already holds, so the
	// delete below cannot race with another writer's change to the key.
	currentRevision, err := queryEntryRevision(ctx, transaction, key, "write")
	if err != nil {
		return domain.DeleteResult{}, err
	}
	if currentRevision == nil {
		return domain.DeleteResult{}, notFound(key)
	}
	switch expectation.Kind {
	case domain.ExpectAny:
	case domain.ExpectExact:
		if domain.Revision(*currentRevision) != expectation.Revision {
			return domain.DeleteResult{}, conflict(
				key,
				expectation.Revision,
				domain.Revision(*currentRevision),
			)
		}
	default:
		return domain.DeleteResult{}, storageError("write", errors.New("unknown expectation"))
	}

	revision, err := nextRevision(ctx, transaction)
	if err != nil {
		return domain.DeleteResult{}, err
	}
	if _, err := transaction.ExecContext(ctx, `DELETE FROM entries WHERE key = ?`, key); err != nil {
		return domain.DeleteResult{}, mapSQLiteError(err, "write")
	}
	if err := updateGlobalRevision(ctx, transaction, revision); err != nil {
		return domain.DeleteResult{}, err
	}
	if err := transaction.commit(ctx); err != nil {
		return domain.DeleteResult{}, err
	}
	return domain.DeleteResult{
		Key:             key,
		DeletedRevision: domain.Revision(*currentRevision),
		Revision:        domain.Revision(revision),
	}, nil
}

func (s *sqliteStore) List(ctx context.Context) (domain.ListResult, error) {
	rows, err := s.connection.QueryContext(
		ctx,
		`SELECT entries.key, entries.value_json, entries.revision,
		        store_metadata.global_revision
		 FROM store_metadata
		 LEFT JOIN entries ON TRUE
		 WHERE store_metadata.singleton = 1
		 ORDER BY entries.key COLLATE BINARY`,
	)
	if err != nil {
		return domain.ListResult{}, mapSQLiteError(err, "read")
	}
	defer rows.Close()

	entries := make([]domain.Entry, 0)
	var globalRevision *int64
	for rows.Next() {
		var key, valueJSON sql.NullString
		var revision sql.NullInt64
		var rowGlobalRevision int64
		if err := rows.Scan(&key, &valueJSON, &revision, &rowGlobalRevision); err != nil {
			return domain.ListResult{}, mapSQLiteError(err, "read")
		}
		if globalRevision == nil {
			globalRevision = &rowGlobalRevision
		} else if *globalRevision != rowGlobalRevision {
			return domain.ListResult{}, revisionInvariant()
		}
		if !key.Valid {
			if valueJSON.Valid || revision.Valid {
				return domain.ListResult{}, malformedSchema()
			}
			continue
		}
		if !valueJSON.Valid || !revision.Valid {
			return domain.ListResult{}, malformedSchema()
		}
		value, err := domain.ParseStoredValue(valueJSON.String)
		if err != nil {
			return domain.ListResult{}, invalidStoredValue(key.String, err)
		}
		entries = append(entries, domain.Entry{
			Key:      key.String,
			Value:    value,
			Revision: domain.Revision(revision.Int64),
		})
	}
	if err := rows.Err(); err != nil {
		return domain.ListResult{}, mapSQLiteError(err, "read")
	}
	if globalRevision == nil ||
		*globalRevision < 0 ||
		*globalRevision > int64(domain.MaxRevision) {
		return domain.ListResult{}, revisionInvariant()
	}
	return domain.ListResult{
		Entries:        entries,
		GlobalRevision: domain.Revision(*globalRevision),
	}, nil
}

func (s *sqliteStore) prepareSchema(ctx context.Context) error {
	transaction, err := beginImmediate(ctx, s.connection, "initialize")
	if err != nil {
		return err
	}
	defer transaction.rollback()

	if version, recognizable := futureSchemaVersion(ctx, transaction); recognizable && version > 1 {
		return &domain.Error{
			Category: "unsupported_schema",
			Details:  map[string]any{"found": version, "supported": int64(1)},
		}
	}
	if err := ensureIntegrity(ctx, transaction); err != nil {
		return err
	}
	objects, err := applicationObjects(ctx, transaction)
	if err != nil {
		return err
	}
	if err := validateDefaultPragmas(ctx, transaction); err != nil {
		return err
	}

	// The three outcomes are mutually exclusive by construction: no
	// application objects (fresh database), an exact legacy v0 shape, or an
	// exact current v1 shape. Anything else is untrusted and rejected rather
	// than guessed at.
	switch {
	case len(objects) == 0:
		if err := initialize(ctx, transaction); err != nil {
			return err
		}
	case isExactV0(objects):
		if err := migrateV0(ctx, transaction); err != nil {
			return err
		}
	case isExactV1(objects):
		if err := validateV1(ctx, transaction); err != nil {
			return err
		}
	default:
		return malformedSchema()
	}
	return transaction.commit(ctx)
}

type immediateTransaction struct {
	connection *sql.Conn
	done       bool
	operation  string
}

func beginImmediate(
	ctx context.Context,
	connection *sql.Conn,
	operation string,
) (*immediateTransaction, error) {
	// BEGIN IMMEDIATE takes the write lock at the start of the transaction
	// rather than on the first write, serializing writers up front so the
	// compare-and-swap checks and writes that follow run against a lock
	// that's already held and stable, instead of risking a late upgrade
	// failure partway through.
	if _, err := connection.ExecContext(ctx, `BEGIN IMMEDIATE`); err != nil {
		return nil, mapSQLiteError(err, operation)
	}
	return &immediateTransaction{connection: connection, operation: operation}, nil
}

func (transaction *immediateTransaction) ExecContext(
	ctx context.Context,
	query string,
	args ...any,
) (sql.Result, error) {
	return transaction.connection.ExecContext(ctx, query, args...)
}

func (transaction *immediateTransaction) QueryContext(
	ctx context.Context,
	query string,
	args ...any,
) (*sql.Rows, error) {
	return transaction.connection.QueryContext(ctx, query, args...)
}

func (transaction *immediateTransaction) QueryRowContext(
	ctx context.Context,
	query string,
	args ...any,
) *sql.Row {
	return transaction.connection.QueryRowContext(ctx, query, args...)
}

func (transaction *immediateTransaction) commit(ctx context.Context) error {
	if transaction.done {
		return nil
	}
	if _, err := transaction.connection.ExecContext(ctx, `COMMIT`); err != nil {
		return mapSQLiteError(err, "commit")
	}
	transaction.done = true
	return nil
}

func (transaction *immediateTransaction) rollback() {
	// Deferred unconditionally after every begin: once commit has run this
	// is a no-op, and on any error path it guarantees the write lock is
	// released and no partial change is left visible.
	if transaction.done {
		return
	}
	_, _ = transaction.connection.ExecContext(context.Background(), `ROLLBACK`)
	transaction.done = true
}

type schemaObject struct {
	objectType string
	name       string
	sqlText    sql.NullString
}

func applicationObjects(
	ctx context.Context,
	transaction *immediateTransaction,
) ([]schemaObject, error) {
	rows, err := transaction.QueryContext(
		ctx,
		`SELECT type, name, sql
		 FROM sqlite_schema
		 WHERE name NOT LIKE 'sqlite_%'
		 ORDER BY type COLLATE BINARY, name COLLATE BINARY`,
	)
	if err != nil {
		return nil, mapSQLiteError(err, "read")
	}
	defer rows.Close()
	var objects []schemaObject
	for rows.Next() {
		var object schemaObject
		if err := rows.Scan(&object.objectType, &object.name, &object.sqlText); err != nil {
			return nil, mapSQLiteError(err, "read")
		}
		objects = append(objects, object)
	}
	if err := rows.Err(); err != nil {
		return nil, mapSQLiteError(err, "read")
	}
	return objects, nil
}

func isExactV0(objects []schemaObject) bool {
	return len(objects) == 1 &&
		objects[0].objectType == "table" &&
		objects[0].name == "entries" &&
		objects[0].sqlText.Valid &&
		canonicalSQL(objects[0].sqlText.String) == v0EntriesCanonical
}

func isExactV1(objects []schemaObject) bool {
	if len(objects) != 2 {
		return false
	}
	matches := 0
	for _, object := range objects {
		if object.objectType != "table" || !object.sqlText.Valid {
			return false
		}
		switch object.name {
		case "entries":
			if canonicalSQL(object.sqlText.String) != v1EntriesCanonical {
				return false
			}
			matches++
		case "store_metadata":
			if canonicalSQL(object.sqlText.String) != v1MetadataCanonical {
				return false
			}
			matches++
		default:
			return false
		}
	}
	return matches == 2
}

func canonicalSQL(value string) string {
	// Strips whitespace and quoting characters so schema comparisons aren't
	// sensitive to how SQLite echoes stored CREATE TABLE text or how the
	// original DDL happened to be authored.
	var canonical strings.Builder
	for _, character := range strings.ToLower(value) {
		switch character {
		case ' ', '\t', '\n', '\r', '"', '\'', '`', '[', ']':
		default:
			canonical.WriteRune(character)
		}
	}
	return canonical.String()
}

func futureSchemaVersion(
	ctx context.Context,
	transaction *immediateTransaction,
) (int64, bool) {
	var version int64
	err := transaction.QueryRowContext(
		ctx,
		`SELECT schema_version FROM store_metadata LIMIT 1`,
	).Scan(&version)
	return version, err == nil
}

func ensureIntegrity(ctx context.Context, transaction *immediateTransaction) error {
	rows, err := transaction.QueryContext(ctx, `PRAGMA integrity_check`)
	if err != nil {
		return integrityFailure(err)
	}
	defer rows.Close()
	var messages []string
	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			return integrityFailure(err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return integrityFailure(err)
	}
	if len(messages) != 1 || messages[0] != "ok" {
		return integrityFailure(errors.New("SQLite integrity check failed"))
	}
	return nil
}

func validateDefaultPragmas(ctx context.Context, transaction *immediateTransaction) error {
	var userVersion, applicationID int64
	if err := transaction.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&userVersion); err != nil {
		return malformedSchema()
	}
	if err := transaction.QueryRowContext(ctx, `PRAGMA application_id`).Scan(&applicationID); err != nil {
		return malformedSchema()
	}
	if userVersion != 0 || applicationID != 0 {
		return malformedSchema()
	}
	return nil
}

func initialize(ctx context.Context, transaction *immediateTransaction) error {
	for _, statement := range []string{createMetadata, createEntries, insertMetadata} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			return mapSQLiteError(err, "initialize")
		}
	}
	return nil
}

type legacyEntry struct {
	key       string
	valueJSON string
}

func migrateV0(ctx context.Context, transaction *immediateTransaction) error {
	// Detection (in prepareSchema) and this rewrite share the single
	// BEGIN IMMEDIATE transaction that already holds the write lock, so a
	// competing opener either blocks before this migration begins or, once
	// it acquires the lock afterward, observes the already-completed v1
	// schema; the rewrites themselves can never interleave.
	rows, err := transaction.QueryContext(
		ctx,
		`SELECT key, value_json FROM entries ORDER BY key COLLATE BINARY`,
	)
	if err != nil {
		return mapSQLiteError(err, "migrate")
	}
	var legacy []legacyEntry
	for rows.Next() {
		var entry legacyEntry
		if err := rows.Scan(&entry.key, &entry.valueJSON); err != nil {
			rows.Close()
			return mapSQLiteError(err, "migrate")
		}
		legacy = append(legacy, entry)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return mapSQLiteError(err, "migrate")
	}
	if err := rows.Close(); err != nil {
		return mapSQLiteError(err, "migrate")
	}

	for index := range legacy {
		if _, err := domain.ParseKey(legacy[index].key); err != nil {
			return &domain.Error{
				Category: "invalid_storage",
				Details: map[string]any{
					"reason": "invalid_key",
					"key":    legacy[index].key,
				},
				Cause: err,
			}
		}
		value, err := domain.ParseValue(json.RawMessage(legacy[index].valueJSON))
		if err != nil {
			return invalidStoredValue(legacy[index].key, err)
		}
		normalized, err := json.Marshal(value)
		if err != nil {
			return storageError("migrate", err)
		}
		legacy[index].valueJSON = string(normalized)
	}

	for _, statement := range []string{
		`ALTER TABLE entries RENAME TO entries_v0_migration`,
		createMetadata,
		createEntries,
		insertMetadata,
	} {
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			return mapSQLiteError(err, "migrate")
		}
	}
	for index, entry := range legacy {
		revision := int64(index + 1)
		if revision > int64(domain.MaxRevision) {
			return revisionExhausted()
		}
		if _, err := transaction.ExecContext(
			ctx,
			`INSERT INTO entries(key, value_json, revision) VALUES (?, ?, ?)`,
			entry.key,
			entry.valueJSON,
			revision,
		); err != nil {
			return mapSQLiteError(err, "migrate")
		}
	}
	if _, err := transaction.ExecContext(
		ctx,
		`UPDATE store_metadata SET global_revision = ? WHERE singleton = 1`,
		int64(len(legacy)),
	); err != nil {
		return mapSQLiteError(err, "migrate")
	}
	if _, err := transaction.ExecContext(ctx, `DROP TABLE entries_v0_migration`); err != nil {
		return mapSQLiteError(err, "migrate")
	}
	return nil
}

func validateV1(ctx context.Context, transaction *immediateTransaction) error {
	rows, err := transaction.QueryContext(
		ctx,
		`SELECT singleton, schema_version, global_revision FROM store_metadata`,
	)
	if err != nil {
		return malformedSchema()
	}
	type metadataRow struct {
		singleton, version, revision int64
	}
	var metadata []metadataRow
	for rows.Next() {
		var row metadataRow
		if err := rows.Scan(&row.singleton, &row.version, &row.revision); err != nil {
			rows.Close()
			return malformedSchema()
		}
		metadata = append(metadata, row)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return malformedSchema()
	}
	if err := rows.Close(); err != nil {
		return malformedSchema()
	}
	if len(metadata) != 1 || metadata[0].singleton != 1 || metadata[0].version != 1 {
		return malformedSchema()
	}
	globalRevision := metadata[0].revision
	if globalRevision < 0 || globalRevision > int64(domain.MaxRevision) {
		return revisionInvariant()
	}

	entryRows, err := transaction.QueryContext(
		ctx,
		`SELECT key, value_json, revision FROM entries ORDER BY key COLLATE BINARY`,
	)
	if err != nil {
		return malformedSchema()
	}
	defer entryRows.Close()
	revisions := make(map[int64]struct{})
	for entryRows.Next() {
		var key, valueJSON string
		var revision int64
		if err := entryRows.Scan(&key, &valueJSON, &revision); err != nil {
			return malformedSchema()
		}
		if _, err := domain.ParseKey(key); err != nil {
			return &domain.Error{
				Category: "invalid_storage",
				Details:  map[string]any{"reason": "invalid_key", "key": key},
				Cause:    err,
			}
		}
		if _, err := domain.ParseStoredValue(valueJSON); err != nil {
			return invalidStoredValue(key, err)
		}
		if revision < 1 || revision > globalRevision {
			return revisionInvariant()
		}
		if _, duplicate := revisions[revision]; duplicate {
			return revisionInvariant()
		}
		revisions[revision] = struct{}{}
	}
	if err := entryRows.Err(); err != nil {
		return malformedSchema()
	}
	return nil
}

func queryEntryRevision(
	ctx context.Context,
	transaction *immediateTransaction,
	key string,
	operation string,
) (*int64, error) {
	var revision int64
	err := transaction.QueryRowContext(
		ctx,
		`SELECT revision FROM entries WHERE key = ?`,
		key,
	).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, mapSQLiteError(err, operation)
	}
	if revision < 1 || revision > int64(domain.MaxRevision) {
		return nil, revisionInvariant()
	}
	return &revision, nil
}

func nextRevision(ctx context.Context, transaction *immediateTransaction) (int64, error) {
	var current int64
	if err := transaction.QueryRowContext(
		ctx,
		`SELECT global_revision FROM store_metadata WHERE singleton = 1`,
	).Scan(&current); err != nil {
		return 0, mapSQLiteError(err, "write")
	}
	if current == int64(domain.MaxRevision) {
		return 0, revisionExhausted()
	}
	if current < 0 || current > int64(domain.MaxRevision) {
		return 0, revisionInvariant()
	}
	return current + 1, nil
}

func updateGlobalRevision(
	ctx context.Context,
	transaction *immediateTransaction,
	revision int64,
) error {
	result, err := transaction.ExecContext(
		ctx,
		`UPDATE store_metadata SET global_revision = ? WHERE singleton = 1`,
		revision,
	)
	if err != nil {
		return mapSQLiteError(err, "write")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return mapSQLiteError(err, "write")
	}
	if affected != 1 {
		return malformedSchema()
	}
	return nil
}

func mapSQLiteError(err error, operation string) error {
	if err == nil {
		return nil
	}
	var sqliteError *sqlite.Error
	if errors.As(err, &sqliteError) {
		switch sqliteError.Code() & 0xff {
		case sqlite3.SQLITE_BUSY, sqlite3.SQLITE_LOCKED:
			return &domain.Error{
				Category: "busy",
				Details:  map[string]any{"timeout_ms": int64(busyTimeoutMS)},
				Cause:    err,
			}
		case sqlite3.SQLITE_CORRUPT, sqlite3.SQLITE_NOTADB:
			return integrityFailure(err)
		}
	}
	return storageError(operation, err)
}

func isBusyError(err error) bool {
	var sqliteError *sqlite.Error
	if !errors.As(err, &sqliteError) {
		return false
	}
	code := sqliteError.Code() & 0xff
	return code == sqlite3.SQLITE_BUSY || code == sqlite3.SQLITE_LOCKED
}

func storageError(operation string, cause error) error {
	return &domain.Error{
		Category: "storage_error",
		Details: map[string]any{
			"operation": operation,
			"reason":    "storage_failure",
		},
		Cause: cause,
	}
}

func malformedSchema() error {
	return &domain.Error{
		Category: "invalid_storage",
		Details:  map[string]any{"reason": "malformed_schema"},
	}
}

func revisionInvariant() error {
	return &domain.Error{
		Category: "invalid_storage",
		Details:  map[string]any{"reason": "revision_invariant"},
	}
}

func integrityFailure(cause error) error {
	return &domain.Error{
		Category: "invalid_storage",
		Details:  map[string]any{"reason": "integrity_check_failed"},
		Cause:    cause,
	}
}

func invalidStoredValue(key string, cause error) error {
	return &domain.Error{
		Category: "invalid_storage",
		Details:  map[string]any{"reason": "invalid_value", "key": key},
		Cause:    cause,
	}
}

func revisionExhausted() error {
	return &domain.Error{
		Category: "revision_exhausted",
		Details:  map[string]any{"maximum": domain.MaxRevision},
	}
}

func notFound(key string) error {
	return &domain.Error{
		Category: "not_found",
		Details:  map[string]any{"key": key},
	}
}

func conflict(key string, expected, actual any) error {
	return &domain.Error{
		Category: "conflict",
		Details: map[string]any{
			"key":      key,
			"expected": expected,
			"actual":   actual,
		},
	}
}

// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: query.sql

package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const acquireTXLock = `-- name: AcquireTXLock :exec
SELECT pg_advisory_xact_lock($1::bigint)
`

func (q *Queries) AcquireTXLock(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, acquireTXLock, id)
	return err
}

const checkPermission = `-- name: CheckPermission :one
SELECT (acl.uri IS NOT NULL) = true AS has_access
FROM document AS d
     LEFT JOIN acl
          ON acl.uuid = d.uuid AND acl.uri = ANY($1::text[])
          AND $2::text = ANY(permissions)
WHERE d.uuid = $3
`

type CheckPermissionParams struct {
	Uri        pgtype.Array[string]
	Permission string
	Uuid       uuid.UUID
}

func (q *Queries) CheckPermission(ctx context.Context, arg CheckPermissionParams) (bool, error) {
	row := q.db.QueryRow(ctx, checkPermission, arg.Uri, arg.Permission, arg.Uuid)
	var has_access bool
	err := row.Scan(&has_access)
	return has_access, err
}

const createStatus = `-- name: CreateStatus :exec
SELECT create_status(
       $1::uuid, $2::varchar(32), $3::bigint, $4::bigint,
       $5::timestamptz, $6::text, $7::jsonb
)
`

type CreateStatusParams struct {
	Uuid       uuid.UUID
	Name       string
	ID         int64
	Version    int64
	Created    pgtype.Timestamptz
	CreatorUri string
	Meta       []byte
}

func (q *Queries) CreateStatus(ctx context.Context, arg CreateStatusParams) error {
	_, err := q.db.Exec(ctx, createStatus,
		arg.Uuid,
		arg.Name,
		arg.ID,
		arg.Version,
		arg.Created,
		arg.CreatorUri,
		arg.Meta,
	)
	return err
}

const createVersion = `-- name: CreateVersion :exec
SELECT create_version(
       $1::uuid, $2::bigint, $3::timestamptz,
       $4::text, $5::jsonb, $6::jsonb
)
`

type CreateVersionParams struct {
	Uuid         uuid.UUID
	Version      int64
	Created      pgtype.Timestamptz
	CreatorUri   string
	Meta         []byte
	DocumentData []byte
}

func (q *Queries) CreateVersion(ctx context.Context, arg CreateVersionParams) error {
	_, err := q.db.Exec(ctx, createVersion,
		arg.Uuid,
		arg.Version,
		arg.Created,
		arg.CreatorUri,
		arg.Meta,
		arg.DocumentData,
	)
	return err
}

const deleteDocument = `-- name: DeleteDocument :exec
SELECT delete_document(
       $1::uuid, $2::text, $3::bigint
)
`

type DeleteDocumentParams struct {
	Uuid     uuid.UUID
	Uri      string
	RecordID int64
}

func (q *Queries) DeleteDocument(ctx context.Context, arg DeleteDocumentParams) error {
	_, err := q.db.Exec(ctx, deleteDocument, arg.Uuid, arg.Uri, arg.RecordID)
	return err
}

const dropACL = `-- name: DropACL :exec
DELETE FROM acl WHERE uuid = $1 AND uri = $2
`

type DropACLParams struct {
	Uuid uuid.UUID
	Uri  string
}

func (q *Queries) DropACL(ctx context.Context, arg DropACLParams) error {
	_, err := q.db.Exec(ctx, dropACL, arg.Uuid, arg.Uri)
	return err
}

const finaliseDelete = `-- name: FinaliseDelete :execrows
DELETE FROM document
WHERE uuid = $1 AND deleting = true
`

func (q *Queries) FinaliseDelete(ctx context.Context, uuid uuid.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, finaliseDelete, uuid)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getDocumentACL = `-- name: GetDocumentACL :many
SELECT uuid, uri, permissions FROM acl WHERE uuid = $1
`

func (q *Queries) GetDocumentACL(ctx context.Context, uuid uuid.UUID) ([]Acl, error) {
	rows, err := q.db.Query(ctx, getDocumentACL, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Acl
	for rows.Next() {
		var i Acl
		if err := rows.Scan(&i.Uuid, &i.Uri, &i.Permissions); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDocumentData = `-- name: GetDocumentData :one
SELECT v.document_data
FROM document as d
     INNER JOIN document_version AS v ON
           v.uuid = d.uuid And v.version = d.current_version
WHERE d.uuid = $1
`

func (q *Queries) GetDocumentData(ctx context.Context, uuid uuid.UUID) ([]byte, error) {
	row := q.db.QueryRow(ctx, getDocumentData, uuid)
	var document_data []byte
	err := row.Scan(&document_data)
	return document_data, err
}

const getDocumentForDeletion = `-- name: GetDocumentForDeletion :one
SELECT uuid, current_version AS delete_record_id FROM document
WHERE deleting = true
ORDER BY created
FOR UPDATE SKIP LOCKED
LIMIT 1
`

type GetDocumentForDeletionRow struct {
	Uuid           uuid.UUID
	DeleteRecordID int64
}

func (q *Queries) GetDocumentForDeletion(ctx context.Context) (GetDocumentForDeletionRow, error) {
	row := q.db.QueryRow(ctx, getDocumentForDeletion)
	var i GetDocumentForDeletionRow
	err := row.Scan(&i.Uuid, &i.DeleteRecordID)
	return i, err
}

const getDocumentForUpdate = `-- name: GetDocumentForUpdate :one
SELECT uri, current_version, deleting FROM document
WHERE uuid = $1
FOR UPDATE
`

type GetDocumentForUpdateRow struct {
	Uri            string
	CurrentVersion int64
	Deleting       bool
}

func (q *Queries) GetDocumentForUpdate(ctx context.Context, uuid uuid.UUID) (GetDocumentForUpdateRow, error) {
	row := q.db.QueryRow(ctx, getDocumentForUpdate, uuid)
	var i GetDocumentForUpdateRow
	err := row.Scan(&i.Uri, &i.CurrentVersion, &i.Deleting)
	return i, err
}

const getDocumentHeads = `-- name: GetDocumentHeads :many
SELECT name, id
FROM status_heads
WHERE uuid = $1
`

type GetDocumentHeadsRow struct {
	Name string
	ID   int64
}

func (q *Queries) GetDocumentHeads(ctx context.Context, uuid uuid.UUID) ([]GetDocumentHeadsRow, error) {
	rows, err := q.db.Query(ctx, getDocumentHeads, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetDocumentHeadsRow
	for rows.Next() {
		var i GetDocumentHeadsRow
		if err := rows.Scan(&i.Name, &i.ID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getDocumentInfo = `-- name: GetDocumentInfo :one
SELECT
        uuid, uri, created, creator_uri, updated, updater_uri, current_version,
        deleting
FROM document
WHERE uuid = $1
`

func (q *Queries) GetDocumentInfo(ctx context.Context, uuid uuid.UUID) (Document, error) {
	row := q.db.QueryRow(ctx, getDocumentInfo, uuid)
	var i Document
	err := row.Scan(
		&i.Uuid,
		&i.Uri,
		&i.Created,
		&i.CreatorUri,
		&i.Updated,
		&i.UpdaterUri,
		&i.CurrentVersion,
		&i.Deleting,
	)
	return i, err
}

const getDocumentStatusForArchiving = `-- name: GetDocumentStatusForArchiving :one
SELECT
        s.uuid, s.name, s.id, s.version, s.created, s.creator_uri, s.meta,
        p.signature AS parent_signature, v.signature AS version_signature
FROM document_status AS s
     INNER JOIN document_version AS v
           ON v.uuid = s.uuid
              AND v.version = s.version
              AND v.signature IS NOT NULL
     LEFT JOIN document_status AS p
          ON p.uuid = s.uuid AND p.name = s.name AND p.id = s.id-1
WHERE s.archived = false
AND (s.id = 1 OR p.archived = true)
ORDER BY s.created
FOR UPDATE OF s SKIP LOCKED
LIMIT 1
`

type GetDocumentStatusForArchivingRow struct {
	Uuid             uuid.UUID
	Name             string
	ID               int64
	Version          int64
	Created          pgtype.Timestamptz
	CreatorUri       string
	Meta             []byte
	ParentSignature  pgtype.Text
	VersionSignature pgtype.Text
}

func (q *Queries) GetDocumentStatusForArchiving(ctx context.Context) (GetDocumentStatusForArchivingRow, error) {
	row := q.db.QueryRow(ctx, getDocumentStatusForArchiving)
	var i GetDocumentStatusForArchivingRow
	err := row.Scan(
		&i.Uuid,
		&i.Name,
		&i.ID,
		&i.Version,
		&i.Created,
		&i.CreatorUri,
		&i.Meta,
		&i.ParentSignature,
		&i.VersionSignature,
	)
	return i, err
}

const getDocumentUnarchivedCount = `-- name: GetDocumentUnarchivedCount :one
SELECT SUM(num) FROM (
       SELECT COUNT(*) as num
              FROM document_status AS s
              WHERE s.uuid = $1 AND s.archived = false
       UNION
       SELECT COUNT(*) as num
              FROM document_version AS v
              WHERE v.uuid = $1 AND v.archived = false
) AS unarchived
`

func (q *Queries) GetDocumentUnarchivedCount(ctx context.Context, uuid uuid.UUID) (int64, error) {
	row := q.db.QueryRow(ctx, getDocumentUnarchivedCount, uuid)
	var sum int64
	err := row.Scan(&sum)
	return sum, err
}

const getDocumentVersionData = `-- name: GetDocumentVersionData :one
SELECT document_data
FROM document_version
WHERE uuid = $1 AND version = $2
`

type GetDocumentVersionDataParams struct {
	Uuid    uuid.UUID
	Version int64
}

func (q *Queries) GetDocumentVersionData(ctx context.Context, arg GetDocumentVersionDataParams) ([]byte, error) {
	row := q.db.QueryRow(ctx, getDocumentVersionData, arg.Uuid, arg.Version)
	var document_data []byte
	err := row.Scan(&document_data)
	return document_data, err
}

const getDocumentVersionForArchiving = `-- name: GetDocumentVersionForArchiving :one
SELECT
        v.uuid, v.version, v.created, v.creator_uri, v.meta, v.document_data,
        p.signature AS parent_signature
FROM document_version AS v
     LEFT JOIN document_version AS p
          ON p.uuid = v.uuid AND p.version = v.version-1
WHERE v.archived = false
AND (v.version = 1 OR p.archived = true)
ORDER BY v.created
FOR UPDATE OF v SKIP LOCKED
LIMIT 1
`

type GetDocumentVersionForArchivingRow struct {
	Uuid            uuid.UUID
	Version         int64
	Created         pgtype.Timestamptz
	CreatorUri      string
	Meta            []byte
	DocumentData    []byte
	ParentSignature pgtype.Text
}

func (q *Queries) GetDocumentVersionForArchiving(ctx context.Context) (GetDocumentVersionForArchivingRow, error) {
	row := q.db.QueryRow(ctx, getDocumentVersionForArchiving)
	var i GetDocumentVersionForArchivingRow
	err := row.Scan(
		&i.Uuid,
		&i.Version,
		&i.Created,
		&i.CreatorUri,
		&i.Meta,
		&i.DocumentData,
		&i.ParentSignature,
	)
	return i, err
}

const getFullDocumentHeads = `-- name: GetFullDocumentHeads :many
SELECT s.uuid, s.name, s.id, s.version, s.created, s.creator_uri, s.meta,
       s.archived, s.signature
FROM status_heads AS h
     INNER JOIN document_status AS s ON
           s.uuid = h.uuid AND s.name = h.name AND s.id = h.id
WHERE h.uuid = $1
`

func (q *Queries) GetFullDocumentHeads(ctx context.Context, uuid uuid.UUID) ([]DocumentStatus, error) {
	rows, err := q.db.Query(ctx, getFullDocumentHeads, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DocumentStatus
	for rows.Next() {
		var i DocumentStatus
		if err := rows.Scan(
			&i.Uuid,
			&i.Name,
			&i.ID,
			&i.Version,
			&i.Created,
			&i.CreatorUri,
			&i.Meta,
			&i.Archived,
			&i.Signature,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getSigningKeys = `-- name: GetSigningKeys :many
SELECT kid, spec FROM signing_keys
`

func (q *Queries) GetSigningKeys(ctx context.Context) ([]SigningKey, error) {
	rows, err := q.db.Query(ctx, getSigningKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SigningKey
	for rows.Next() {
		var i SigningKey
		if err := rows.Scan(&i.Kid, &i.Spec); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getStatuses = `-- name: GetStatuses :many
SELECT uuid, name, id, version, created, creator_uri, meta
FROM document_status
WHERE uuid = $1 AND name = $2 AND ($3 = 0 OR id < $3)
ORDER BY id DESC
LIMIT $4
`

type GetStatusesParams struct {
	Uuid    uuid.UUID
	Name    string
	Column3 interface{}
	Limit   int32
}

type GetStatusesRow struct {
	Uuid       uuid.UUID
	Name       string
	ID         int64
	Version    int64
	Created    pgtype.Timestamptz
	CreatorUri string
	Meta       []byte
}

func (q *Queries) GetStatuses(ctx context.Context, arg GetStatusesParams) ([]GetStatusesRow, error) {
	rows, err := q.db.Query(ctx, getStatuses,
		arg.Uuid,
		arg.Name,
		arg.Column3,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStatusesRow
	for rows.Next() {
		var i GetStatusesRow
		if err := rows.Scan(
			&i.Uuid,
			&i.Name,
			&i.ID,
			&i.Version,
			&i.Created,
			&i.CreatorUri,
			&i.Meta,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getVersion = `-- name: GetVersion :one
SELECT created, creator_uri, meta, archived
FROM document_version
WHERE uuid = $1 AND version = $2
`

type GetVersionParams struct {
	Uuid    uuid.UUID
	Version int64
}

type GetVersionRow struct {
	Created    pgtype.Timestamptz
	CreatorUri string
	Meta       []byte
	Archived   bool
}

func (q *Queries) GetVersion(ctx context.Context, arg GetVersionParams) (GetVersionRow, error) {
	row := q.db.QueryRow(ctx, getVersion, arg.Uuid, arg.Version)
	var i GetVersionRow
	err := row.Scan(
		&i.Created,
		&i.CreatorUri,
		&i.Meta,
		&i.Archived,
	)
	return i, err
}

const getVersions = `-- name: GetVersions :many
SELECT version, created, creator_uri, meta, archived
FROM document_version
WHERE uuid = $1 AND ($2::bigint = 0 OR version < $2::bigint)
ORDER BY version DESC
LIMIT $3
`

type GetVersionsParams struct {
	Uuid   uuid.UUID
	Before int64
	Count  int32
}

type GetVersionsRow struct {
	Version    int64
	Created    pgtype.Timestamptz
	CreatorUri string
	Meta       []byte
	Archived   bool
}

func (q *Queries) GetVersions(ctx context.Context, arg GetVersionsParams) ([]GetVersionsRow, error) {
	rows, err := q.db.Query(ctx, getVersions, arg.Uuid, arg.Before, arg.Count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetVersionsRow
	for rows.Next() {
		var i GetVersionsRow
		if err := rows.Scan(
			&i.Version,
			&i.Created,
			&i.CreatorUri,
			&i.Meta,
			&i.Archived,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const granteesWithPermission = `-- name: GranteesWithPermission :many
SELECT uri
FROM acl
WHERE uuid = $1
      AND $2::text = ANY(permissions)
`

type GranteesWithPermissionParams struct {
	Uuid       uuid.UUID
	Permission string
}

func (q *Queries) GranteesWithPermission(ctx context.Context, arg GranteesWithPermissionParams) ([]string, error) {
	rows, err := q.db.Query(ctx, granteesWithPermission, arg.Uuid, arg.Permission)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		items = append(items, uri)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertACLAuditEntry = `-- name: InsertACLAuditEntry :exec
INSERT INTO acl_audit(uuid, updated, updater_uri, state)
SELECT $1::uuid, $2::timestamptz, $3::text, json_agg(l)
FROM (
       SELECT uri, permissions
       FROM acl
       WHERE uuid = $1::uuid
) AS l
`

type InsertACLAuditEntryParams struct {
	Uuid       uuid.UUID
	Updated    pgtype.Timestamptz
	UpdaterUri string
}

func (q *Queries) InsertACLAuditEntry(ctx context.Context, arg InsertACLAuditEntryParams) error {
	_, err := q.db.Exec(ctx, insertACLAuditEntry, arg.Uuid, arg.Updated, arg.UpdaterUri)
	return err
}

const insertDeleteRecord = `-- name: InsertDeleteRecord :one
INSERT INTO delete_record(
       uuid, uri, version, created, creator_uri, meta
) VALUES(
       $1, $2, $3, $4, $5, $6
) RETURNING id
`

type InsertDeleteRecordParams struct {
	Uuid       uuid.UUID
	Uri        string
	Version    int64
	Created    pgtype.Timestamptz
	CreatorUri string
	Meta       []byte
}

func (q *Queries) InsertDeleteRecord(ctx context.Context, arg InsertDeleteRecordParams) (int64, error) {
	row := q.db.QueryRow(ctx, insertDeleteRecord,
		arg.Uuid,
		arg.Uri,
		arg.Version,
		arg.Created,
		arg.CreatorUri,
		arg.Meta,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertSigningKey = `-- name: InsertSigningKey :exec
INSERT INTO signing_keys(kid, spec) VALUES($1, $2)
`

type InsertSigningKeyParams struct {
	Kid  string
	Spec []byte
}

func (q *Queries) InsertSigningKey(ctx context.Context, arg InsertSigningKeyParams) error {
	_, err := q.db.Exec(ctx, insertSigningKey, arg.Kid, arg.Spec)
	return err
}

const notify = `-- name: Notify :exec
SELECT pg_notify($1::text, $2::text)
`

type NotifyParams struct {
	Channel string
	Message string
}

func (q *Queries) Notify(ctx context.Context, arg NotifyParams) error {
	_, err := q.db.Exec(ctx, notify, arg.Channel, arg.Message)
	return err
}

const setDocumentStatusAsArchived = `-- name: SetDocumentStatusAsArchived :exec
UPDATE document_status
SET archived = true, signature = $1::text
WHERE uuid = $2 AND id = $3
`

type SetDocumentStatusAsArchivedParams struct {
	Signature string
	Uuid      uuid.UUID
	ID        int64
}

func (q *Queries) SetDocumentStatusAsArchived(ctx context.Context, arg SetDocumentStatusAsArchivedParams) error {
	_, err := q.db.Exec(ctx, setDocumentStatusAsArchived, arg.Signature, arg.Uuid, arg.ID)
	return err
}

const setDocumentVersionAsArchived = `-- name: SetDocumentVersionAsArchived :exec
UPDATE document_version
SET archived = true, signature = $1::text
WHERE uuid = $2 AND version = $3
`

type SetDocumentVersionAsArchivedParams struct {
	Signature string
	Uuid      uuid.UUID
	Version   int64
}

func (q *Queries) SetDocumentVersionAsArchived(ctx context.Context, arg SetDocumentVersionAsArchivedParams) error {
	_, err := q.db.Exec(ctx, setDocumentVersionAsArchived, arg.Signature, arg.Uuid, arg.Version)
	return err
}

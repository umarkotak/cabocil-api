package book_repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"b.id",
		"b.created_at",
		"b.updated_at",
		"b.deleted_at",
		"b.slug",
		"b.title",
		"b.description",
		"b.cover_file_guid",
		"b.tags",
		"b.type",
		"b.pdf_file_guid",
		"b.active",
		"b.original_pdf_url",
		"b.access_tags",
		"b.storage",
	}, ", ")

	queryGetByID = fmt.Sprintf(`
		SELECT
			%s,
			COALESCE(fb.exact_path, '') AS cover_file_path
		FROM books b
		LEFT JOIN file_bucket fb ON fb.guid = b.cover_file_guid
		WHERE
			b.id = :id
			AND b.deleted_at IS NULL
	`, allColumns)

	queryGetBySlug = fmt.Sprintf(`
		SELECT
			%s,
			COALESCE(fb.exact_path, '') AS cover_file_path
		FROM books b
		LEFT JOIN file_bucket fb ON fb.guid = b.cover_file_guid
		WHERE
			b.slug = :slug
			AND b.deleted_at IS NULL
	`, allColumns)

	queryGetByParams = fmt.Sprintf(`
		SELECT
			%s,
			COALESCE(fb.exact_path, '') AS cover_file_path
		FROM books b
		LEFT JOIN file_bucket fb ON fb.guid = b.cover_file_guid
		WHERE
			1 = 1
			AND (:title = '' OR b.title ILIKE :title)
			AND (:tags = '{}' OR b.tags && :tags)
			AND (:types = '{}' OR b.type = ANY(:types))
			AND (
				CASE
					WHEN :active = 'true' THEN b.active
					WHEN :active = 'false' THEN NOT b.active
					ELSE TRUE
				END
			)
			AND b.deleted_at IS NULL
			AND (:exclude_ids = '{}' OR NOT(b.id = ANY(:exclude_ids)))
			AND (:access = '{}' OR b.access_tags && :access)
			AND (:exclude_access = '{}' OR NOT(b.access_tags && :exclude_access))
		ORDER BY
			CASE WHEN :sort = 'title_asc' THEN b.title END ASC,
			CASE WHEN :sort = 'title_desc' THEN b.title END DESC,
			CASE WHEN :sort = 'id_asc' THEN b.id END ASC,
			CASE WHEN :sort = 'id_desc' THEN b.id END DESC,
			CASE WHEN :sort = 'random' THEN RANDOM() END
		LIMIT :limit OFFSET :offset
	`, allColumns)

	queryInsert = `
		INSERT INTO books (
			slug,
			title,
			description,
			cover_file_guid,
			tags,
			type,
			pdf_file_guid,
			original_pdf_url,
			access_tags,
			storage
		) VALUES (
			:slug,
			:title,
			:description,
			:cover_file_guid,
			:tags,
			:type,
			:pdf_file_guid,
			:original_pdf_url,
			:access_tags,
			:storage
		) RETURNING id
	`

	queryUpdate = `
		UPDATE books
		SET
			slug = :slug,
			title = :title,
			description = :description,
			cover_file_guid = :cover_file_guid,
			tags = :tags,
			type = :type,
			pdf_file_guid = :pdf_file_guid,
			original_pdf_url = :original_pdf_url,
			access_tags = :access_tags,
			storage = :storage,
			active = :active
		WHERE
			id = :id
	`

	querySoftDelete = `
		UPDATE books
		SET
			deleted_at = NOW()
		WHERE
			id = :id
	`

	queryDelete = `
		DELETE FROM books
		WHERE
			id = :id
	`

	queryGetTags = `
		SELECT ARRAY_AGG(DISTINCT t.tag) AS unique_tags
		FROM (
			SELECT UNNEST(tags) AS tag
			FROM books
		) t;
	`
)

var (
	stmtGetByID     *sqlx.NamedStmt
	stmtGetBySlug   *sqlx.NamedStmt
	stmtGetByParams *sqlx.NamedStmt
	stmtInsert      *sqlx.NamedStmt
	stmtUpdate      *sqlx.NamedStmt
	stmtSoftDelete  *sqlx.NamedStmt
	stmtDelete      *sqlx.NamedStmt
	stmtGetTags     *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetByID, err = datastore.Get().Db.PrepareNamed(queryGetByID)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetBySlug, err = datastore.Get().Db.PrepareNamed(queryGetBySlug)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetByParams, err = datastore.Get().Db.PrepareNamed(queryGetByParams)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtInsert, err = datastore.Get().Db.PrepareNamed(queryInsert)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtUpdate, err = datastore.Get().Db.PrepareNamed(queryUpdate)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtSoftDelete, err = datastore.Get().Db.PrepareNamed(querySoftDelete)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtDelete, err = datastore.Get().Db.PrepareNamed(queryDelete)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetTags, err = datastore.Get().Db.PrepareNamed(queryGetTags)
	if err != nil {
		logrus.Fatal(err)
	}
}

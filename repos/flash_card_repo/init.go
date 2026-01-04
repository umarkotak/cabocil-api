package flash_card_repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"fc.id",
		"fc.created_at",
		"fc.updated_at",
		"fc.deleted_at",
		"fc.slug",
		"fc.name_id",
		"fc.name_en",
		"fc.picture_url",
		"fc.tags",
		"fc.metadata",
	}, ", ")

	queryGetByID = fmt.Sprintf(`
		SELECT
			%s
		FROM flash_cards fc
		WHERE
			1 = 1
			AND fc.id = :id
			AND fc.deleted_at IS NULL
	`, allColumns)

	queryGetByTags = fmt.Sprintf(`
		SELECT
			%s
		FROM flash_cards fc
		WHERE
			1 = 1
			AND fc.tags && :tags
			AND fc.deleted_at IS NULL
		ORDER BY fc.id ASC
		LIMIT :limit OFFSET :offset
	`, allColumns)

	queryBulkInsert = `
		INSERT INTO flash_cards (
			slug,
			name_id,
			name_en,
			picture_url,
			tags,
			metadata
		) VALUES (
			:slug,
			:name_id,
			:name_en,
			:picture_url,
			:tags,
			:metadata
		) RETURNING id
	`
)

var (
	stmtGetByID    *sqlx.NamedStmt
	stmtGetByTags  *sqlx.NamedStmt
	stmtBulkInsert *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetByID, err = datastore.Get().Db.PrepareNamed(queryGetByID)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetByTags, err = datastore.Get().Db.PrepareNamed(queryGetByTags)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtBulkInsert, err = datastore.Get().Db.PrepareNamed(queryBulkInsert)
	if err != nil {
		logrus.Fatal(err)
	}
}

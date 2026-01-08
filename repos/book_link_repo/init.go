package book_link_repo

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"bl.id",
		"bl.created_at",
		"bl.updated_at",
		"bl.deleted_at",
		"bl.group_name",
		"bl.name",
		"bl.url",
		"bl.image_url",
		"bl.premium",
		"bl.metadata",
	}, ", ")

	queryGetAll = `
		SELECT
			` + allColumns + `
		FROM book_links bl
		WHERE
			bl.deleted_at IS NULL
		ORDER BY bl.group_name ASC, bl.name ASC
	`
)

var (
	stmtGetAll *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetAll, err = datastore.Get().Db.PrepareNamed(queryGetAll)
	if err != nil {
		logrus.Fatal(err)
	}
}

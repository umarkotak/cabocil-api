package youtube_channel_repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"ytch.id",
		"ytch.created_at",
		"ytch.updated_at",
		"ytch.deleted_at",
		"ytch.external_id",
		"ytch.name",
		"ytch.username",
		"ytch.image_url",
		"ytch.tags",
		"ytch.active",
		"ytch.channel_link",
	}, ", ")

	queryGetByID = fmt.Sprintf(`
		SELECT
			%s
		FROM youtube_channels ytch
		WHERE
			ytch.id = :id
			AND ytch.deleted_at IS NULL
	`, allColumns)

	queryGetByExternalID = fmt.Sprintf(`
		SELECT
			%s
		FROM youtube_channels ytch
		WHERE
			ytch.external_id = :external_id
			AND ytch.deleted_at IS NULL
	`, allColumns)

	queryGetForSearch = fmt.Sprintf(`
		SELECT
			%s
		FROM youtube_channels ytch
		WHERE
			1 = 1
			AND (:name = '' OR ytch.name = :name)
			AND (:tags = '{}' OR ytch.tags @> :tags)
			AND (
				CASE
					WHEN :active = 'true' THEN ytch.active
					WHEN :active = 'false' THEN NOT ytch.active
					ELSE TRUE
				END
			)
			AND ytch.deleted_at IS NULL
		ORDER BY ytch.name ASC
	`, allColumns)

	queryInsert = `
		INSERT INTO youtube_channels (
			external_id,
			name,
			username,
			image_url,
			tags,
			channel_link
		) VALUES (
			:external_id,
			:name,
			:username,
			:image_url,
			:tags,
			:channel_link
		) RETURNING id
	`

	queryUpdate = `
		UPDATE youtube_channels
		SET
			name = :name,
			username = :username,
			image_url = :image_url,
			tags = :tags,
			active = :active,
			channel_link = :channel_link,
			updated_at = NOW()
		WHERE
			id = :id
	`

	queryUpdateActive = `
		UPDATE youtube_channels
		SET
			active = :active
		WHERE
			id = :id
	`

	querySoftDelete = `
		UPDATE youtube_channels
		SET
			deleted_at = NOW()
		WHERE
			id = :id
	`

	queryDelete = `
		DELETE FROM youtube_channels
		WHERE
			id = :id
	`
)

var (
	stmtGetByID         *sqlx.NamedStmt
	stmtGetByExternalID *sqlx.NamedStmt
	stmtGetForSearch    *sqlx.NamedStmt
	stmtInsert          *sqlx.NamedStmt
	stmtUpdate          *sqlx.NamedStmt
	stmtUpdateActive    *sqlx.NamedStmt
	stmtSoftDelete      *sqlx.NamedStmt
	stmtDelete          *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetByID, err = datastore.Get().Db.PrepareNamed(queryGetByID)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetByExternalID, err = datastore.Get().Db.PrepareNamed(queryGetByExternalID)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetForSearch, err = datastore.Get().Db.PrepareNamed(queryGetForSearch)
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
	stmtUpdateActive, err = datastore.Get().Db.PrepareNamed(queryUpdateActive)
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
}

package user_repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"u.id",
		"u.created_at",
		"u.updated_at",
		"u.deleted_at",
		"u.guid",
		"u.email",
		"u.about",
		"u.password",
		"u.name",
		"u.username",
		"u.photo_url",
		"u.user_role",
	}, ", ")

	queryGetByEmail = fmt.Sprintf(`
		SELECT
			%s
		FROM users u
		WHERE
			u.email = :email
			AND u.deleted_at IS NULL
	`, allColumns)

	queryGetByParams = fmt.Sprintf(`
		SELECT
			%s
		FROM users u
		WHERE
			1 = 1
			AND (:guid = '' OR u.guid = :guid)
			AND (:email = '' OR u.email = :email)
			AND (:name = '' OR u.name = :name)
			AND (:username = '' OR u.username = :username)
			AND u.deleted_at IS NULL
	`, allColumns)

	queryGetByParamsWithSubscription = fmt.Sprintf(`
		SELECT
			%s,
			COALESCE(us.ended_at, CAST('1970-01-01' AS TIMESTAMP)) AS subscription_ended_at
		FROM users u
		LEFT JOIN (
			SELECT
				user_id,
				MAX(ended_at) AS ended_at
			FROM user_subscriptions
			WHERE deleted_at IS NULL
			GROUP BY user_id
		) us ON u.id = us.user_id
		WHERE
			1 = 1
			AND (:guid = '' OR u.guid = :guid)
			AND (:email = '' OR u.email = :email)
			AND (:name = '' OR u.name = :name)
			AND (:username = '' OR u.username = :username)
			AND u.deleted_at IS NULL
		ORDER BY u.id ASC
	`, allColumns)

	queryGetByID = fmt.Sprintf(`
		SELECT
			%s
		FROM users u
		WHERE
			u.id = :id
			AND u.deleted_at IS NULL
	`, allColumns)

	queryGetByGuid = fmt.Sprintf(`
		SELECT
			%s
		FROM users u
		WHERE
			u.guid = :guid
			AND u.deleted_at IS NULL
	`, allColumns)

	queryInsert = `
		INSERT INTO users (
			guid,
			email,
			about,
			password,
			name,
			photo_url,
			username
		) VALUES (
			:guid,
			:email,
			:about,
			:password,
			:name,
			:photo_url,
			:username
		) RETURNING id
	`

	queryUpdate = `
		UPDATE users
		SET
			guid = :guid,
			email = :email,
			about = :about,
			password = :password,
			name = :name,
			username = :username,
			photo_url = :photo_url,
			updated_at = NOW()
		WHERE
			id = :id
	`

	querySoftDelete = `
		UPDATE users
		SET
			email = guid,
			name = guid,
			username = guid,
			photo_url = guid,
			deleted_at = NOW()
		WHERE
			id = :id
	`
)

var (
	stmtGetByEmail                  *sqlx.NamedStmt
	stmtGetByParams                 *sqlx.NamedStmt
	stmtGetByParamsWithSubscription *sqlx.NamedStmt
	stmtGetByID                     *sqlx.NamedStmt
	stmtGetByGuid                   *sqlx.NamedStmt
	stmtInsert                      *sqlx.NamedStmt
	stmtUpdate                      *sqlx.NamedStmt
	stmtSoftDelete                  *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetByEmail, err = datastore.Get().Db.PrepareNamed(queryGetByEmail)
	if err != nil {
		logrus.Fatal(err)
	}

	stmtGetByParams, err = datastore.Get().Db.PrepareNamed(queryGetByParams)
	if err != nil {
		logrus.Fatal(err)
	}

	stmtGetByParamsWithSubscription, err = datastore.Get().Db.PrepareNamed(queryGetByParamsWithSubscription)
	if err != nil {
		logrus.Fatal(err)
	}

	stmtGetByID, err = datastore.Get().Db.PrepareNamed(queryGetByID)
	if err != nil {
		logrus.Fatal(err)
	}

	stmtGetByGuid, err = datastore.Get().Db.PrepareNamed(queryGetByGuid)
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
}

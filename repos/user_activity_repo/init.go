package user_activity_repo

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/datastore"
)

var (
	allColumns = strings.Join([]string{
		"ua.id",
		"ua.created_at",
		"ua.updated_at",
		"ua.deleted_at",
		"ua.user_id",
		"ua.app_session",
		"ua.youtube_video_id",
		"ua.book_id",
		"ua.book_content_id",
		"ua.metadata",
	}, ", ")

	queryGetByParams = fmt.Sprintf(`
		SELECT
			%s
		FROM user_activities ua
		WHERE
			1 = 1
			AND ua.user_id = :user_id
			AND ua.app_session = :app_session
			AND ua.deleted_at IS NULL
		ORDER BY ua.updated_at DESC
		LIMIT :limit OFFSET :offset
	`, allColumns)

	queryGetFullByParams = `
		SELECT
			ua.id,
			ua.created_at,
			ua.updated_at,
			ua.deleted_at,
			ua.user_id,
			ua.app_session,
			ua.youtube_video_id,
			ua.book_id,
			ua.book_content_id,
			ua.metadata,
			yv.title AS youtube_video_title,
			yv.image_url AS youtube_video_image_url,
			yc.name AS youtube_channel_name,
			yc.image_url AS youtube_channel_image_url,
			b.title AS book_title,
			b.cover_file_guid AS book_cover_file_guid,
			b.slug AS book_slug,
			fb.storage AS book_cover_storage,
			fb.exact_path AS book_cover_exact_path,
			b.type AS book_type,
			bc.page_number AS book_last_read_page_number
		FROM user_activities ua
		LEFT JOIN books b ON b.id = ua.book_id
		LEFT JOIN book_contents bc ON bc.id = CAST(ua.metadata->>'last_read_book_content_id' AS BIGINT)
		LEFT JOIN youtube_videos yv ON yv.id = ua.youtube_video_id
		LEFT JOIN youtube_channels yc ON yc.id = yv.youtube_channel_id
		LEFT JOIN file_bucket fb ON fb.guid = b.cover_file_guid
		WHERE
			1 = 1
			AND ua.user_id = :user_id
			AND ua.app_session = :app_session
			AND ua.deleted_at IS NULL
		ORDER BY ua.updated_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryGetByUserActivity = fmt.Sprintf(`
		SELECT
			%s
		FROM user_activities ua
		WHERE
			1 = 1
			AND ua.user_id = :user_id
			AND ua.app_session = :app_session
			AND ua.youtube_video_id = :youtube_video_id
			AND ua.book_id = :book_id
			AND ua.book_content_id = :book_content_id
			AND ua.deleted_at IS NULL
	`, allColumns)

	queryGetFullByUserActivity = `
		SELECT
			ua.id,
			ua.created_at,
			ua.updated_at,
			ua.deleted_at,
			ua.user_id,
			ua.app_session,
			ua.youtube_video_id,
			ua.book_id,
			ua.book_content_id,
			ua.metadata,
			ua.metadata,
			yv.title AS youtube_video_title,
			yv.image_url AS youtube_video_image_url,
			yc.name AS youtube_channel_name,
			yc.image_url AS youtube_channel_image_url,
			b.title AS book_title,
			b.cover_file_guid AS book_cover_file_guid,
			b.slug AS book_slug,
			fb.storage AS book_cover_storage,
			fb.exact_path AS book_cover_exact_path,
			b.type AS book_type,
			bc.page_number AS book_last_read_page_number
		FROM user_activities ua
		LEFT JOIN books b ON b.id = ua.book_id
		LEFT JOIN book_contents bc ON bc.id = CAST(ua.metadata->>'last_read_book_content_id' AS BIGINT)
		LEFT JOIN youtube_videos yv ON yv.id = ua.youtube_video_id
		LEFT JOIN youtube_channels yc ON yc.id = yv.youtube_channel_id
		LEFT JOIN file_bucket fb ON fb.guid = b.cover_file_guid
		WHERE
			1 = 1
			AND ua.user_id = :user_id
			AND ua.app_session = :app_session
			AND ua.youtube_video_id = :youtube_video_id
			AND ua.book_id = :book_id
			AND ua.book_content_id = :book_content_id
			AND ua.deleted_at IS NULL
	`

	queryInsert = `
		INSERT INTO user_activities (
			user_id,
			app_session,
			youtube_video_id,
			book_id,
			book_content_id,
			metadata
		) VALUES (
			:user_id,
			:app_session,
			:youtube_video_id,
			:book_id,
			:book_content_id,
			:metadata
		) RETURNING id
	`

	queryUpsert = `
		INSERT INTO user_activities (
			user_id,
			app_session,
			youtube_video_id,
			book_id,
			book_content_id,
			metadata
		) VALUES (
			:user_id,
			:app_session,
			:youtube_video_id,
			:book_id,
			:book_content_id,
			:metadata
		)
		ON CONFLICT (user_id, app_session, youtube_video_id, book_id, book_content_id)
		DO UPDATE SET
			updated_at = NOW(),
			metadata = :metadata
		RETURNING id
	`

	queryDeleteByYoutubeVideoIDs = `
		DELETE FROM user_activities
		WHERE
			youtube_video_id = ANY(:youtube_video_ids)
	`

	queryDeleteByBookIDs = `
		DELETE FROM user_activities
		WHERE
			book_id = ANY(:book_ids)
	`
)

var (
	stmtGetByParams             *sqlx.NamedStmt
	stmtGetFullByParams         *sqlx.NamedStmt
	stmtGetByUserActivity       *sqlx.NamedStmt
	stmtGetFullByUserActivity   *sqlx.NamedStmt
	stmtInsert                  *sqlx.NamedStmt
	stmtUpsert                  *sqlx.NamedStmt
	stmtDeleteByYoutubeVideoIDs *sqlx.NamedStmt
	stmtDeleteByBookIDs         *sqlx.NamedStmt
)

func Initialize() {
	var err error

	stmtGetByParams, err = datastore.Get().Db.PrepareNamed(queryGetByParams)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetFullByParams, err = datastore.Get().Db.PrepareNamed(queryGetFullByParams)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetByUserActivity, err = datastore.Get().Db.PrepareNamed(queryGetByUserActivity)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtGetFullByUserActivity, err = datastore.Get().Db.PrepareNamed(queryGetFullByUserActivity)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtInsert, err = datastore.Get().Db.PrepareNamed(queryInsert)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtUpsert, err = datastore.Get().Db.PrepareNamed(queryUpsert)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtDeleteByYoutubeVideoIDs, err = datastore.Get().Db.PrepareNamed(queryDeleteByYoutubeVideoIDs)
	if err != nil {
		logrus.Fatal(err)
	}
	stmtDeleteByBookIDs, err = datastore.Get().Db.PrepareNamed(queryDeleteByBookIDs)
	if err != nil {
		logrus.Fatal(err)
	}
}

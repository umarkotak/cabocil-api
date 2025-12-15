package contract

import (
	"github.com/lib/pq"
	"github.com/umarkotak/ytkidd-api/model"
)

type (
	YoutubeVideoSearch struct {
		Name   string         `db:"name"`
		Tags   pq.StringArray `db:"tags"`
		Active bool           `db:"active"`
	}

	GetYoutubeVideos struct {
		UserGuid          string         `db:"-"`
		UserRole          string         `db:"-"`
		Tags              pq.StringArray `db:"tags"`
		ChannelIDs        pq.Int64Array  `db:"channel_ids"`
		ExcludeIDs        pq.Int64Array  `db:"exclude_ids"`
		ExcludeChannelIDs pq.Int64Array  `db:"exclude_channel_ids"`
		Sort              string         `db:"sort"`
		Active            string         `db:"active"`
		model.Pagination
	}
)

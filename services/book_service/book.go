package book_service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_content_repo"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/repos/file_bucket_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_activity_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_stroke_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_subscription_repo"
	"github.com/umarkotak/ytkidd-api/utils"
	"github.com/umarkotak/ytkidd-api/utils/file_bucket"
	"github.com/umarkotak/ytkidd-api/utils/random"
)

func GetBooks(ctx context.Context, params contract.GetBooks) (contract_resp.GetBooks, error) {
	books, err := book_repo.GetByParams(ctx, params)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.GetBooks{}, err
	}

	bookDatas := []contract_resp.Book{}
	for _, book := range books {
		coverFileUrl := file_bucket.GenFinalUrl(ctx, book.Storage, book.CoverFilePath)

		bookData := contract_resp.Book{
			ID:           book.ID,
			Slug:         book.Slug,
			Title:        book.Title,
			Description:  book.Description,
			CoverFileUrl: file_bucket.GenCacheUrl(coverFileUrl, model.PRESET_COVER, 0, 0),
			Tags:         book.Tags,
			Type:         book.Type,
			IsFree:       book.IsFree(),
			Active:       book.Active,
			Storage:      book.Storage,
			ContentCount: book.ContentCount.Int64,
		}
		bookDatas = append(bookDatas, bookData)
	}

	tags, err := book_repo.GetTags(ctx)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
	}

	tagMap := map[string][]string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, "age:") {
			if _, found := tagMap["age"]; !found {
				tagMap["age"] = []string{tag}
			} else {
				tagMap["age"] = append(tagMap["age"], tag)
			}
		}

		// Add ass needed later
	}

	tagGroup := []contract_resp.TagGroup{
		{
			Name: "Umur",
			Tags: tagMap["age"],
		},
		{
			Name: "Semua Tag",
			Tags: tags,
		},
	}

	return contract_resp.GetBooks{
		TagGroup: tagGroup,
		Books:    bookDatas,
	}, nil
}

func GetBookDetail(ctx context.Context, params contract.GetBooks) (contract_resp.BookDetail, error) {
	var err error

	isSlug := utils.StringMustInt64(params.Slug) == 0

	var book model.Book
	if isSlug {
		book, err = book_repo.GetBySlug(ctx, params.Slug)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return contract_resp.BookDetail{}, err
		}
	} else {
		book, err = book_repo.GetByID(ctx, utils.StringMustInt64(params.Slug))
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return contract_resp.BookDetail{}, err
		}
	}

	user, err := user_repo.GetByGuid(ctx, params.UserGuid)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.BookDetail{}, err
	}

	isSubscribed := false
	if slices.Contains(model.ADMIN_ROLES, params.UserRole) {
		isSubscribed = true
	} else {
		subs, err := user_subscription_repo.GetActiveByUserID(ctx, user.ID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
		}
		isSubscribed = len(subs) > 0
	}

	bookContents, err := book_content_repo.GetByBookID(ctx, book.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.BookDetail{}, err
	}

	bookContentDatas := make([]contract_resp.BookContent, 0, len(bookContents))
	lastIndex := len(bookContents)
	if !book.IsFree() && !isSubscribed {
		lastIndex = model.ALLOWED_PREMIUM_FREE_PAGE_COUNT
	}
	for _, bookContent := range bookContents[0:lastIndex] {
		imageFileUrl := file_bucket.GenFinalUrl(ctx, book.Storage, bookContent.ImageFilePath)

		bookContentData := contract_resp.BookContent{
			ID:           bookContent.ID,
			BookID:       bookContent.BookID,
			PageNumber:   bookContent.PageNumber,
			ImageFileUrl: imageFileUrl,
			Description:  bookContent.Description,
			ThumbnailUrl: file_bucket.GenCacheUrl(imageFileUrl, model.PRESET_THUMBNAIL, 0, 0),
		}

		bookContentDatas = append(bookContentDatas, bookContentData)
	}

	var pdfUrl string
	// TODO: implement logic
	// if book.Storage == model.STORAGE_R2 {
	// 	pdfUrl, _ = datastore.GetObjectUrl(ctx, book.PdfFileGuid)
	// } else {
	// 	pdfUrl = utils.GenRawFileUrl(config.Get().FileBucketPath, book.PdfFileGuid)
	// }

	coverFileUrl := file_bucket.GenFinalUrl(ctx, book.Storage, book.CoverFilePath)
	bookDetail := contract_resp.BookDetail{
		ID:           book.ID,
		Slug:         book.Slug,
		Title:        book.Title,
		Description:  book.Description,
		CoverFileUrl: coverFileUrl,
		ThumbnailUrl: file_bucket.GenCacheUrl(coverFileUrl, model.PRESET_COVER, 0, 0),
		Tags:         book.Tags,
		Type:         book.Type,
		AccessTags:   book.AccessTags,
		Contents:     bookContentDatas,
		Active:       book.Active,
		PdfUrl:       pdfUrl,
		CanAction:    slices.Contains(model.ADMIN_ROLES, params.UserRole),
		IsFree:       book.IsFree(),
		IsSubscribed: isSubscribed,
		MaxPage:      len(bookContents),
	}

	return bookDetail, nil
}

func DeleteBook(ctx context.Context, params contract.DeleteBook) error {
	book, err := book_repo.GetByID(ctx, params.BookID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	bookContents, err := book_content_repo.GetByBookID(ctx, book.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	fileBucketGuids := make([]string, 0, len(bookContents)+2)
	fileBucketGuids = append(fileBucketGuids, book.PdfFileGuid, book.CoverFileGuid)
	for _, bookContent := range bookContents {
		fileBucketGuids = append(fileBucketGuids, bookContent.ImageFileGuid)
	}

	fileBuckets, err := file_bucket_repo.GetByGuids(ctx, fileBucketGuids)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	err = datastore.Transaction(ctx, datastore.Get().Db, func(tx *sqlx.Tx) error {
		err = user_stroke_repo.DeleteByBookID(ctx, tx, book.ID)
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = user_activity_repo.DeleteByBookIDs(ctx, tx, pq.Int64Array{book.ID})
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = book_content_repo.DeleteByBookID(ctx, tx, book.ID)
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = book_repo.Delete(ctx, tx, book.ID)
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = file_bucket_repo.DeleteByGuids(ctx, tx, fileBucketGuids)
		if err != nil && err != sql.ErrNoRows {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		if book.Storage == model.STORAGE_R2 {
			keys := make([]string, 0, len(fileBuckets))
			for _, fileBucket := range fileBuckets {
				keys = append(keys, fileBucket.ExactPath)
			}
			err = datastore.DeleteByKeys(ctx, keys)

		} else {
			err = file_bucket.DeleteFolder(fmt.Sprintf("%s/books/%s", config.Get().FileBucketPath, book.Slug))
		}
		if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func UpdateBook(ctx context.Context, params contract.UpdateBook) error {
	book, err := book_repo.GetByID(ctx, params.ID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	book.Slug = params.Slug
	book.Title = params.Title
	book.Description = params.Description
	book.Tags = params.Tags
	book.Type = params.Type
	book.Active = params.Active
	book.OriginalPdfUrl = params.OriginalPdfUrl
	book.AccessTags = params.AccessTags

	err = book_repo.Update(ctx, nil, book)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	return nil
}

func RemoveBookPage(ctx context.Context, params contract.RemoveBookPage) error {
	var err error

	book, err := book_repo.GetByID(ctx, params.BookID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	bookContents, err := book_content_repo.GetByIDs(ctx, params.BookContentIDs)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	fileBucketGuids := make([]string, 0, len(bookContents)+1)
	for _, bookContent := range bookContents {
		fileBucketGuids = append(fileBucketGuids, bookContent.ImageFileGuid)
	}

	fileBuckets, err := file_bucket_repo.GetByGuids(ctx, fileBucketGuids)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	if book.Storage == model.STORAGE_R2 {
		keys := make([]string, 0, len(fileBuckets))
		for _, fileBucket := range fileBuckets {
			keys = append(keys, fileBucket.ExactPath)
		}
		err = datastore.DeleteByKeys(ctx, keys)

	} else {
		for _, fileBucket := range fileBuckets {
			err = file_bucket.DeleteFileIfExists(fmt.Sprintf("%s/%s", config.Get().FileBucketPath, fileBucket.ExactPath))
		}
	}
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	err = file_bucket_repo.DeleteByGuids(ctx, nil, fileBucketGuids)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	for _, bookContent := range bookContents {
		err = book_content_repo.Delete(ctx, nil, bookContent.ID)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
	}

	return nil
}

func UpdateBookCover(ctx context.Context, params contract.UpdateBookCover) error {
	book, err := book_repo.GetByID(ctx, params.BookID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	coverFileUrl := file_bucket.GenFinalUrl(ctx, book.Storage, book.CoverFilePath)
	err = datastore.DeleteCache(coverFileUrl, model.PRESET_COVER, "", "")
	if err != nil {
		logrus.WithContext(ctx).Error(err)
	}

	fileBucket, err := file_bucket_repo.GetByGuid(ctx, book.CoverFileGuid)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	defer params.CoverFile.Close()
	coverFileBytes, err := io.ReadAll(params.CoverFile)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	bookDir := fmt.Sprintf("%s/books/%s", config.Get().FileBucketPath, book.Slug)
	file_bucket.CreateFolderIfNotExists(bookDir)

	coverObjectKey := fmt.Sprintf("books/%s/cover.jpeg", book.Slug)
	coverPath := fmt.Sprintf("%s/cover.jpeg", bookDir)

	err = os.WriteFile(coverPath, coverFileBytes, 0644)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	if fileBucket.Metadata.Purpose == model.PURPOSE_BOOK_COVER && book.Storage == model.STORAGE_LOCAL {
		return nil
	}

	if fileBucket.Metadata.Purpose == model.PURPOSE_BOOK_COVER && book.Storage == model.STORAGE_R2 {
		err = datastore.UploadFileToR2(ctx, coverPath, coverObjectKey, true, model.DEFAULT_R2_IMAGE_CACHE_SECONDS)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
		return nil
	}

	coverGuid := random.MustGenUUIDTimes(2)
	coverFileBucket := model.FileBucket{
		Guid:            coverGuid,
		Name:            fmt.Sprintf("book %v - cover", book.ID),
		BaseType:        "image",
		Extension:       "jpeg",
		HttpContentType: "image/jpeg",
		Metadata:        model.FileBucketMetadata{Purpose: model.PURPOSE_BOOK_COVER},
		Data:            []byte{},
		ExactPath:       coverObjectKey,
		Storage:         book.Storage,
		SizeKb:          0,
	}
	_, book.CoverFileGuid, err = file_bucket_repo.Insert(ctx, nil, coverFileBucket)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	err = book_repo.Update(ctx, nil, book)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	if book.Storage == model.STORAGE_R2 {
		err = datastore.UploadFileToR2(ctx, coverPath, coverObjectKey, true, model.DEFAULT_R2_IMAGE_CACHE_SECONDS)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
		return nil
	}

	return nil
}

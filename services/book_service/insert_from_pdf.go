package book_service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/config"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/datastore"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/book_content_repo"
	"github.com/umarkotak/ytkidd-api/repos/book_repo"
	"github.com/umarkotak/ytkidd-api/repos/file_bucket_repo"
	"github.com/umarkotak/ytkidd-api/utils/file_bucket"
	"github.com/umarkotak/ytkidd-api/utils/random"
)

const CAESIUMCLT_PATH = "/opt/homebrew/bin/caesiumclt"

func InsertFromPdf(ctx context.Context, params contract.InsertFromPdf, uploadState *model.UploadState) error {
	var err error

	unixStr := strconv.FormatInt(time.Now().Unix(), 10)
	lastFour := unixStr[len(unixStr)-4:]
	params.Slug = fmt.Sprintf("%s-%s", lastFour, params.Slug)

	if uploadState != nil {
		uploadState.StatusMap[params.Slug] = model.UploadBookStatus{
			Slug:        params.Slug,
			CreatedAt:   time.Now(),
			CurrentPage: 0,
			MaxPage:     0,
		}
		defer func() {
			delete(uploadState.StatusMap, params.Slug)
		}()
	}

	if params.Storage == "" {
		params.Storage = model.STORAGE_LOCAL
	}

	book, _ := book_repo.GetBySlug(ctx, params.Slug)
	if book.ID != 0 {
		return fmt.Errorf("book already exists")
	}

	logrus.Infof("EXECUTING UPLOAD: %+v", map[string]any{
		"img_format": params.ImgFormat,
	})

	bookDir := fmt.Sprintf("%s/books/%s", config.Get().FileBucketPath, params.Slug)

	err = file_bucket.CreateFolderIfNotExists(bookDir)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}
	defer func() {
		if err != nil {
			file_bucket.DeleteFolder(bookDir)
			if params.Storage == model.STORAGE_R2 {
				datastore.DeleteObjectsByPrefix(ctx, fmt.Sprintf("books/%s", params.Slug))
			}
		}
		if params.Storage == model.STORAGE_R2 {
			file_bucket.DeleteFolder(bookDir)
		}
	}()

	pdfFilePath := fmt.Sprintf("%s/book.pdf", bookDir)
	err = os.WriteFile(pdfFilePath, params.PdfBytes, 0644)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	outputPattern := filepath.Join(bookDir, "%04d.jpeg")
	if params.ImgFormat == model.IMAGE_PNG {
		outputPattern = filepath.Join(bookDir, "%04d.png")
	}

	gsArgs := []string{
		"-dNOPAUSE", //
		"-dBATCH",   //
		"-dSAFER",   //
		"-r225",     // Resolution in DPI
	}
	if params.ImgFormat == model.IMAGE_PNG {
		gsArgs = append(gsArgs, "-sDEVICE=png16m")
	} else {
		gsArgs = append(gsArgs, "-sDEVICE=jpeg", "-dJPEGQ=80")
	}
	gsArgs = append(gsArgs,
		fmt.Sprintf("-sOutputFile=%s", outputPattern), //
		pdfFilePath, //
	)

	cmd := exec.Command("/opt/homebrew/bin/gs", gsArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"cmd_output": output,
		}).Error(err)
		return err
	}

	pattern := filepath.Join(bookDir, "*.jpeg")
	if params.ImgFormat == model.IMAGE_PNG {
		pattern = filepath.Join(bookDir, "*.png")
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	if len(matches) == 0 {
		err = fmt.Errorf("empty pages")
		logrus.WithContext(ctx).Error(err)
		return err
	}

	bookObjectKey := fmt.Sprintf("books/%s/book.pdf", params.Slug)
	if params.Storage == model.STORAGE_R2 {
		if params.StorePdf {
			err = datastore.UploadFileToR2(ctx, pdfFilePath, bookObjectKey, false, model.DEFAULT_R2_IMAGE_CACHE_SECONDS)
			if err != nil {
				logrus.WithContext(ctx).Error(err)
				return err
			}
		}
	}

	successFilePaths := []string{}
	err = datastore.Transaction(ctx, datastore.Get().Db, func(tx *sqlx.Tx) error {
		pdfFileGuid := random.MustGenUUIDTimes(2)
		bookType := model.BOOK_TYPE_DEFAULT
		if params.BookType != "" {
			bookType = params.BookType
		}
		book := model.Book{
			Slug:           params.Slug,
			Title:          params.Title,
			Description:    params.Description,
			CoverFileGuid:  "",
			Tags:           params.Tags,
			Type:           bookType,
			PdfFileGuid:    pdfFileGuid,
			Active:         true,
			OriginalPdfUrl: params.OriginalPdfUrl,
			AccessTags:     pq.StringArray{model.ACCESS_TAG_FREE, model.ACCESS_TAG_PREMIUM},
			Storage:        params.Storage,
		}
		book.ID, err = book_repo.Insert(ctx, tx, book)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
		_, _, err = file_bucket_repo.Insert(ctx, tx, model.FileBucket{
			Guid:            pdfFileGuid,
			Name:            params.Slug,
			BaseType:        "application",
			Extension:       "pdf",
			HttpContentType: "application/pdf",
			Metadata:        model.FileBucketMetadata{Purpose: model.PURPOSE_BOOK_PDF},
			Data:            []byte{},
			ExactPath:       bookObjectKey,
			Storage:         params.Storage,
			SizeKb:          0,
		})
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		// iterate each book page and save to storage
		if uploadState != nil {
			tempUploadState := uploadState.StatusMap[params.Slug]
			tempUploadState.MaxPage = len(matches)
			uploadState.StatusMap[params.Slug] = tempUploadState
		}
		for idx, filePath := range matches {
			if uploadState != nil {
				tempUploadState := uploadState.StatusMap[params.Slug]
				tempUploadState.CurrentPage = idx + 1
				uploadState.StatusMap[params.Slug] = tempUploadState
			}

			if params.ImgFormat == model.IMAGE_PNG {
				cmdComp1 := exec.Command(
					CAESIUMCLT_PATH, "--lossless",
					"-o", bookDir, filePath,
				)
				output1, err := cmdComp1.CombinedOutput()
				if err != nil {
					logrus.WithContext(ctx).WithFields(logrus.Fields{
						"cmd_output": output1,
					}).Error(err)
					return err
				}

			} else {
				cmdComp1 := exec.Command(
					CAESIUMCLT_PATH, "--lossless",
					"-o", bookDir, filePath,
				)
				output1, err := cmdComp1.CombinedOutput()
				if err != nil {
					logrus.WithContext(ctx).WithFields(logrus.Fields{
						"cmd_output": output1,
					}).Error(err)
					return err
				}

				cmdComp2 := exec.Command(
					CAESIUMCLT_PATH, "-q", "60",
					"-o", bookDir, filePath,
				)
				output2, err := cmdComp2.CombinedOutput()
				if err != nil {
					logrus.WithContext(ctx).WithFields(logrus.Fields{
						"cmd_output": output2,
					}).Error(err)
					return err
				}
			}
			logrus.Infof("comporessing image finish")

			bookContentObjectKey := fmt.Sprintf("books/%s/%04d.jpeg", params.Slug, idx+1)
			if params.Storage == model.STORAGE_R2 {
				err = datastore.UploadFileToR2(ctx, filePath, bookContentObjectKey, false, model.DEFAULT_R2_IMAGE_CACHE_SECONDS)
				if err != nil {
					logrus.WithContext(ctx).Error(err)
					return err
				}
			}

			guid := random.MustGenUUIDTimes(2)
			fileBucket := model.FileBucket{
				Guid:            guid,
				Name:            fmt.Sprintf("book %v - page %v", book.ID, idx+1),
				BaseType:        "image",
				Extension:       params.ImgFormat,
				HttpContentType: fmt.Sprintf("image/%s", params.ImgFormat),
				Metadata:        model.FileBucketMetadata{Purpose: model.PURPOSE_BOOK_CONTENT},
				Data:            []byte{},
				ExactPath:       bookContentObjectKey,
				Storage:         params.Storage,
				SizeKb:          0,
			}
			_, fileGuid, err := file_bucket_repo.Insert(ctx, tx, fileBucket)
			if err != nil {
				logrus.WithContext(ctx).Error(err)
				return err
			}

			bookContent := model.BookContent{
				BookID:        book.ID,
				PageNumber:    int64(idx + 1),
				ImageFileGuid: fileGuid,
				Description:   "",
				Metadata:      model.BookContentMetadata{},
			}
			_, err = book_content_repo.Insert(ctx, tx, bookContent)
			if err != nil {
				logrus.WithContext(ctx).Error(err)
				return err
			}

			successFilePaths = append(successFilePaths, filePath)
			logrus.WithContext(ctx).Infof("success inserting image %v/%v", idx+1, len(matches))
		}

		err = file_bucket.CopyFile(matches[0], fmt.Sprintf("%s/%s.%s", bookDir, "cover", params.ImgFormat))
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}
		bookCoverObjectKey := fmt.Sprintf("books/%s/cover.%s", params.Slug, params.ImgFormat)
		if params.Storage == model.STORAGE_R2 {
			err = datastore.UploadFileToR2(
				ctx, fmt.Sprintf("%s/%s", config.Get().FileBucketPath, bookCoverObjectKey), bookCoverObjectKey, false, model.DEFAULT_R2_IMAGE_CACHE_SECONDS,
			)
			if err != nil {
				logrus.WithContext(ctx).Error(err)
				return err
			}
		}

		coverGuid := random.MustGenUUIDTimes(2)
		fileBucket := model.FileBucket{
			Guid:            coverGuid,
			Name:            fmt.Sprintf("book %v - cover", book.ID),
			BaseType:        "image",
			Extension:       params.ImgFormat,
			HttpContentType: fmt.Sprintf("image/%s", params.ImgFormat),
			Metadata:        model.FileBucketMetadata{Purpose: model.PURPOSE_BOOK_COVER},
			Data:            []byte{},
			ExactPath:       bookCoverObjectKey,
			Storage:         params.Storage,
			SizeKb:          0,
		}
		_, book.CoverFileGuid, err = file_bucket_repo.Insert(ctx, tx, fileBucket)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		err = book_repo.Update(ctx, tx, book)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}

	if !params.StorePdf {
		err = file_bucket.DeleteFileIfExists(pdfFilePath)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
		}
	}

	return nil
}

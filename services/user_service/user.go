package user_service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/google_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/utils/random"
	"github.com/umarkotak/ytkidd-api/utils/user_auth"
)

func SignIn(ctx context.Context, params contract.UserSignIn) (contract_resp.UserSignIn, error) {
	// logrus.Infof("GOOGLE TOKEN: %+v\n", params.GoogleCredential)

	googleUser, err := google_repo.ValidateGoogleJWT(params.GoogleCredential)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.UserSignIn{}, err
	}

	// logrus.Infof("GOOGLE DATA: %+v\n", googleUser)

	user, err := user_repo.GetByEmail(ctx, googleUser.Email)
	if err != nil && err != sql.ErrNoRows {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.UserSignIn{}, err
	}

	if user.ID == 0 {
		user = model.User{
			Email:    googleUser.Email,
			Name:     googleUser.Name,
			Username: strings.ReplaceAll(googleUser.Email, "@gmail.com", ""),
			PhotoUrl: googleUser.Picture,
		}

		user.ID, user.Guid, err = user_repo.Insert(ctx, nil, user)
		if err != nil {
			logrus.WithContext(ctx).Error(err)
			return contract_resp.UserSignIn{}, err
		}
	}

	authPayload := user_auth.Payload{
		JWTID:          random.MustGenUUIDTimes(2),
		SID:            random.MustGenUUIDTimes(2),
		Issuer:         "cookiekid-be",
		IssuedAt:       time.Now().Unix(),
		ExpirationTime: time.Now().Add(model.USER_AUTH_EXPIRY).Unix(),
		GUID:           user.Guid,
		Name:           user.Name,
		Username:       user.Username,
		Email:          user.Email,
		PhotoUrl:       user.PhotoUrl,
		UserRole:       user.UserRole,
	}
	accessToken, err := user_auth.GenToken(ctx, authPayload)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.UserSignIn{}, err
	}

	return contract_resp.UserSignIn{
		AccessToken: accessToken,
	}, nil
}

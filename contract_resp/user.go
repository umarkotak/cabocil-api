package contract_resp

import (
	"time"
)

type (
	UserSignUp struct{}

	UserSignIn struct {
		AccessToken string `json:"access_token"`
	}

	User struct {
		ID                  int64     `json:"id"`         //
		CreatedAt           time.Time `json:"created_at"` //
		UpdatedAt           time.Time `json:"updated_at"` //
		Guid                string    `json:"guid"`       //
		Email               string    `json:"email"`      //
		About               string    `json:"about"`      //
		Password            string    `json:"password"`   //
		Name                string    `json:"name"`       //
		Username            string    `json:"username"`   //
		PhotoUrl            string    `json:"photo_url"`  //
		UserRole            string    `json:"user_role"`  // Enum: basic, admin
		SubscriptionEndedAt time.Time `json:"subscription_ended_at"`
	}
)

package contract

type (
	UserSignUp struct {
		Email    string `json:"email" validate:"required"`
		Password string `json:"password" validate:"required"`
		Name     string `json:"name"`
		Username string `json:"username"`
		About    string `json:"about"`
	}

	UserSignIn struct {
		GoogleCredential string `json:"google_credential"`
	}

	UserGetParams struct {
		Guid     string `db:"guid"`
		Email    string `db:"email"`
		Name     string `db:"name"`
		Username string `db:"username"`
	}
)

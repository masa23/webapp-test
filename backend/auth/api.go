package auth

/*
const (
	MinAccessTokenLength = 64
	MinSecretTokenLength = 72
)

// APIキー認証: ヘッダーから認証情報取得・検証
func APIKeyAuth(c echo.Context, db *gorm.DB) (*uint64, error) {
	accessToken, secretToken, err := parseAPIKeyHeader(c)
	if err != nil {
		return nil, err
	}

	// APIトークン取得
	apiToken := &model.APIToken{}
	if err := db.Where("access_token = ?", accessToken).First(apiToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, unauthorized(c)
		}
		return nil, errorMessage(c, "Database error: "+err.Error())
	}

	// シークレットトークン検証
	if err := bcrypt.CompareHashAndPassword([]byte(apiToken.SecretToken), []byte(secretToken)); err != nil {
		return nil, unauthorized(c)
	}

	// ユーザー取得
	user := &model.User{}
	if err := db.Where("api_token_id = ?", apiToken.ID).First(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, unauthorized(c)
		}
		return nil, errorMessage(c, "Database error: "+err.Error())
	}

	return &user.ID, nil
}

func parseAPIKeyHeader(c echo.Context) (string, string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", "", unauthorized(c)
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "apikey" {
		return "", "", unauthorized(c)
	}

	tokens := strings.SplitN(parts[1], ":", 2)
	if len(tokens) != 2 || len(tokens[0]) < MinAccessTokenLength || len(tokens[1]) < MinSecretTokenLength {
		return "", "", unauthorized(c)
	}
	return tokens[0], tokens[1], nil
}
*/

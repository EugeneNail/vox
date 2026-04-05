package application

import "time"

// TokenSalt signs JWTs until a dedicated secret is introduced.
const TokenSalt = "TEMP_SALT"

// LoginTokenType identifies a short-lived authentication token.
const LoginTokenType = "login-token"

// RefreshTokenType identifies a refresh token used to issue a new token pair.
const RefreshTokenType = "refresh-token"

// LoginTokenTTL defines how long the login token stays valid.
const LoginTokenTTL = 15 * time.Minute

// RefreshTokenTTL defines how long the refresh token stays valid.
const RefreshTokenTTL = 7 * 24 * time.Hour

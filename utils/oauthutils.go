package utils

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/context"
	"github.com/sirupsen/logrus"
)

const InvalidRequest = "invalid_request"
const InvalidClient = "invalid_client"
const InvalidGrant = "invalid_grant"
const UnauthorizedClient = "unauthorized_client"
const UnsupportedGrantType = "unsupported_grant_type"
const InvalidScope = "invalid_scope"
const AccessDenied = "access_denied"
const ServerError = "server_error"
const TemporarilyUnavailable = "temporarily_unavailable"

// UserIDKey Key for context access to get the validated userID
const ClaimsContextKey = "ClaimsKey"

const MobileAuthorizedKey = "MobileAuthKey"

const JsonBodyKey = "JsonBodyKey"
const JsonBodyNakedKey = "JsonBodyNakedKey"

var logger = logrus.New().WithField("module", "oauth")
var signingMethod = jwt.SigningMethodHS256

// CustomClaims Structure of JWT body, contains standard JWT claims and userID as a custom claim
type CustomClaims struct {
	UserID   uint64 `json:"userID"`
	AppID    uint64 `json:"appID"`
	DeviceID uint64 `json:"deviceID"`
	Package  string `json:"package"`
	Theme    string `json:"theme"`
	jwt.StandardClaims
}

// OAuthResponse Structure of an successful OAuth response
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// OAuthErrorResponse Structure of an OAuth error response
type OAuthErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// CreateAccessToken Creates a new access token for a given user
func CreateAccessToken(userID, appID, deviceID uint64, pkg, theme string) (string, int, error) {
	expiresIn := Config.Frontend.JwtValidityInMinutes * 60

	standardlaims := jwt.StandardClaims{
		ExpiresAt: createExpiration(int64(expiresIn)),
		Issuer:    Config.Frontend.JwtIssuer,
	}

	token := jwt.NewWithClaims(signingMethod, CustomClaims{
		userID,
		appID,
		deviceID,
		pkg,
		theme,
		standardlaims,
	})

	signKey, err := getSignKey()
	if err != nil {
		return "", 0, err
	}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		logger.Errorf("Error signing an jwt token: %v", err)
		return "", 0, err
	}

	return tokenString, expiresIn, nil
}

// ValidateAccessTokenGetClaims validates the jwt token and returns the UserID
func ValidateAccessTokenGetClaims(tokenString string) (*CustomClaims, error) {
	return accessTokenGetClaims(tokenString, true)
}

// UnsafeGetClaims this method returns the userID of a given jwt token WITHOUT VALIDATION
// DO NOT USE THIS METHOD AS RELIABLE SOURCE FOR USERID
func UnsafeGetClaims(tokenString string) (*CustomClaims, error) {
	return accessTokenGetClaims(tokenString, false)
}

func stripOffBearerFromToken(tokenString string) (string, error) {
	if len(tokenString) > 6 && strings.ToUpper(tokenString[0:6]) == "BEARER" {
		return tokenString[7:], nil
	}
	return tokenString, nil //"", errors.New("Only bearer tokens are supported, got: " + tokenString)
}

func accessTokenGetClaims(tokenStringFull string, validate bool) (*CustomClaims, error) {
	tokenString, err := stripOffBearerFromToken(tokenStringFull)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return getSignKey()
	})

	if err != nil && validate {
		if !strings.Contains(err.Error(), "token is expired") {
			logger.WithFields(
				logrus.Fields{
					"error":       err,
					"token":       token,
					"tokenString": tokenString,
				},
			).Warn("Error parsing jwt token")
		}

		return nil, err
	}

	if token == nil {
		return nil, fmt.Errorf("error token is not defined %v", tokenStringFull)
	}

	// Make sure header hasnt been tampered with
	if token.Method != signingMethod {
		return nil, errors.New("only SHA256hmac as signature method is allowed")
	}

	claims, ok := token.Claims.(*CustomClaims)

	// Check issuer claim
	if claims.Issuer != Config.Frontend.JwtIssuer {
		return nil, errors.New("invalid issuer claim")
	}

	valid := ok && token.Valid

	if valid || !validate {
		return claims, nil
	}

	return nil, errors.New("token validity or claims cannot be verified")
}

func createExpiration(validForSeconds int64) int64 {
	return time.Now().Unix() + validForSeconds
}

func getSignKey() ([]byte, error) {
	signSecret, err := hex.DecodeString(Config.Frontend.JwtSigningSecret)
	if err != nil {
		logger.Errorf("Error decoding jwtSecretKey, not in hex format or missing from config? %v", err)
		return nil, err
	}
	return signSecret, nil
}

// SendOAuthResponse creates and sends a OAuth response according to RFC6749
func SendOAuthResponse(j *json.Encoder, route, accessToken, refreshToken string, expiresIn int) {
	response := OAuthResponse{
		AccessToken:  accessToken,
		ExpiresIn:    expiresIn,
		TokenType:    "bearer",
		RefreshToken: refreshToken,
	}
	err := j.Encode(response)

	if err != nil {
		logger.Errorf("error serializing json error for API %v route: %v", route, err)
	}
}

// SendOAuthErrorResponse creates and sends a OAuth error response according to RFC6749
func SendOAuthErrorResponse(j *json.Encoder, route, errString, description string) {
	response := OAuthErrorResponse{
		Error:       errString,
		Description: description,
	}
	err := j.Encode(response)

	if err != nil {
		logger.Errorf("error serializing json error for API %v route: %v", route, err)
	}
}

func GetAuthorizationClaims(r *http.Request) *CustomClaims {
	accessToken := r.Header.Get("Authorization")
	if len(accessToken) <= 0 {
		return nil
	}

	claims, err := ValidateAccessTokenGetClaims(accessToken)
	if err != nil {
		logger.Warnf("ValidateAccessTokenGetClaims failed") // #REMOVE just for test purpose, can be removed after testing
		return nil
	}
	return claims
}

// AuthorizedAPIMiddleware Demands an Authorization header to be present with a valid user api token
// Once authorization passes, this middleware sets a context entry with the authenticated userID
func AuthorizedAPIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accessToken := r.Header.Get("Authorization")
		if len(accessToken) <= 0 {
			j := json.NewEncoder(w)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			SendOAuthErrorResponse(j, r.URL.String(), InvalidRequest, "missing authorization header")
			return
		}

		claims, err := ValidateAccessTokenGetClaims(accessToken)
		if err != nil {
			j := json.NewEncoder(w)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			SendOAuthErrorResponse(j, r.URL.String(), UnauthorizedClient, "authorization failed")
			return
		}

		// Parse a json post request right here and attach the content
		// as context to the request. By doing this here,
		// we can use base.FormValueOrJson(key) without multiple parsings of the same body
		if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
			body, err := ioutil.ReadAll(r.Body)
			if err == nil {
				keyVal := make(map[string]interface{})
				json.Unmarshal(body, &keyVal)
				context.Set(r, JsonBodyKey, keyVal)
				context.Set(r, JsonBodyNakedKey, body)
			}
		}

		context.Set(r, ClaimsContextKey, claims)
		context.Set(r, MobileAuthorizedKey, true)
		next.ServeHTTP(w, r)
	})
}

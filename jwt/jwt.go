package jwt

import (
	"cerpApi/cfg_details"
	"errors"
	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"strings"
)

func VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	tokenSlice := strings.Split(tokenStr, " ")
	var bearerToken string
	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}

	// if no bearer token set return unauthorized.
	if bearerToken == "" {
		return nil, errors.New("unauthorized")
	}

	jwks, err := fetchJWKS()
	if err != nil {
		return nil, err
	}

	// Parse takes the token string using function to looking up the key.
	token, err := jwt.Parse(bearerToken, jwks.Keyfunc)
	if err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok {
			if verr.Errors == jwt.ValidationErrorMalformed {
				return nil, errors.New("unauthorized")
			}
			if verr.Errors == jwt.ValidationErrorExpired {
				return nil, errors.New("token is expired")
			}
		}
		return nil, err
	}

	// handle nil token scenario, unlikely to happen.
	if token == nil {
		return nil, errors.New("no token after JWT parsing")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	// check if claims are present and token is valid.
	if ok && token.Valid {
		// return Allow authResponse with userEntity in authorizer context for next lambda in chain.
		err = validateClaims(claims)
		return claims, err
	}
	return nil, nil
}

func validateClaims(claims jwt.MapClaims) error {
	if !claims.VerifyIssuer(cfg_details.CLAIM_ISS, true) || claims["azp"] != cfg_details.CLAIM_CLIENT_ID {
		return errors.New("Issuer/CLIENT_ID is wrong")
	}
	return nil
}

func fetchJWKS() (*keyfunc.JWKS, error) {
	options := keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.KeyFunc\nError:%s\n", err.Error())
		},
		RefreshUnknownKID: true,
	}
	return keyfunc.Get(cfg_details.JWKS, options)
}

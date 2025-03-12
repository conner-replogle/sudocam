package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"messages/jwtmsg"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func SetJWTKey(key []byte) {
	jwtKey = key
}

type ContextKey string

const ContextUserKey ContextKey = "UserID"
const ContextClaimKey ContextKey = "Claims"

func AuthMiddleware(next http.HandlerFunc, cameraAllowed bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := ExtractToken(r)
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized User", http.StatusUnauthorized)
			slog.Error("Error validating JWT", "error", err)
			return
		}
		if !cameraAllowed && claims.EntityType != jwtmsg.EntityTypeUser {
			http.Error(w, "Unauthorized User", http.StatusUnauthorized)
			slog.Error("Invalid Entity Type", "entity_type", claims.EntityType)
			return
		}
		ctx := context.WithValue(r.Context(), ContextUserKey, claims.EntityID)
		ctx = context.WithValue(ctx, ContextClaimKey, *claims)

		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func ExtractToken(r *http.Request) string {
	// First try to get token from cookie
	cookie, err := r.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Fallback to Authorization header
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	slog.Error("No token found")
	return ""
}

func ValidateJWT(tokenString string) (*jwtmsg.AuthClaims, error) {
	claims := &jwtmsg.AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

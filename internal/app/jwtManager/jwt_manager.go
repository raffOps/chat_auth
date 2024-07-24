package jwtManager

import (
	"github.com/dgrijalva/jwt-go"
	authModels "github.com/raffops/auth/internal/app/auth/model"
	userModels "github.com/raffops/auth/internal/app/user/models"
	"github.com/raffops/chat/pkg/errs"
	"net/http"
	"slices"
	"strings"
	"time"
)

type jwtManager struct {
	secretKey   []byte
	jwtDuration time.Duration
	roles       map[string][]string
}

func (j *jwtManager) VerifyToken(tokenString string) (*authModels.Claims, errs.ChatError) {
	tokenString, _ = strings.CutPrefix(tokenString, "Bearer ")
	token, err := jwt.ParseWithClaims(tokenString, &authModels.Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errs.NewError(errs.ErrNotAuthorized, nil)
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, errs.NewError(errs.ErrNotAuthorized, err)
	}
	claims, ok := token.Claims.(*authModels.Claims)
	if !ok {
		return nil, errs.NewError(errs.ErrNotAuthorized, nil)
	}
	expired := claims.VerifyExpiresAt(time.Now().Unix(), true)
	if !expired {
		return nil, errs.NewError(errs.ErrNotAuthorized, nil)
	}
	return claims, nil
}

func (j *jwtManager) GenerateToken(user userModels.User, sessionId string) (string, errs.ChatError) {
	claims := &authModels.Claims{
		RoleId:    user.Role,
		SessionId: sessionId,
		StandardClaims: jwt.StandardClaims{
			Subject:   user.Id,
			ExpiresAt: time.Now().Add(j.jwtDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokeString, errSign := token.SignedString(j.secretKey)
	if errSign != nil {
		return "", errs.NewError(errs.ErrInternal, errSign)
	}
	return tokeString, nil
}

func (j *jwtManager) CheckIfAuthorized(f http.HandlerFunc, roles []authModels.RoleId) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "no token provided", http.StatusUnauthorized)
			return
		}
		claims, err := j.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if slices.Contains(roles, claims.RoleId) {
			f(w, r)
		} else {
			http.Error(w, "not authorized", http.StatusForbidden)
		}
	}
}

func NewJwtManager(secretKey []byte, jwtDuration time.Duration) JwtManager {
	return &jwtManager{
		secretKey:   secretKey,
		jwtDuration: jwtDuration,
	}
}

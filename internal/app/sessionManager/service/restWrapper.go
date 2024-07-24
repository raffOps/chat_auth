package service

import (
	"github.com/raffops/auth/internal/app/auth/model"
	"net/http"
	"strings"
)

func (s service) CheckRestSession(next http.HandlerFunc, roles []auth.RoleId) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(token, "Bearer ")
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		result, err := s.repo.GetEncryptedHashmap(r.Context(), "session:"+token, s.secret)
		for _, role := range roles {
			if err == nil && float64(role) == result["role"].(float64) {
				next(w, r)
				return
			}
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

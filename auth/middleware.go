package auth

import (
	"context"
	"net/http"
	"strings"
)

type ContextKey string
const UserContextKey ContextKey = "user"


func AuthMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request)  {
		
	
		token := r.Header.Get("Authorization")

		if strings.TrimSpace(token) == ""{
			http.Error(w, "no authorization token found", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := ValidateToken(token)

		if err != nil{
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return	
		}

		rCtx := context.WithValue(r.Context(), UserContextKey, claims)



		next.ServeHTTP(w, r.WithContext(rCtx))
	})
}

func AdminRoleMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*Claims)

		if !ok{
			http.Error(w, "invalid auth context", http.StatusInternalServerError)
			return 
		}

		if strings.ToLower(claims.Role) != "admin"{
			http.Error(w, "authorization required", http.StatusForbidden)
			return 
		} 

		next.ServeHTTP(w, r)
	})
}
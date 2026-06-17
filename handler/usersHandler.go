package handler

import (
	"encoding/json"
	"go_crud/auth"
	"go_crud/service"
	"go_crud/store"
	"net/http"
)

type UserService interface{
	Register(service.UserRegisterInput) (*store.User, error)
	Login(service.UserLoginInput) (*store.User, error)
	Profile(service.UserProfileInput) (*store.User, error)
}

type UserHandler struct{
	service UserService
}

type RegisterRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
	Role string `json:"role"`
}

type RegisterResponse struct{
	UserID int `json:"user_id"`
	Username string `json:"username"`
	Role string `json:"role"`
}

type LoginRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
}


type ProfileResponse struct{
	Username string `json:"username"`
	Role string `json:"role"`
}


func NewUserHandler(userService UserService) *UserHandler{
	return &UserHandler{
		service: userService,
	}
}

func (handler *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var body RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil{
		http.Error(w, "error decoding request body", http.StatusInternalServerError)
		return
	}

	user, err := handler.service.Register(service.UserRegisterInput{Username: body.Username, Password: body.Password, Role: body.Role})

	status := map[error]int{
		service.ErrUsernameAlreadyExists: http.StatusConflict,
		service.ErrInvalidUsername: http.StatusBadRequest,
		service.ErrInavlidPassword: http.StatusBadRequest,
		service.ErrUsernameLengthIsShort: http.StatusBadRequest,
		service.ErrPasswordLengthIsShort: http.StatusBadRequest,
	}

	if err != nil{
		code, exist := status[err]

		if !exist{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusCreated)

	
	if err = json.NewEncoder(w).Encode(&RegisterResponse{UserID: user.ID, Username: user.Username, Role: user.Role}); err != nil{
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}

	
}

func (handler *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	var body LoginRequest
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil{
		http.Error(w, "error decoding request body", http.StatusInternalServerError)
		return
	}

	user, err := handler.service.Login(service.UserLoginInput{Username: body.Username, Password: body.Password})

	status := map[error]int{
		store.ErrUserNotFound: http.StatusNotFound,
		service.ErrInvalidUsername: http.StatusBadRequest,
		service.ErrInavlidPassword: http.StatusBadRequest,
		service.ErrUsernameLengthIsShort: http.StatusBadRequest,
		service.ErrPasswordLengthIsShort: http.StatusBadRequest,
		service.ErrIncorrectUsernameOrPassword: http.StatusUnauthorized,
	}

	if err != nil{
		code, exist := status[err]

		if !exist{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), code)
		return
	}

	token, err := auth.GenerateToken(
		user.ID,
		user.Username,
		user.Role,
	)

	if err != nil{
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&LoginResponse{Token: token}); err != nil{
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}
	

}

func (handler *UserHandler) Profile(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	claims, ok := r.Context().Value(auth.UserContextKey).(*auth.Claims)

	if !ok{
		http.Error(w, "invalid auth context", http.StatusInternalServerError)
		return
	}

	user, err := handler.service.Profile(service.UserProfileInput{ID: claims.UserID})

	status := map[error]int{
		store.ErrUserNotFound: http.StatusNotFound,
		service.ErrIncorrectUsernameOrPassword: http.StatusUnauthorized,
	}

	if err != nil{
		code, exist := status[err]

		if !exist{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), code)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&ProfileResponse{Username: user.Username, Role: user.Role}); err != nil{
		http.Error(w, "failed encode response", http.StatusInternalServerError)
		return
	}
}


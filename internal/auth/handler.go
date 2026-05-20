package auth

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func decode[T any](r *http.Request, v *T) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func respond(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := decode(r, &body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user, err := h.service.Register(r.Context(), body.Email, body.Password)
	if err != nil {
		switch err {
		case ErrEmailTaken:
			respond(w, http.StatusConflict, map[string]string{"error": "email already taken"})
		default:
			respond(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}
	respond(w, http.StatusCreated, map[string]string{"id": user.ID, "email": user.Email})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := decode(r, &body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	accessToken, refreshToken, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			respond(w, http.StatusForbidden, map[string]string{"error": "invalid email or password"})
		default:
			respond(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	respond(w, http.StatusOK, map[string]string{"access_token": accessToken, "refresh_token": refreshToken})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}

	if err := decode(r, &body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.service.Logout(r.Context(), body.Token); err != nil {
		respond(w, http.StatusConflict, map[string]string{"error": "facing issues in logging you out"})
		return
	}

	respond(w, http.StatusOK, map[string]string{"message": "successfully logged out"})

}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}

	if err := decode(r, &body); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	accessToken, refreshToken, err := h.service.RefreshToken(r.Context(), body.Token)
	if err != nil {
		respond(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	respond(w, http.StatusOK, map[string]string{"access_token": accessToken, "refresh_token": refreshToken})

}

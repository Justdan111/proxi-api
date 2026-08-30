package auth

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/go-playground/validator/v10"
    "github.com/Justdan111/proxi-api/pkg/response"

)

type Handler struct {
    service  *Service
    validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
    return &Handler{
        service:  service,
        validate: validator.New(),
    }
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
    var input SignupInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.validate.Struct(input); err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }

    result, err := h.service.Signup(r.Context(), input)
    if err != nil {
        if errors.Is(err, ErrEmailTaken) {
            response.Error(w, http.StatusConflict, "email already in use")
            return
        }
        response.Error(w, http.StatusInternalServerError, "failed to create account")
        return
    }

    response.Success(w, http.StatusCreated, "account created", result)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var input LoginInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.validate.Struct(input); err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }

    result, err := h.service.Login(r.Context(), input)
    if err != nil {
        if errors.Is(err, ErrInvalidCreds) {
            response.Error(w, http.StatusUnauthorized, "invalid email or password")
            return
        }
        response.Error(w, http.StatusInternalServerError, "login failed")
        return
    }

    response.Success(w, http.StatusOK, "login successful", result)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
    userID := GetUserID(r)

    user, err := h.service.GetMe(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusNotFound, "user not found")
        return
    }

    response.Success(w, http.StatusOK, "user profile", user)
}

// DELETE /api/auth/me
//
// Deletion is idempotent: a user that is already gone reports success, so a
// double tap, or a retry after a partially-completed purge, still lets the
// client finish its local teardown instead of stalling on a 404.
func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
    userID := GetUserID(r)

    if err := h.service.DeleteAccount(r.Context(), userID); err != nil {
        if !errors.Is(err, ErrUserNotFound) {
            response.Error(w, http.StatusInternalServerError, "failed to delete account")
            return
        }
    }

    response.Success(w, http.StatusOK, "account deleted", nil)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
   
    response.Success(w, http.StatusOK, "logged out", nil)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    response.Success(w, http.StatusOK, "password reset link sent", nil)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
    response.Success(w, http.StatusOK, "password updated", nil)
}
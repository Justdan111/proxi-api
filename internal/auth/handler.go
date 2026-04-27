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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
   
    response.Success(w, http.StatusOK, "logged out", nil)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
    response.Success(w, http.StatusOK, "password reset link sent", nil)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
    response.Success(w, http.StatusOK, "password updated", nil)
}
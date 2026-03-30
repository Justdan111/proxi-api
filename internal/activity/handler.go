package activity

import (
    "encoding/json"
    "net/http"

    "github.com/go-playground/validator/v10"
    "github.com/Justdan111/proxi-api/internal/auth"
    "github.com/Justdan111/proxi-api/pkg/response"
)

type Handler struct {
    service  *Service
    validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
    return &Handler{service: service, validate: validator.New()}
}

// GET /api/activities
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r)

    activities, err := h.service.GetAll(r.Context(), userID)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "failed to fetch activities")
        return
    }

    response.Success(w, http.StatusOK, "activities fetched", activities)
}

// POST /api/activities
func (h *Handler) Log(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetUserID(r)

    var input LogInput
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        response.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    if err := h.validate.Struct(input); err != nil {
        response.Error(w, http.StatusBadRequest, err.Error())
        return
    }

    activity, err := h.service.Log(r.Context(), userID, input)
    if err != nil {
        response.Error(w, http.StatusInternalServerError, "failed to log activity")
        return
    }

    response.Success(w, http.StatusCreated, "activity logged", activity)
}
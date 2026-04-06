package reminder

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/Justdan111/proxi-api/internal/auth"
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

// GET /api/reminders
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	reminders, err := h.service.GetAll(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch reminders")
		return
	}

	response.Success(w, http.StatusOK, "reminders fetched", reminders)
}

// POST /api/reminders
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	reminder, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create reminder")
		return
	}

	response.Success(w, http.StatusCreated, "reminder created", reminder)
}

// GET /api/reminders/:id
func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	reminderID := chi.URLParam(r, "id")

	reminder, err := h.service.GetOne(r.Context(), userID, reminderID)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid reminder id")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "reminder not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to fetch reminder")
		return
	}

	response.Success(w, http.StatusOK, "reminder fetched", reminder)
}

// PUT /api/reminders/:id
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	reminderID := chi.URLParam(r, "id")

	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(input); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.service.Update(r.Context(), userID, reminderID, input)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid reminder id")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "reminder not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update reminder")
		return
	}

	response.Success(w, http.StatusOK, "reminder updated", updated)
}

// PATCH /api/reminders/:id/toggle
func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	reminderID := chi.URLParam(r, "id")

	updated, err := h.service.Toggle(r.Context(), userID, reminderID)
	if err != nil {
		if errors.Is(err, ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid reminder id")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "reminder not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to toggle reminder")
		return
	}

	response.Success(w, http.StatusOK, "reminder toggled", updated)
}

// DELETE /api/reminders/:id
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)
	reminderID := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), userID, reminderID); err != nil {
		if errors.Is(err, ErrInvalidID) {
			response.Error(w, http.StatusBadRequest, "invalid reminder id")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "reminder not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete reminder")
		return
	}

	response.Success(w, http.StatusOK, "reminder deleted", nil)
}

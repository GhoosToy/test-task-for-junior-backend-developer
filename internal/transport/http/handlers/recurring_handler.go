package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	taskdomain "example.com/taskservice/internal/domain/task"
	taskusecase "example.com/taskservice/internal/usecase/task"
)

// RecurringHandler - хендлер для периодических задач
type RecurringHandler struct {
	usecase taskusecase.RecurringUsecase
}

// NewRecurringHandler - конструктор
func NewRecurringHandler(usecase taskusecase.RecurringUsecase) *RecurringHandler {
	return &RecurringHandler{usecase: usecase}
}

// ========== ЭНДПОИНТЫ ДЛЯ ШАБЛОНОВ ==========

// CreateTemplate - POST /api/v1/templates
func (h *RecurringHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreateTemplateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.CreateTemplate(r.Context(), taskusecase.CreateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  req.Recurrence,
	})
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, ToTemplateResponse(created))
}

// GetTemplateByID - GET /api/v1/templates/{id}
func (h *RecurringHandler) GetTemplateByID(w http.ResponseWriter, r *http.Request) {
	id, err := getTemplateIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	template, err := h.usecase.GetTemplateByID(r.Context(), id)
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ToTemplateResponse(template))
}

// UpdateTemplate - PUT /api/v1/templates/{id}
func (h *RecurringHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getTemplateIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req UpdateTemplateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	input := taskusecase.UpdateTemplateInput{
		Title:       req.Title,
		Description: req.Description,
		Recurrence:  req.Recurrence,
	}

	updated, err := h.usecase.UpdateTemplate(r.Context(), id, input)
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ToTemplateResponse(updated))
}

// DeleteTemplate - DELETE /api/v1/templates/{id}
func (h *RecurringHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := getTemplateIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.DeleteTemplate(r.Context(), id); err != nil {
		writeRecurringError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTemplates - GET /api/v1/templates
func (h *RecurringHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.usecase.ListTemplates(r.Context())
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	response := make([]TemplateResponse, 0, len(templates))
	for i := range templates {
		response = append(response, ToTemplateResponse(&templates[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

// ========== ЭНДПОИНТЫ ДЛЯ ЭКЗЕМПЛЯРОВ ==========

// GetTasksForDate - GET /api/v1/tasks?date=2026-04-07
func (h *RecurringHandler) GetTasksForDate(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, use YYYY-MM-DD"))
		return
	}

	instances, err := h.usecase.GetTasksForDate(r.Context(), date)
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	response := make([]InstanceResponse, 0, len(instances))
	for i := range instances {
		response = append(response, ToInstanceResponse(&instances[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

// GetTasksForDateWithStatus - GET /api/v1/tasks/recurring?date=2026-04-07&status=skipped
func (h *RecurringHandler) GetTasksForDateWithStatus(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid date format, use YYYY-MM-DD"))
		return
	}

	statusFilter := r.URL.Query().Get("status") // new, done, skipped

	instances, err := h.usecase.GetTasksForDate(r.Context(), date)
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	// Фильтруем по статусу если указан
	if statusFilter != "" {
		filtered := make([]taskdomain.TaskInstance, 0)
		for _, inst := range instances {
			if string(inst.Status) == statusFilter {
				filtered = append(filtered, inst)
			}
		}
		instances = filtered
	}

	response := make([]InstanceResponse, 0, len(instances))
	for i := range instances {
		response = append(response, ToInstanceResponse(&instances[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

// CompleteInstance - PATCH /api/v1/instances/{id}/complete
func (h *RecurringHandler) CompleteInstance(w http.ResponseWriter, r *http.Request) {
	id, err := getInstanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.CompleteInstance(r.Context(), id); err != nil {
		writeRecurringError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SkipInstance - PATCH /api/v1/instances/{id}/skip
func (h *RecurringHandler) SkipInstance(w http.ResponseWriter, r *http.Request) {
	id, err := getInstanceIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.usecase.SkipInstance(r.Context(), id); err != nil {
		writeRecurringError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetOccurrences - GET /api/v1/templates/{id}/occurrences?from=2026-04-01&to=2026-04-30
func (h *RecurringHandler) GetOccurrences(w http.ResponseWriter, r *http.Request) {
	id, err := getTemplateIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, errors.New("from and to parameters are required"))
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid from date format, use YYYY-MM-DD"))
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid to date format, use YYYY-MM-DD"))
		return
	}

	occurrences, err := h.usecase.GetOccurrences(r.Context(), id, from, to)
	if err != nil {
		writeRecurringError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, OccurrencesResponse{
		TemplateID:  id,
		Occurrences: occurrences,
	})
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

func getTemplateIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing template id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid template id")
	}

	if id <= 0 {
		return 0, errors.New("invalid template id")
	}

	return id, nil
}

func getInstanceIDFromRequest(r *http.Request) (int64, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return 0, errors.New("missing instance id")
	}

	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, errors.New("invalid instance id")
	}

	if id <= 0 {
		return 0, errors.New("invalid instance id")
	}

	return id, nil
}

func writeRecurringError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, taskdomain.ErrTemplateNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskdomain.ErrInstanceNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, taskdomain.ErrInvalidRecurrenceRule):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, taskdomain.ErrInvalidRecurrenceType):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, taskdomain.ErrAlreadyCompleted):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, taskusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

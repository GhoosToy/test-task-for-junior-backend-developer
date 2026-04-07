package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// ========== DTO ДЛЯ ШАБЛОНОВ ==========

// CreateTemplateRequest - DTO для создания шаблона
type CreateTemplateRequest struct {
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Recurrence  *taskdomain.RecurrenceRule `json:"recurrence"`
}

// UpdateTemplateRequest - DTO для обновления шаблона
type UpdateTemplateRequest struct {
	Title       *string                    `json:"title,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Recurrence  *taskdomain.RecurrenceRule `json:"recurrence,omitempty"`
}

// TemplateResponse - DTO для ответа с шаблоном
type TemplateResponse struct {
	ID          int64                      `json:"id"`
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Recurrence  *taskdomain.RecurrenceRule `json:"recurrence,omitempty"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

// ========== DTO ДЛЯ ЭКЗЕМПЛЯРОВ ==========

// InstanceResponse - DTO для ответа с экземпляром задачи
type InstanceResponse struct {
	ID          int64      `json:"id"`
	TemplateID  int64      `json:"template_id"`
	DueDate     time.Time  `json:"due_date"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// OccurrencesResponse - DTO для ответа со списком дат вхождений
type OccurrencesResponse struct {
	TemplateID  int64       `json:"template_id"`
	Occurrences []time.Time `json:"occurrences"`
}

// ========== КОНВЕРТЕРЫ ==========

// ToTemplateResponse - конвертирует TaskTemplate в TemplateResponse
func ToTemplateResponse(t *taskdomain.TaskTemplate) TemplateResponse {
	return TemplateResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Recurrence:  t.Recurrence,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// ToInstanceResponse - конвертирует TaskInstance в InstanceResponse
func ToInstanceResponse(i *taskdomain.TaskInstance) InstanceResponse {
	return InstanceResponse{
		ID:          i.ID,
		TemplateID:  i.TemplateID,
		DueDate:     i.DueDate,
		Status:      string(i.Status),
		CompletedAt: i.CompletedAt,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
	}
}

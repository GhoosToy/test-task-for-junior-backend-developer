package task

import (
	"context"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

type Usecase interface {
	Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.Task, error)
}

// ========== НОВЫЕ ИНТЕРФЕЙСЫ ДЛЯ ПЕРИОДИЧЕСКИХ ЗАДАЧ ==========

// Repository для шаблонов
type TemplateRepository interface {
	Create(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	Update(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]taskdomain.TaskTemplate, error)
	GetAll(ctx context.Context) ([]taskdomain.TaskTemplate, error)
}

// Repository для экземпляров
type InstanceRepository interface {
	Create(ctx context.Context, instance *taskdomain.TaskInstance) (*taskdomain.TaskInstance, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.TaskInstance, error)
	GetByTemplateAndDate(ctx context.Context, templateID int64, dueDate time.Time) (*taskdomain.TaskInstance, error)
	FindOrCreate(ctx context.Context, templateID int64, dueDate time.Time) (*taskdomain.TaskInstance, error)
	Update(ctx context.Context, instance *taskdomain.TaskInstance) (*taskdomain.TaskInstance, error)
	ListByDate(ctx context.Context, date time.Time) ([]taskdomain.TaskInstance, error)
	ListByTemplate(ctx context.Context, templateID int64) ([]taskdomain.TaskInstance, error)
	MarkOverdue(ctx context.Context) error
}

// Usecase для периодических задач
type RecurringUsecase interface {
	// Работа с шаблонами
	CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.TaskTemplate, error)
	GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error)
	UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.TaskTemplate, error)
	DeleteTemplate(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error)

	// Работа с экземплярами
	GetTasksForDate(ctx context.Context, date time.Time) ([]taskdomain.TaskInstance, error)
	CompleteInstance(ctx context.Context, instanceID int64) error
	SkipInstance(ctx context.Context, instanceID int64) error

	// Генерация дат - ИСПРАВЛЕННЫЙ ТИП
	GetOccurrences(ctx context.Context, templateID int64, from, to time.Time) ([]time.Time, error)

	// Фоновая задача
	GenerateFutureInstances(ctx context.Context, daysAhead int) error

	// НОВЫЙ МЕТОД: пометить просроченные задачи
	MarkOverdue(ctx context.Context) error
}

type CreateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

type UpdateInput struct {
	Title       string
	Description string
	Status      taskdomain.Status
}

// DTO для создания шаблона
type CreateTemplateInput struct {
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Recurrence  *taskdomain.RecurrenceRule `json:"recurrence"`
}

// DTO для обновления шаблона
type UpdateTemplateInput struct {
	Title       *string                    `json:"title,omitempty"`
	Description *string                    `json:"description,omitempty"`
	Recurrence  *taskdomain.RecurrenceRule `json:"recurrence,omitempty"`
}

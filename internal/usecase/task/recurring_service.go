package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// RecurringService - сервис для работы с периодическими задачами
type RecurringService struct {
	templateRepo TemplateRepository
	instanceRepo InstanceRepository
	now          func() time.Time
	generator    *RecurrenceGenerator
}

// NewRecurringService - конструктор
func NewRecurringService(templateRepo TemplateRepository, instanceRepo InstanceRepository) *RecurringService {
	return &RecurringService{
		templateRepo: templateRepo,
		instanceRepo: instanceRepo,
		now:          func() time.Time { return time.Now().UTC() },
		generator:    NewRecurrenceGenerator(),
	}
}

// ========== РАБОТА С ШАБЛОНАМИ ==========

// CreateTemplate - создание шаблона периодической задачи
func (s *RecurringService) CreateTemplate(ctx context.Context, input CreateTemplateInput) (*taskdomain.TaskTemplate, error) {
	// Нормализация и валидация
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	// Валидация правила повторения
	if input.Recurrence == nil {
		return nil, fmt.Errorf("%w: recurrence rule is required for recurring task", ErrInvalidInput)
	}

	if err := input.Recurrence.Validate(); err != nil {
		return nil, err
	}

	// Создаем шаблон
	now := s.now()
	template := &taskdomain.TaskTemplate{
		Title:       input.Title,
		Description: input.Description,
		Recurrence:  input.Recurrence,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	created, err := s.templateRepo.Create(ctx, template)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetTemplateByID - получить шаблон по ID
func (s *RecurringService) GetTemplateByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.templateRepo.GetByID(ctx, id)
}

// UpdateTemplate - обновление шаблона
func (s *RecurringService) UpdateTemplate(ctx context.Context, id int64, input UpdateTemplateInput) (*taskdomain.TaskTemplate, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	// Получаем существующий шаблон
	template, err := s.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	if input.Title != nil {
		template.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil {
		template.Description = strings.TrimSpace(*input.Description)
	}
	if input.Recurrence != nil {
		if err := input.Recurrence.Validate(); err != nil {
			return nil, err
		}
		template.Recurrence = input.Recurrence
	}

	template.UpdatedAt = s.now()

	return s.templateRepo.Update(ctx, template)
}

// DeleteTemplate - удаление шаблона (каскадно удаляет все экземпляры)
func (s *RecurringService) DeleteTemplate(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.templateRepo.Delete(ctx, id)
}

// ListTemplates - список всех шаблонов
func (s *RecurringService) ListTemplates(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	return s.templateRepo.List(ctx)
}

// ========== РАБОТА С ЭКЗЕМПЛЯРАМИ ==========

// GetTasksForDate - получить задачи на конкретную дату
func (s *RecurringService) GetTasksForDate(ctx context.Context, date time.Time) ([]taskdomain.TaskInstance, error) {
	// 1. Получаем все шаблоны
	templates, err := s.templateRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var instances []taskdomain.TaskInstance

	// 2. Для каждого шаблона проверяем, должна ли быть задача на эту дату
	for _, template := range templates {
		if template.Recurrence == nil {
			continue
		}

		// Генерируем даты на +/- 7 дней от запрошенной
		from := date.AddDate(0, 0, -7)
		to := date.AddDate(0, 0, 7)

		dates, err := s.generator.GenerateDates(template.Recurrence, from, to)
		if err != nil {
			continue
		}

		// Проверяем, есть ли наша дата
		shouldHave := false
		for _, d := range dates {
			if isSameDay(d, date) {
				shouldHave = true
				break
			}
		}

		if shouldHave {
			// FindOrCreate - если экземпляр уже существует, берем его
			instance, err := s.instanceRepo.FindOrCreate(ctx, template.ID, date)
			if err != nil {
				continue
			}
			instances = append(instances, *instance)
		}
	}

	return instances, nil
}

// CompleteInstance - отметить экземпляр выполненным
func (s *RecurringService) CompleteInstance(ctx context.Context, instanceID int64) error {
	if instanceID <= 0 {
		return fmt.Errorf("%w: invalid instance id", ErrInvalidInput)
	}

	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}

	if instance.Status == taskdomain.StatusDone {
		return taskdomain.ErrAlreadyCompleted
	}

	now := s.now()
	instance.Status = taskdomain.StatusDone
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	_, err = s.instanceRepo.Update(ctx, instance)
	return err
}

// SkipInstance - пропустить экземпляр
// SkipInstance - пропустить экземпляр
func (s *RecurringService) SkipInstance(ctx context.Context, instanceID int64) error {
	if instanceID <= 0 {
		return fmt.Errorf("%w: invalid instance id", ErrInvalidInput)
	}

	instance, err := s.instanceRepo.GetByID(ctx, instanceID)
	if err != nil {
		return err
	}

	// Проверяем, не выполнена ли уже задача
	if instance.Status == taskdomain.StatusDone {
		return taskdomain.ErrAlreadyCompleted
	}

	// Проверяем, не пропущена ли уже
	if instance.Status == taskdomain.StatusSkipped {
		return taskdomain.ErrAlreadySkipped
	}

	now := s.now()
	instance.Status = taskdomain.StatusSkipped
	instance.CompletedAt = &now
	instance.UpdatedAt = now

	_, err = s.instanceRepo.Update(ctx, instance)
	return err
}

// GetOccurrences - получить даты вхождений для шаблона
func (s *RecurringService) GetOccurrences(ctx context.Context, templateID int64, from, to time.Time) ([]time.Time, error) {
	if templateID <= 0 {
		return nil, fmt.Errorf("%w: invalid template id", ErrInvalidInput)
	}

	template, err := s.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if template.Recurrence == nil {
		return nil, taskdomain.ErrNotRecurringTask
	}

	return s.generator.GenerateDates(template.Recurrence, from, to)
}

// GenerateFutureInstances - фоновая генерация экземпляров на будущее
func (s *RecurringService) GenerateFutureInstances(ctx context.Context, daysAhead int) error {
	templates, err := s.templateRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	today := s.now()
	futureDate := today.AddDate(0, 0, daysAhead)

	for _, template := range templates {
		if template.Recurrence == nil {
			continue
		}

		dates, err := s.generator.GenerateDates(template.Recurrence, today, futureDate)
		if err != nil {
			continue
		}

		for _, date := range dates {
			_, err := s.instanceRepo.FindOrCreate(ctx, template.ID, date)
			if err != nil {
				// Логируем ошибку, но продолжаем
				continue
			}
		}
	}

	return nil
}

// MarkOverdue - помечает просроченные задачи (due_date < today и status = 'new')
func (s *RecurringService) MarkOverdue(ctx context.Context) error {
	return s.instanceRepo.MarkOverdue(ctx)
}

// ========== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ==========

func isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

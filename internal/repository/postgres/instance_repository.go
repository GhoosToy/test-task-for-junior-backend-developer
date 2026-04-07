package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// InstanceRepository - репозиторий для работы с экземплярами задач
type InstanceRepository struct {
	pool *pgxpool.Pool
}

// NewInstanceRepository - конструктор
func NewInstanceRepository(pool *pgxpool.Pool) *InstanceRepository {
	return &InstanceRepository{pool: pool}
}

// Create - создание экземпляра
func (r *InstanceRepository) Create(ctx context.Context, instance *taskdomain.TaskInstance) (*taskdomain.TaskInstance, error) {
	query := `
		INSERT INTO task_instances (template_id, due_date, status, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, template_id, due_date, status, completed_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		instance.TemplateID,
		instance.DueDate,
		instance.Status,
		instance.CompletedAt,
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	return scanInstance(row)
}

// GetByID - получение экземпляра по ID
func (r *InstanceRepository) GetByID(ctx context.Context, id int64) (*taskdomain.TaskInstance, error) {
	query := `
		SELECT id, template_id, due_date, status, completed_at, created_at, updated_at
		FROM task_instances
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	instance, err := scanInstance(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrInstanceNotFound
		}
		return nil, err
	}

	return instance, nil
}

// GetByTemplateAndDate - получить экземпляр по шаблону и дате
func (r *InstanceRepository) GetByTemplateAndDate(ctx context.Context, templateID int64, dueDate time.Time) (*taskdomain.TaskInstance, error) {
	query := `
		SELECT id, template_id, due_date, status, completed_at, created_at, updated_at
		FROM task_instances
		WHERE template_id = $1 AND due_date = $2
	`

	row := r.pool.QueryRow(ctx, query, templateID, dueDate)
	instance, err := scanInstance(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrInstanceNotFound
		}
		return nil, err
	}

	return instance, nil
}

// FindOrCreate - найти или создать экземпляр
func (r *InstanceRepository) FindOrCreate(ctx context.Context, templateID int64, dueDate time.Time) (*taskdomain.TaskInstance, error) {
	// Пытаемся найти существующий
	existing, err := r.GetByTemplateAndDate(ctx, templateID, dueDate)
	if err == nil {
		return existing, nil
	}

	if !errors.Is(err, taskdomain.ErrInstanceNotFound) {
		return nil, err
	}

	// Создаем новый
	now := time.Now()
	instance := &taskdomain.TaskInstance{
		TemplateID: templateID,
		DueDate:    dueDate,
		Status:     taskdomain.StatusNew,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return r.Create(ctx, instance)
}

// Update - обновление экземпляра
func (r *InstanceRepository) Update(ctx context.Context, instance *taskdomain.TaskInstance) (*taskdomain.TaskInstance, error) {
	query := `
		UPDATE task_instances
		SET status = $1,
			completed_at = $2,
			updated_at = $3
		WHERE id = $4
		RETURNING id, template_id, due_date, status, completed_at, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query,
		instance.Status,
		instance.CompletedAt,
		instance.UpdatedAt,
		instance.ID,
	)

	updated, err := scanInstance(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrInstanceNotFound
		}
		return nil, err
	}

	return updated, nil
}

// ListByDate - получить все экземпляры на конкретную дату
func (r *InstanceRepository) ListByDate(ctx context.Context, date time.Time) ([]taskdomain.TaskInstance, error) {
	query := `
		SELECT id, template_id, due_date, status, completed_at, created_at, updated_at
		FROM task_instances
		WHERE due_date = $1
		ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]taskdomain.TaskInstance, 0)
	for rows.Next() {
		instance, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *instance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

// ListByTemplate - получить все экземпляры шаблона
func (r *InstanceRepository) ListByTemplate(ctx context.Context, templateID int64) ([]taskdomain.TaskInstance, error) {
	query := `
		SELECT id, template_id, due_date, status, completed_at, created_at, updated_at
		FROM task_instances
		WHERE template_id = $1
		ORDER BY due_date
	`

	rows, err := r.pool.Query(ctx, query, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]taskdomain.TaskInstance, 0)
	for rows.Next() {
		instance, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *instance)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

// MarkOverdue - помечает просроченные задачи (due_date < today и status = 'new')
func (r *InstanceRepository) MarkOverdue(ctx context.Context) error {
	query := `
		UPDATE task_instances
		SET status = $1, updated_at = $2
		WHERE due_date < $3 AND status = $4
	`

	today := time.Now().UTC().Truncate(24 * time.Hour)
	_, err := r.pool.Exec(ctx, query, taskdomain.StatusDone, time.Now(), today, taskdomain.StatusNew)
	return err
}

// scanInstance - сканирует строку в TaskInstance
func scanInstance(row pgx.Row) (*taskdomain.TaskInstance, error) {
	var instance taskdomain.TaskInstance
	var status string

	err := row.Scan(
		&instance.ID,
		&instance.TemplateID,
		&instance.DueDate,
		&status,
		&instance.CompletedAt,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	instance.Status = taskdomain.Status(status)
	return &instance, nil
}

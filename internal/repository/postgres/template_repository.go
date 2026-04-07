package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// TemplateRepository - репозиторий для работы с шаблонами задач
type TemplateRepository struct {
	pool *pgxpool.Pool
}

// NewTemplateRepository - конструктор
func NewTemplateRepository(pool *pgxpool.Pool) *TemplateRepository {
	return &TemplateRepository{pool: pool}
}

// Create - создание шаблона
func (r *TemplateRepository) Create(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	query := `
		INSERT INTO task_templates (title, description, recurrence, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, recurrence, created_at, updated_at
	`

	// Преобразуем RecurrenceRule в JSON
	var recurrenceJSON []byte
	if template.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(template.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query,
		template.Title,
		template.Description,
		recurrenceJSON,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return scanTemplate(row)
}

// GetByID - получение шаблона по ID
func (r *TemplateRepository) GetByID(ctx context.Context, id int64) (*taskdomain.TaskTemplate, error) {
	query := `
		SELECT id, title, description, recurrence, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	template, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrTemplateNotFound
		}
		return nil, err
	}

	return template, nil
}

// Update - обновление шаблона
func (r *TemplateRepository) Update(ctx context.Context, template *taskdomain.TaskTemplate) (*taskdomain.TaskTemplate, error) {
	query := `
		UPDATE task_templates
		SET title = $1,
			description = $2,
			recurrence = $3,
			updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, recurrence, created_at, updated_at
	`

	// Преобразуем RecurrenceRule в JSON
	var recurrenceJSON []byte
	if template.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(template.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query,
		template.Title,
		template.Description,
		recurrenceJSON,
		template.UpdatedAt,
		template.ID,
	)

	updated, err := scanTemplate(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrTemplateNotFound
		}
		return nil, err
	}

	return updated, nil
}

// Delete - удаление шаблона (экземпляры удалятся каскадно)
func (r *TemplateRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM task_templates WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return taskdomain.ErrTemplateNotFound
	}

	return nil
}

// List - список всех шаблонов
func (r *TemplateRepository) List(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	query := `
		SELECT id, title, description, recurrence, created_at, updated_at
		FROM task_templates
		ORDER BY id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.TaskTemplate, 0)
	for rows.Next() {
		template, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

// GetAll - получить все шаблоны (для фоновой генерации)
func (r *TemplateRepository) GetAll(ctx context.Context) ([]taskdomain.TaskTemplate, error) {
	// Так же как List, но без сортировки
	query := `
		SELECT id, title, description, recurrence, created_at, updated_at
		FROM task_templates
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]taskdomain.TaskTemplate, 0)
	for rows.Next() {
		template, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

// scanTemplate - сканирует строку в TaskTemplate
func scanTemplate(row pgx.Row) (*taskdomain.TaskTemplate, error) {
	var template taskdomain.TaskTemplate
	var recurrenceJSON []byte

	err := row.Scan(
		&template.ID,
		&template.Title,
		&template.Description,
		&recurrenceJSON,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Десериализуем JSON в RecurrenceRule
	if len(recurrenceJSON) > 0 {
		var recurrence taskdomain.RecurrenceRule
		if err := json.Unmarshal(recurrenceJSON, &recurrence); err != nil {
			return nil, err
		}
		template.Recurrence = &recurrence
	}

	return &template, nil
}

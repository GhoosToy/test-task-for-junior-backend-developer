package task

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusSkipped    Status = "skipped" // НОВЫЙ СТАТУС
)

type Task struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// НОВЫЕ СТРУКТУРЫ
// Тип периодичности
type RecurrenceType string

const (
	RecurrenceDaily    RecurrenceType = "daily"    // ежедневно
	RecurrenceMonthly  RecurrenceType = "monthly"  // ежемесячно
	RecurrenceSpecific RecurrenceType = "specific" // конкретные даты
	RecurrenceParity   RecurrenceType = "parity"   // четные/нечетные
)

// Правило повторения
type RecurrenceRule struct {
	Type          RecurrenceType `json:"type"`
	Interval      *int           `json:"interval,omitempty"`       // для daily: через сколько дней
	DayOfMonth    *int           `json:"day_of_month,omitempty"`   // для monthly: число месяца 1-31
	SpecificDates []time.Time    `json:"specific_dates,omitempty"` // для specific: массив дат
	Parity        *string        `json:"parity,omitempty"`         // для parity: "even" или "odd"
	StartDate     time.Time      `json:"start_date"`               // с какой даты начинать
	EndDate       *time.Time     `json:"end_date,omitempty"`       // по какую дату (опционально)
}

// Шаблон задачи (для периодических задач)
type TaskTemplate struct {
	ID          int64           `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description" db:"description"`
	Recurrence  *RecurrenceRule `json:"recurrence,omitempty" db:"recurrence"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// Экземпляр задачи (конкретное вхождение)
type TaskInstance struct {
	ID          int64      `json:"id" db:"id"`
	TemplateID  int64      `json:"template_id" db:"template_id"`
	DueDate     time.Time  `json:"due_date" db:"due_date"`
	Status      Status     `json:"status" db:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Валидация для RecurrenceRule
func (r *RecurrenceRule) Validate() error {
	switch r.Type {
	case RecurrenceDaily:
		if r.Interval == nil || *r.Interval < 1 {
			return ErrInvalidRecurrenceRule
		}
	case RecurrenceMonthly:
		if r.DayOfMonth == nil || *r.DayOfMonth < 1 || *r.DayOfMonth > 31 {
			return ErrInvalidRecurrenceRule
		}
	case RecurrenceParity:
		if r.Parity == nil || (*r.Parity != "even" && *r.Parity != "odd") {
			return ErrInvalidRecurrenceRule
		}
	case RecurrenceSpecific:
		if len(r.SpecificDates) == 0 {
			return ErrInvalidRecurrenceRule
		}
	default:
		return ErrInvalidRecurrenceType
	}

	if r.StartDate.IsZero() {
		return ErrInvalidRecurrenceRule
	}

	if r.EndDate != nil && r.EndDate.Before(r.StartDate) {
		return ErrInvalidRecurrenceRule
	}

	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone, StatusSkipped:
		return true
	default:
		return false
	}
}

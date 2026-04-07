package task

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

// RecurrenceGenerator - генератор дат на основе правил повторения
type RecurrenceGenerator struct{}

// NewRecurrenceGenerator - конструктор
func NewRecurrenceGenerator() *RecurrenceGenerator {
	return &RecurrenceGenerator{}
}

// GenerateDates - генерирует даты для правила на период from-to
func (g *RecurrenceGenerator) GenerateDates(rule *taskdomain.RecurrenceRule, from, to time.Time) ([]time.Time, error) {
	if rule == nil {
		return nil, taskdomain.ErrNotRecurringTask
	}

	// Если from позже to - меняем местами
	if from.After(to) {
		from, to = to, from
	}

	// Если период заканчивается раньше start_date - пустой результат
	if to.Before(rule.StartDate) {
		return []time.Time{}, nil
	}

	// Корректируем from, если он раньше start_date
	if from.Before(rule.StartDate) {
		from = rule.StartDate
	}

	// Корректируем to, если есть end_date и он раньше
	if rule.EndDate != nil && to.After(*rule.EndDate) {
		to = *rule.EndDate
	}

	switch rule.Type {
	case taskdomain.RecurrenceDaily:
		return g.generateDaily(rule, from, to), nil
	case taskdomain.RecurrenceMonthly:
		return g.generateMonthly(rule, from, to), nil
	case taskdomain.RecurrenceSpecific:
		return g.generateSpecific(rule, from, to), nil
	case taskdomain.RecurrenceParity:
		return g.generateParity(rule, from, to), nil
	default:
		return nil, taskdomain.ErrInvalidRecurrenceType
	}
}

// generateDaily - каждые N дней
func (g *RecurrenceGenerator) generateDaily(rule *taskdomain.RecurrenceRule, from, to time.Time) []time.Time {
	var dates []time.Time

	interval := 1
	if rule.Interval != nil {
		interval = *rule.Interval
	}

	// Находим первую дату >= from
	current := from

	for !current.After(to) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, interval)
	}

	return dates
}

// generateMonthly - каждый месяц в определенный день
func (g *RecurrenceGenerator) generateMonthly(rule *taskdomain.RecurrenceRule, from, to time.Time) []time.Time {
	var dates []time.Time

	dayOfMonth := 1
	if rule.DayOfMonth != nil {
		dayOfMonth = *rule.DayOfMonth
	}

	// Начинаем с первого дня месяца
	current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())

	for !current.After(to) {
		// Получаем последний день месяца
		lastDay := time.Date(current.Year(), current.Month()+1, 0, 0, 0, 0, 0, current.Location()).Day()

		// Если запрошенный день превышает последний день месяца - берем последний
		targetDay := dayOfMonth
		if targetDay > lastDay {
			targetDay = lastDay
		}

		date := time.Date(current.Year(), current.Month(), targetDay, 0, 0, 0, 0, current.Location())

		if !date.Before(from) && !date.After(to) {
			dates = append(dates, date)
		}

		current = current.AddDate(0, 1, 0)
	}

	return dates
}

// generateSpecific - конкретные даты
func (g *RecurrenceGenerator) generateSpecific(rule *taskdomain.RecurrenceRule, from, to time.Time) []time.Time {
	var dates []time.Time

	for _, date := range rule.SpecificDates {
		if !date.Before(from) && !date.After(to) {
			dates = append(dates, date)
		}
	}

	return dates
}

// generateParity - четные/нечетные дни
func (g *RecurrenceGenerator) generateParity(rule *taskdomain.RecurrenceRule, from, to time.Time) []time.Time {
	var dates []time.Time

	isEven := true
	if rule.Parity != nil {
		isEven = *rule.Parity == "even"
	}

	current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	for !current.After(to) {
		day := current.Day()
		isDayEven := day%2 == 0

		if isEven == isDayEven {
			dates = append(dates, current)
		}

		current = current.AddDate(0, 0, 1)
	}

	return dates
}

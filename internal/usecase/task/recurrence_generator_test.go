package task

import (
	"testing"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func TestRecurrenceGenerator_GenerateDates_Daily(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: каждые 2 дня с 1 по 10 апреля
	rule := &taskdomain.RecurrenceRule{
		Type:      taskdomain.RecurrenceDaily,
		Interval:  intPtr(2),
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []time.Time{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}

	for i := range expected {
		if !dates[i].Equal(expected[i]) {
			t.Errorf("expected %v, got %v", expected[i], dates[i])
		}
	}
}

func TestRecurrenceGenerator_GenerateDates_Monthly(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: каждое 15-е число с января по март
	rule := &taskdomain.RecurrenceRule{
		Type:       taskdomain.RecurrenceMonthly,
		DayOfMonth: intPtr(15),
		StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []time.Time{
		time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}

	for i := range expected {
		if !dates[i].Equal(expected[i]) {
			t.Errorf("expected %v, got %v", expected[i], dates[i])
		}
	}
}

func TestRecurrenceGenerator_GenerateDates_Monthly_LastDay(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: 31-е число (должно работать даже в феврале)
	rule := &taskdomain.RecurrenceRule{
		Type:       taskdomain.RecurrenceMonthly,
		DayOfMonth: intPtr(31),
		StartDate:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Январь: 31, Февраль: 28 (последний день), Март: 31
	expected := []time.Time{
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}

	for i := range expected {
		if !dates[i].Equal(expected[i]) {
			t.Errorf("expected %v, got %v", expected[i], dates[i])
		}
	}
}

func TestRecurrenceGenerator_GenerateDates_Specific(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: конкретные даты
	specificDates := []time.Time{
		time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC),
	}

	rule := &taskdomain.RecurrenceRule{
		Type:          taskdomain.RecurrenceSpecific,
		SpecificDates: specificDates,
		StartDate:     time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dates) != len(specificDates) {
		t.Errorf("expected %d dates, got %d", len(specificDates), len(dates))
	}

	for i := range specificDates {
		if !dates[i].Equal(specificDates[i]) {
			t.Errorf("expected %v, got %v", specificDates[i], dates[i])
		}
	}
}

func TestRecurrenceGenerator_GenerateDates_Parity_Even(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: четные дни апреля
	rule := &taskdomain.RecurrenceRule{
		Type:      taskdomain.RecurrenceParity,
		Parity:    stringPtr("even"),
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []time.Time{
		time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}

	for i := range expected {
		if !dates[i].Equal(expected[i]) {
			t.Errorf("expected %v, got %v", expected[i], dates[i])
		}
	}
}

func TestRecurrenceGenerator_GenerateDates_WithStartDate(t *testing.T) {
	generator := &RecurrenceGenerator{}

	// Тест: start_date позже than from
	rule := &taskdomain.RecurrenceRule{
		Type:      taskdomain.RecurrenceDaily,
		Interval:  intPtr(1),
		StartDate: time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
	}

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Должно начаться с 5 апреля, а не с 1
	expected := []time.Time{
		time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}
}

func TestRecurrenceGenerator_GenerateDates_WithEndDate(t *testing.T) {
	generator := &RecurrenceGenerator{}

	endDate := time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC)

	rule := &taskdomain.RecurrenceRule{
		Type:      taskdomain.RecurrenceDaily,
		Interval:  intPtr(1),
		StartDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   &endDate,
	}

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	dates, err := generator.GenerateDates(rule, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Должно остановиться на 7 апреля
	expected := []time.Time{
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 7, 0, 0, 0, 0, time.UTC),
	}

	if len(dates) != len(expected) {
		t.Errorf("expected %d dates, got %d", len(expected), len(dates))
	}
}

// Вспомогательные функции
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

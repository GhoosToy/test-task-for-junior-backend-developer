package task

import "errors"

var ErrNotFound = errors.New("task not found")

var (
	ErrInvalidRecurrenceType = errors.New("invalid recurrence type")
	ErrInvalidRecurrenceRule = errors.New("invalid recurrence rule")
	ErrNotRecurringTask      = errors.New("task is not recurring")
	ErrTemplateNotFound      = errors.New("task template not found")
	ErrInstanceNotFound      = errors.New("task instance not found")
	ErrAlreadyCompleted      = errors.New("task instance already completed")
	ErrAlreadySkipped        = errors.New("task instance already skipped")
	ErrInvalidInput          = errors.New("invalid input")
)

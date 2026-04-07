package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httphandlers "example.com/taskservice/internal/transport/http/handlers"
)

func NewRouter(
	taskHandler *httphandlers.TaskHandler,
	recurringHandler *httphandlers.RecurringHandler, // НОВЫЙ параметр
	docsHandler *swaggerdocs.Handler,
) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9]+}", taskHandler.Delete).Methods(http.MethodDelete)

	// ========== НОВЫЕ МАРШРУТЫ ДЛЯ ПЕРИОДИЧЕСКИХ ЗАДАЧ ==========

	// Получить задачи на конкретную дату
	// GET /api/v1/tasks/recurring?date=2026-04-07
	api.HandleFunc("/tasks/recurring", recurringHandler.GetTasksForDate).Methods(http.MethodGet)

	// Шаблоны периодических задач
	api.HandleFunc("/templates", recurringHandler.CreateTemplate).Methods(http.MethodPost)
	api.HandleFunc("/templates", recurringHandler.ListTemplates).Methods(http.MethodGet)
	api.HandleFunc("/templates/{id:[0-9]+}", recurringHandler.GetTemplateByID).Methods(http.MethodGet)
	api.HandleFunc("/templates/{id:[0-9]+}", recurringHandler.UpdateTemplate).Methods(http.MethodPut)
	api.HandleFunc("/templates/{id:[0-9]+}", recurringHandler.DeleteTemplate).Methods(http.MethodDelete)

	// Получить даты вхождений для шаблона
	// GET /api/v1/templates/{id}/occurrences?from=2026-04-01&to=2026-04-30
	api.HandleFunc("/templates/{id:[0-9]+}/occurrences", recurringHandler.GetOccurrences).Methods(http.MethodGet)

	// Экземпляры задач (конкретные вхождения)
	api.HandleFunc("/instances/{id:[0-9]+}/complete", recurringHandler.CompleteInstance).Methods(http.MethodPatch)
	api.HandleFunc("/instances/{id:[0-9]+}/skip", recurringHandler.SkipInstance).Methods(http.MethodPatch)

	return router
}

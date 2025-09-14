package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Aidajy111/go-final-project-main/pkg/db"
)

func Init() {
	http.HandleFunc("/api/nextdate", NextDateHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", doneTaskHandler)
	http.HandleFunc("/api/tasks", GetTasks)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Println(r.Body)
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
			return
		}
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, taskResponse{Error: "method not allowed"})
	}
}

func GetTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tasksHandler(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, taskResponse{Error: "method not allowed"})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error, status int) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

// Обработчик для GET /api/task?id=<id>
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// Обработчик для PUT /api/task
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
		return
	}

	// Валидация названия задачи
	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Название задачи не может быть пустым"})
		return
	}

	// Валидация даты (должна быть в формате YYYYMMDD и быть корректной датой)
	if task.Date != "" {
		if len(task.Date) != 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты"})
			return
		}

		// Проверка что дата корректна (аналогично NextDate)
		_, err := time.Parse(dateFormat, task.Date)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты"})
			return
		}
	}

	// Валидация правила повторения (если указано)
	if task.Repeat != "" {
		// Проверяем что правило повторения валидно
		if !isValidRepeatRule(task.Repeat) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат правила повторения"})
			return
		}
	}

	// Обновляем задачу в базе данных
	err = db.UpdateTask(&task)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	// Возвращаем пустой JSON при успешном обновлении
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}
func isValidRepeatRule(repeat string) bool {
	if repeat == "" {
		return true
	}

	// Проверяем базовые форматы: d X, y X, w X, m X
	if len(repeat) < 2 {
		return false
	}

	// Простая проверка - более сложная логика должна быть в NextDate
	switch repeat[0] {
	case 'd', 'y', 'w', 'm':
		// Должен быть пробел и число после него
		if len(repeat) < 3 || repeat[1] != ' ' {
			return false
		}
		// Проверяем что после пробела число
		for _, char := range repeat[2:] {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	default:
		return false
	}
}

type taskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			writeJSON(w, http.StatusInternalServerError, taskResponse{Error: "internal server error"})
		}
	}()

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, taskResponse{Error: "method not allowed"})
		return
	}

	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "invalid JSON format"})
		return
	}

	if task.Title == "" {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "title is required"})
		return
	}

	now := time.Now()
	currentDate := now.Format(dateFormat)

	// Если дата не указана — ставим сегодняшнюю
	if task.Date == "" {
		task.Date = currentDate
	}

	// Валидация формата даты
	_, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "invalid date format"})
		return
	}

	// Если правило повторения указано — просто валидируем его,
	// но НЕ меняем task.Date на nextDate.
	if task.Repeat != "" {
		if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
			writeJSON(w, http.StatusBadRequest, taskResponse{Error: err.Error()})
			return
		}
	} else {
		// Для одноразовой задачи: если дата в прошлом — поставить today
		taskTime, _ := time.Parse(dateFormat, task.Date)
		if taskTime.Before(now) {
			task.Date = currentDate
		}
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, taskResponse{Error: "database error"})
		return
	}

	writeJSON(w, http.StatusOK, taskResponse{ID: id})
}

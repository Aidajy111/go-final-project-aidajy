package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	r := strings.Fields(strings.TrimSpace(repeat))
	if len(r) == 0 {
		return true
	}
	if r[0] == "y" {
		if len(r) == 1 {
			return true
		}
		if len(r) == 2 {
			n, err := strconv.Atoi(r[1])
			return err == nil && n > 0
		}
		return false
	}
	if len(r) == 2 && (r[0] == "d" || r[0] == "w" || r[0] == "m") {
		n, err := strconv.Atoi(r[1])
		if r[0] == "d" && (err != nil || n <= 0 || n > 400) {
			return false
		}
		return err == nil && n > 0
	}
	return false
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
	todayStr := now.In(time.Local).Format(dateFormat)
	if task.Date == "" {
		task.Date = todayStr
	}

	// валидируем формат даты
	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "invalid date format"})
		return
	}
	// нормализуем "сегодня" к полуночи по локали
	today, _ := time.Parse(dateFormat, todayStr)

	if t.Before(today) {
		// дата в прошлом
		if strings.TrimSpace(task.Repeat) == "" {
			task.Date = todayStr
		} else {
			next, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeJSON(w, http.StatusOK, taskResponse{Error: err.Error()})
				return
			}
			task.Date = next
		}
	} else {
		// дата не в прошлом — просто валидируем правило, если оно указано
		if strings.TrimSpace(task.Repeat) != "" {
			if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
				writeJSON(w, http.StatusOK, taskResponse{Error: err.Error()})
				return
			}
		}
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, taskResponse{Error: "database error"})
		return
	}

	writeJSON(w, http.StatusOK, taskResponse{ID: id})
}

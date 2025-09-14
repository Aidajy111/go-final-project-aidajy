package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Aidajy111/go-final-project-main/pkg/db"
)

func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("REQ %s %s query=%v\n", r.Method, r.URL.Path, r.URL.RawQuery)
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, taskResponse{Error: "id is required"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, taskResponse{Error: "задача не найдена"})
		return
	}

	// Если одноразовая — удаляем
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, http.StatusInternalServerError, taskResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{})
		return
	}

	// Повторяющаяся задача — парсим дату и решаем, переносить ли
	taskDate, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, taskResponse{Error: "invalid task date"})
		return
	}

	now := time.Now()
	// Если дата в прошлом или равна сегодня, считаем следующую:
	if !taskDate.After(now) {
		nextDate, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, taskResponse{Error: err.Error()})
			return
		}
		if err := db.UpdateDate(nextDate, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, taskResponse{Error: err.Error()})
			return
		}
		// успешно
		writeJSON(w, http.StatusOK, map[string]string{})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

package api

import (
	"net/http"

	"github.com/Aidajy111/go-final-project-main/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.Tasks(50)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, TasksResp{Tasks: tasks})
}

package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const dateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("empty repeat rule")
	}

	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return "", errors.New("invalid date format")
	}

	parts := strings.Fields(repeat)
	if len(parts) != 2 {
		return "", errors.New("invalid repeat format")
	}

	unit := parts[0]
	n, err := strconv.Atoi(parts[1])
	if err != nil || n <= 0 {
		return "", errors.New("invalid repeat number")
	}

	result := date

	switch unit {
	case "d":
		// прибавляем n дней пока result <= now
		for !result.After(now) {
			result = result.AddDate(0, 0, n)
		}
	case "w":
		for !result.After(now) {
			result = result.AddDate(0, 0, 7*n)
		}
	case "m":
		for !result.After(now) {
			result = result.AddDate(0, n, 0)
		}
	case "y":
		for !result.After(now) {
			result = result.AddDate(n, 0, 0)
		}
	default:
		return "", errors.New("unsupported repeat unit")
	}

	return result.Format(dateFormat), nil
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Если now не указан, используем текущую дату
	var now time.Time
	if nowStr == "" {
		now = time.Now()
	} else {
		var err error
		now, err = time.Parse(dateFormat, nowStr)
		if err != nil {
			http.Error(w, "invalid now parameter", http.StatusBadRequest)
			return
		}
	}

	// Вычисляем следующую дату
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем результат
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

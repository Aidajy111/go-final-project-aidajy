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

	date, err := time.Parse(dateFormat, strings.TrimSpace(dstart))
	if err != nil {
		return "", errors.New("invalid date format")
	}

	parts := strings.Fields(strings.TrimSpace(repeat))
	if len(parts) == 0 {
		return "", errors.New("invalid repeat format")
	}

	unit := parts[0]
	n := 0

	switch unit {
	case "y":
		if len(parts) == 1 {
			next := addYearKeepingFeb29(date)
			for !next.After(now) {
				next = addYearKeepingFeb29(next)
			}
			return next.Format(dateFormat), nil
		}
		if len(parts) != 2 {
			return "", errors.New("invalid repeat format")
		}
		n, err = strconv.Atoi(parts[1])
		if err != nil || n <= 0 {
			return "", errors.New("invalid years number")
		}
		next := date
		for i := 0; i < n; i++ {
			next = addYearKeepingFeb29(next)
		}
		for !next.After(now) {
			for i := 0; i < n; i++ {
				next = addYearKeepingFeb29(next)
			}
		}
		return next.Format(dateFormat), nil

	case "d":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat format")
		}
		n, err = strconv.Atoi(parts[1])
		if err != nil {
			return "", errors.New("invalid repeat number")
		}
		if n <= 0 || n > 400 {
			return "", errors.New("days must be between 1 and 400")
		}
		// первый обязательный шаг
		next := date.AddDate(0, 0, n)
		// затем догоняем now
		for !next.After(now) {
			next = next.AddDate(0, 0, n)
		}
		return next.Format(dateFormat), nil

	case "w":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat format")
		}
		weeks, err := strconv.Atoi(parts[1])
		if err != nil || weeks <= 0 {
			return "", errors.New("invalid repeat number")
		}
		next := date.AddDate(0, 0, 7*weeks)
		for !next.After(now) {
			next = next.AddDate(0, 0, 7*weeks)
		}
		return next.Format(dateFormat), nil

	case "m":
		if len(parts) != 2 {
			return "", errors.New("invalid repeat format")
		}
		months, err := strconv.Atoi(parts[1])
		if err != nil || months <= 0 {
			return "", errors.New("invalid repeat number")
		}
		next := date.AddDate(0, months, 0)
		for !next.After(now) {
			next = next.AddDate(0, months, 0)
		}
		return next.Format(dateFormat), nil
	}

	return "", errors.New("unsupported repeat unit")
}

func addYearKeepingFeb29(t time.Time) time.Time {
	if t.Month() == time.February && t.Day() == 29 {
		ny := t.AddDate(1, 0, 0)
		if ny.Month() == time.February && ny.Day() == 28 {
			return ny.AddDate(0, 0, 1)
		}
		return ny
	}
	return t.AddDate(1, 0, 0)
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

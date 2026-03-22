package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"
)

func (app *application) listCourses(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, code, title, credits, enrolled FROM courses ORDER BY id`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var courses []Course
	for rows.Next() {
		var c Course
		err := rows.Scan(&c.ID, &c.Code, &c.Title, &c.Credits, &c.Enrolled)
		if err != nil {
			app.serverError(w, err)
			return
		}
		courses = append(courses, c)
	}
	if err = rows.Err(); err != nil {
		app.serverError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"courses": courses}, nil)
}

func (app *application) getCourse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `SELECT id, code, title, credits, enrolled FROM courses WHERE id = $1`

	var c Course
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Code, &c.Title, &c.Credits, &c.Enrolled)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFound(w)
		default:
			app.serverError(w, err)
		}
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"course": c}, nil)
}

func (app *application) createCourse(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code     string `json:"code"`
		Title    string `json:"title"`
		Credits  int    `json:"credits"`
		Enrolled int    `json:"enrolled"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, err.Error())
		return
	}

	v := newValidator()
	v.Check(input.Code != "", "code", "must be provided")
	v.Check(input.Title != "", "title", "must be provided")
	v.Check(input.Credits > 0, "credits", "must be greater than 0")

	if !v.Valid() {
		app.failedValidation(w, v.Errors)
		return
	}

	query := `INSERT INTO courses (code, title, credits, enrolled) VALUES ($1, $2, $3, $4) RETURNING id`

	var newID int64
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, input.Code, input.Title, input.Credits, input.Enrolled).Scan(&newID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	newCourse := Course{ID: newID, Code: input.Code, Title: input.Title, Credits: input.Credits, Enrolled: input.Enrolled}
	extra := http.Header{"Location": []string{"/courses/" + strconv.FormatInt(newID, 10)}}
	app.writeJSON(w, http.StatusCreated, envelope{"course": newCourse}, extra)
}

func (app *application) deleteCourse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `DELETE FROM courses WHERE id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.serverError(w, err)
		return
	}
	if rowsAffected == 0 {
		app.notFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

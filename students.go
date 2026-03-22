package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (app *application) listStudents(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, programme, year FROM students ORDER BY id`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := app.db.QueryContext(ctx, query)
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		err := rows.Scan(&s.ID, &s.Name, &s.Programme, &s.Year)
		if err != nil {
			app.serverError(w, err)
			return
		}
		students = append(students, s)
	}
	if err = rows.Err(); err != nil {
		app.serverError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"students": students}, nil)
}

func (app *application) getStudent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `SELECT id, name, programme, year FROM students WHERE id = $1`

	var s Student
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.Name, &s.Programme, &s.Year)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFound(w)
		default:
			app.serverError(w, err)
		}
		return
	}

	extra := http.Header{"X-Resource-Id": []string{strconv.FormatInt(id, 10)}}
	app.writeJSON(w, http.StatusOK, envelope{"student": s}, extra)
}

func (app *application) createStudent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string `json:"name"`
		Programme string `json:"programme"`
		Year      int    `json:"year"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, err.Error())
		return
	}

	v := newValidator()
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(len(input.Name) <= 100, "name", "must not exceed 100 characters")
	v.Check(input.Programme != "", "programme", "must be provided")
	v.Check(between(input.Year, 1, 4), "year", "must be between 1 and 4")

	if !v.Valid() {
		app.failedValidation(w, v.Errors)
		return
	}

	query := `INSERT INTO students (name, programme, year) VALUES ($1, $2, $3) RETURNING id`

	var newID int64
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, input.Name, input.Programme, input.Year).Scan(&newID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	newStudent := Student{ID: newID, Name: input.Name, Programme: input.Programme, Year: input.Year}
	extra := http.Header{"Location": []string{"/students/" + strconv.FormatInt(newID, 10)}}
	app.writeJSON(w, http.StatusCreated, envelope{"student": newStudent}, extra)
}

func (app *application) updateStudent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	var input struct {
		Name      string `json:"name"`
		Programme string `json:"programme"`
		Year      int    `json:"year"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, err.Error())
		return
	}

	v := newValidator()
	v.Check(input.Name != "", "name", "must be provided")
	v.Check(len(input.Name) <= 100, "name", "must not exceed 100 characters")
	v.Check(input.Programme != "", "programme", "must be provided")
	v.Check(between(input.Year, 1, 4), "year", "must be between 1 and 4")

	if !v.Valid() {
		app.failedValidation(w, v.Errors)
		return
	}

	query := `UPDATE students SET name = $1, programme = $2, year = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, input.Name, input.Programme, input.Year, id)
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

	updated := Student{ID: id, Name: input.Name, Programme: input.Programme, Year: input.Year}
	app.writeJSON(w, http.StatusOK, envelope{"student": updated}, nil)
}

func (app *application) deleteStudent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	query := `DELETE FROM students WHERE id = $1`

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

// ─── Error helpers ────────────────────────────────────────────────────────────

func (app *application) serverError(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %v", err)
	app.writeJSON(w, http.StatusInternalServerError, envelope{
		"error": "the server encountered a problem and could not process your request",
	}, nil)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.writeJSON(w, http.StatusNotFound, envelope{
		"error": "the requested resource could not be found",
	}, nil)
}

func (app *application) badRequest(w http.ResponseWriter, msg string) {
	app.writeJSON(w, http.StatusBadRequest, envelope{"error": msg}, nil)
}

func (app *application) failedValidation(w http.ResponseWriter, errors map[string]string) {
	app.writeJSON(w, http.StatusUnprocessableEntity, envelope{"errors": errors}, nil)
}

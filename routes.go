package main

import "net/http"

func serve(app *application) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /students", app.listStudents)
	mux.HandleFunc("GET /students/{id}", app.getStudent)
	mux.HandleFunc("POST /students", app.createStudent)
	mux.HandleFunc("PUT /students/{id}", app.updateStudent)
	mux.HandleFunc("DELETE /students/{id}", app.deleteStudent)

	mux.HandleFunc("GET /courses", app.listCourses)
	mux.HandleFunc("GET /courses/{id}", app.getCourse)
	mux.HandleFunc("POST /courses", app.createCourse)
	mux.HandleFunc("DELETE /courses/{id}", app.deleteCourse)

	mux.HandleFunc("GET /health", app.health)

	return http.ListenAndServe(":4000", mux)
}

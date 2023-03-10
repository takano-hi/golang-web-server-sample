package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/midwhite/golang-web-server-sample/todo-api/db"
	"github.com/midwhite/golang-web-server-sample/todo-api/models"
	"github.com/midwhite/golang-web-server-sample/todo-api/serializers"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func HandleTodos(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		createTodo(w, req)
	case "GET", "":
		getTodoList(w, req)
	default:
		data := serializers.ErrorResponse{Message: "Not Found"}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	}
}

func HandleTodoDetail(w http.ResponseWriter, req *http.Request) {
	pathname := strings.TrimPrefix(req.URL.Path, "/todos/")
	paths := regexp.MustCompile("[/?]").Split(pathname, -1)
	todoId := paths[0]

	if todoId == "" {
		data := serializers.ErrorResponse{Message: "ID is not set."}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
		return
	}

	switch req.Method {
	case "GET", "":
		getTodoDetail(w, req, todoId)
	case "PUT":
		updateTodo(w, req, todoId)
	case "DELETE":
		deleteTodo(w, req, todoId)
	default:
		data := serializers.ErrorResponse{Message: "No route matches."}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	}
}

type CreateTodoParams struct {
	Title string `json:"title" validate:"required"`
}

func createTodo(w http.ResponseWriter, req *http.Request) {
	reqBody, _ := io.ReadAll(req.Body)
	params := new(CreateTodoParams)
	json.Unmarshal(reqBody, params)

	validation := validator.New()
	err := validation.Struct(params)

	if err != nil {
		result := make([]string, 0)
		for _, validationError := range err.(validator.ValidationErrors) {
			result = append(result, serializers.TranslateErrorMessage(validationError, "Todo"))
		}
		message := strings.Join(result, "\n")

		data := serializers.ErrorResponse{Message: message}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusBadRequest)
		w.Write(body)
	} else {
		todo := models.Todo{Title: params.Title, CreatedAt: time.Now()}
		todo.Insert(req.Context(), db.Conn, boil.Infer())
		body, _ := json.Marshal(todo)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	}
}

type GetTodoListResponse struct {
	Todos []*models.Todo `json:"todos"`
}

func getTodoList(w http.ResponseWriter, req *http.Request) {
	todos, _ := models.Todos(qm.OrderBy("created_at")).All(req.Context(), db.Conn)
	if todos == nil {
		todos = make([]*models.Todo, 0)
	}
	response := GetTodoListResponse{Todos: todos}
	body, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func getTodoDetail(w http.ResponseWriter, req *http.Request, todoId string) {
	todo, err := models.FindTodo(req.Context(), db.Conn, todoId, "id", "title")

	if err != nil {
		data := serializers.ErrorResponse{Message: "todo is not found."}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	} else {
		body, _ := json.Marshal(todo)

		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}
}

type UpdateTodoParams struct {
	Title string `json:"title" validate:"required"`
}

func updateTodo(w http.ResponseWriter, req *http.Request, todoId string) {
	ctx := req.Context()
	todo, err := models.FindTodo(ctx, db.Conn, todoId, "id", "title", "created_at")

	if err != nil {
		data := serializers.ErrorResponse{Message: "todo is not found."}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	} else {
		reqBody, _ := io.ReadAll(req.Body)

		params := new(UpdateTodoParams)
		json.Unmarshal(reqBody, params)

		validation := validator.New()
		err := validation.Struct(params)

		if err != nil {
			errorMessages := make([]string, 0)
			for _, fieldError := range err.(validator.ValidationErrors) {
				errorMessages = append(errorMessages, serializers.TranslateErrorMessage(fieldError, "Todo"))
			}
			message := strings.Join(errorMessages, "\n")

			data := serializers.ErrorResponse{Message: message}
			body, _ := json.Marshal(data)

			w.WriteHeader(http.StatusBadRequest)
			w.Write(body)
		} else {
			todo.Title = params.Title
			todo.Update(ctx, db.Conn, boil.Infer())

			body, _ := json.Marshal(todo)
			w.Write(body)
		}
	}
}

type DeleteTodoResponse struct {
	Success bool `json:"success"`
}

func deleteTodo(w http.ResponseWriter, req *http.Request, todoId string) {
	ctx := req.Context()
	todo, err := models.FindTodo(ctx, db.Conn, todoId, "id")

	if err != nil {
		data := serializers.ErrorResponse{Message: "todo is not found."}
		body, _ := json.Marshal(data)

		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	} else {
		todo.Delete(ctx, db.Conn)

		data := DeleteTodoResponse{Success: true}
		body, _ := json.Marshal(data)

		w.Write(body)
	}
}

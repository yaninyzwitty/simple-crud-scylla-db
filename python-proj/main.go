package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gocql/gocql"
)

var session *gocql.Session

type Todo struct {
	ID     gocql.UUID `json:"id"`
	Title  string     `json:"title"`
	Status string     `json:"status"`
}

// CREATE TABLE IF NOT EXISTS todos ( id uuid, title text, status text, PRIMARY KEY (id) );
func main() {

	var err error
	cluster := gocql.NewCluster("node-0.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud", "node-1.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud", "node-2.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud")
	cluster.Authenticator = gocql.PasswordAuthenticator{Username: "scylla", Password: "t5IJAg9zdGYB6NL"}
	cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy("AWS_EU_CENTRAL_1")

	cluster.Keyspace = "mykeyspace"
	session, err = cluster.CreateSession()

	if err != nil {
		log.Fatalf("Error creating session: %v", err)
	}

	defer session.Close()

	// initialize http router (chi / gin)

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	// Define routes
	router.Get("/todos", ListTodos)
	router.Get("/todos/{id}", GetTodo)
	router.Post("/todos", CreateTodo)
	router.Patch("/todos/{id}", UpdateTodo)
	router.Delete("/todos/{id}", DeleteTodo)

	// Start the server
	log.Println("Server started on :3000 😂")
	log.Fatal(http.ListenAndServe(":3000", router))

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func ListTodos(w http.ResponseWriter, r *http.Request) {
	var todos []Todo
	iter := session.Query("SELECT id, title, status FROM todos").Iter()

	for {
		var todo Todo
		if !iter.Scan(&todo.ID, &todo.Title, &todo.Status) {

			break //alreay a for loop
		}
		todos = append(todos, todo)
	}
	if err := iter.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	respondWithJSON(w, http.StatusOK, todos)

}

func GetTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var todo Todo
	if err := session.Query("SELECT id, title, status FROM todos WHERE id = ?", id).Scan(&todo.ID, &todo.Title, &todo.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, todo)

}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	todo.ID = gocql.TimeUUID()
	if err := session.Query("INSERT INTO todos (id, title, status) VALUES (?, ?, ?)", todo.ID, todo.Title, todo.Status).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusCreated, todo)

}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := session.Query("UPDATE todos SET title=?, status=? WHERE id=?", todo.Title, todo.Status, id).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, todo)

}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := session.Query("DELETE FROM todos WHERE id = ?", id).Exec(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusNoContent)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Deleted todo successfully"))

}

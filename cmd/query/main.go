package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "mydb"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчики HTTP-запросов
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name parameter is required", http.StatusBadRequest)
		return
	}

	user, err := h.dbProvider.SelectUser(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, " + user + "!"))
}

func (h *Handlers) PostUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.dbProvider.InsertUser(input.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		fmt.Println("Вы лоххх")
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Запись добавлена!"))
}

func (dp *DatabaseProvider) SelectUser(name string) (string, error) {
	var user string

	row := dp.db.QueryRow("SELECT name FROM mytable WHERE name = $1", name)
	err := row.Scan(&user)
	if err != nil {
		return "", err
	}

	return user, nil
}

func (dp *DatabaseProvider) InsertUser(name string) error {
	_, err := dp.db.Exec("INSERT INTO mytable (name) VALUES ($1)", name)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	address := flag.String("address", "127.0.0.1:8080", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
		fmt.Println("Вы лох")
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	http.HandleFunc("/api/user", h.GetUser)
	http.HandleFunc("/api/user/create", h.PostUser)

	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}

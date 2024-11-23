package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (dp *DatabaseProvider) GetCounter() (int, error) {
	var counter int
	row := dp.db.QueryRow("SELECT value FROM counter_table LIMIT 1")
	err := row.Scan(&counter)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func (dp *DatabaseProvider) UpdateCounter(value int) error {
	_, err := dp.db.Exec("UPDATE counter_table SET value = value + $1", value)
	return err
}

func (h *Handlers) HandleCount(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Получаем текущее значение счетчика
		counter, err := h.dbProvider.GetCounter()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Увеличиваем счетчик на 1
		err = h.dbProvider.UpdateCounter(1)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Возвращаем текущее значение счетчика после увеличения
		fmt.Fprintf(w, "%d", counter+1) // Увеличиваем на 1 для ответа
	case "POST":
		count, err := strconv.Atoi(r.FormValue("count"))
		if err != nil {
			http.Error(w, "это не число", http.StatusBadRequest)
			return
		}
		err = h.dbProvider.UpdateCounter(count)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Success")
	default:
		http.Error(w, "Неизвестный метод", http.StatusMethodNotAllowed)
	}
}

func main() {
	address := flag.String("address", "127.0.0.1:3333", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	http.HandleFunc("/count", h.HandleCount)

	fmt.Println("Сервер запущен на порту :3333")
	err = http.ListenAndServe(*address, nil)
	if err != nil {
		log.Fatal(err)
	}
}

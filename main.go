package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// keyValuePutHandler ожидает получить PUT-запрос с ресурсом
// "/v1/key/{key}".

func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Получить ключ из запроса
	key := vars["key"]

	value, err := io.ReadAll(r.Body) // Тело запроса хранит значение
	defer r.Body.Close()

	if err != nil { // Если возникла ошибка сообщить о ней!
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = Put(key, string(value))
	if err != nil {
		http.Error(w,
			err.Error(),
			http.StatusInternalServerError)
		return
	}

	fmt.Print(key)
	fmt.Print(value)
	w.WriteHeader(http.StatusCreated)
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Извлечь ключ из запроса
	key := vars["key"]

	value, err := Get(key) // Получить значение для данного ключа
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(value))
}

func keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Извлечь ключ из запроса
	key := vars["key"]

	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func testovich(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Testovich</h1>"))
}

func main() {
	r := mux.NewRouter()

	// зарегистрировать keyValuePutHandler как обработчик HTTP-запросов PUT,
	// в которых указан путь "/v1/{key}"
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", keyValueDeleteHandler).Methods("DELETE")
	r.HandleFunc("/testovich", testovich).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}

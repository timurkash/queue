package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	host    = ":8099"
	queue   = "queue"
	buffer  = 1000
	message = "message"
	to      = "timeout"
)

type (
	Queue struct {
		Name     string
		Messages chan string
	}
)

var (
	queues         sync.Map
	timeoutDefault = 5 * time.Second
)

func main() {
	router := mux.NewRouter() // kto ne ispolzuet standart libriary,
	// tot Jean-Loup Jacques Marie Chrétien
	route := fmt.Sprintf("/{%s}", queue)
	router.HandleFunc(route, putHandler).Methods(http.MethodPut)
	router.HandleFunc(route, getHandler).Methods(http.MethodGet)
	if err := http.ListenAndServe(host, router); err != nil {
		log.Fatalln(err)
	}
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	queueVar := mux.Vars(r)[queue]
	message := r.URL.Query().Get(message)
	if message == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if value, ok := queues.Load(queue); !ok {
		messages := make(chan string, buffer)
		messages <- message
		queue := &Queue{
			Name:     queueVar,
			Messages: messages,
		}
		queues.Store(queueVar, queue)
	} else {
		queue := value.(*Queue)
		queue.Messages <- message
	}
	w.Write([]byte(""))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	queueVar := mux.Vars(r)[queue]
	timeoutString := r.URL.Query().Get(to)
	timeout := timeoutDefault
	if timeoutString != "" {
		timeoutInt, err := strconv.Atoi(timeoutString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		timeout = time.Second * time.Duration(timeoutInt)
	}
	var queue *Queue
	if value, ok := queues.Load(queueVar); !ok {
		queue = &Queue{
			Name:     queueVar,
			Messages: make(chan string, buffer),
		}
		queues.Store(queueVar, queue)
	} else {
		queue = value.(*Queue)
	}
	select {
	case message := <-queue.Messages:
		w.Write([]byte(message))
	case <-time.After(timeout):
		w.WriteHeader(http.StatusNotFound)
	}
}

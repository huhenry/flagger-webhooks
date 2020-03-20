package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// VERSION semantic versioning format
const VERSION = "0.1.0"

var healthy int32

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := registerRoutes()

	log.Println("Server is starting...")
	//r := registerRoutes()
	// Add your routes as needed

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Fatal(srv.ListenAndServe())

		log.Println("Server is starting at 0.0.0.0:8080")
		atomic.StoreInt32(&healthy, 1)

	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// health
	atomic.StoreInt32(&healthy, 0)

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("Server is shutting down...")
	os.Exit(0)
}

func registerRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz)

	r.HandleFunc("/event", handlerEvent).Methods("Post")

	return r
}

func handlerEvent(w http.ResponseWriter, req *http.Request) {

	defer func() {
		err := req.Body.Close()

		if err != nil {
			log.Errorf("req body close err : %s", err.Error())
			return
		}
	}()

	vars := mux.Vars(req)
	w.WriteHeader(http.StatusOK)
	log.Println(vars)

	var event Event
	if req.Body == nil {
		log.Errorln("request body is empty ")
		return
	}
	err := json.NewDecoder(req.Body).Decode(&event)
	if err != nil {
		log.Errorf("Json decoder err: %s", err)
		return
	}

	log.Println(event)
}

func healthz(w http.ResponseWriter, req *http.Request) {
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

/*
Event message body
{
  "name": "string (canary name)",
  "namespace": "string (canary namespace)",
  "phase": "string (canary phase)",
  "metadata": {
    "eventMessage": "string (canary event message)",
    "eventType": "string (canary event type)",
    "timestamp": "string (unix timestamp ms)"
  }
}
*/
type Event struct {
	Name      string   `json:"name"`
	Namespace string   `json:"namespace"`
	Phase     string   `json:"phase"`
	MetaData  MetaData `json:"metadata"`
}

type MetaData struct {
	EventMessage string `json:"eventMessage"`
	EventType    string `json:"eventType"`
	Timestamp    string `json:"timestamp"`
}

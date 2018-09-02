package http

import (
	"encoding/json"
	"github.com/jvdanker/mug/store"
	"log"
	"net/http"
	"os"
)

type Decorator func(http.HandlerFunc) http.HandlerFunc
type JsonHandler func(*http.Request) (interface{}, error)

func Handle(pattern string, handler JsonHandler) {
	var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

	http.HandleFunc(pattern, Decorate(
		handler,
		WithJsonHandler(),
		WithLogger(logger),
		WithCors()))
}

func Decorate(h JsonHandler, decorators ...Decorator) http.HandlerFunc {
	var handler = func(w http.ResponseWriter, r *http.Request) {
		data, err := h(r)
		if err != nil {
			he, ok := err.(store.HandlerError)
			if ok {
				http.Error(w, err.Error(), he.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if data == nil {
			var s struct{}
			data = s
		}

		j, err := json.Marshal(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}

	for _, d := range decorators {
		handler = d(handler)
	}

	return handler
}

func WithJsonHandler() Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			h.ServeHTTP(w, r)
		})
	}
}

func WithLogger(l *log.Logger) Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Println(r.Method, r.URL.Path)
			//req, _ := httputil.DumpRequest(r, true)
			//fmt.Printf("%s\n", string(req))

			h.ServeHTTP(w, r)
		})
	}
}

func WithCors() Decorator {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

			if r.Method == "OPTIONS" {
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

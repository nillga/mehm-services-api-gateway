package router

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type ApiGatewayRouter interface {
	GET(uri string, f func(w http.ResponseWriter, r *http.Request))
	POST(uri string, f func(w http.ResponseWriter, r *http.Request))
	SERVE(port string)
}

type muxRouter struct{}

var router = mux.NewRouter()

func NewApiGatewayRouter() ApiGatewayRouter {
	return &muxRouter{}
}

func (m *muxRouter) GET(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	router.HandleFunc(uri, f).Methods("GET").Schemes("http")
}

func (m *muxRouter) POST(uri string, f func(w http.ResponseWriter, r *http.Request)) {
	router.HandleFunc(uri, f).Methods("POST").Schemes("http")
}

func (m *muxRouter) SERVE(port string) {
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Credentials", "Cookie"},
	})
	l := log.Logger{}
	l.SetOutput(os.Stdout)
	c.Log = &l

	log.Fatalln(http.ListenAndServe(":"+port, c.Handler(router)))
}

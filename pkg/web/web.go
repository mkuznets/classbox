package web

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/opts"
	"log"
	"net/http"
	"time"
)

type Web struct {
	API       *client.Client
	DocsURL   string
	WebURL    string
	Templates *Templates
}

type Server struct {
	Addr   string
	Sentry *opts.Sentry
	Env    *opts.Env
	Port   int
	Web    *Web
}

func (s *Server) Start() {
	log.Printf("[INFO] environment: %s", s.Env.Type)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	if s.Sentry.Init(s.Env.Type, "web") {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
			Timeout: 10 * time.Second,
		})
		router.Use(sentryMiddleware.Handle)
	}

	router.Route("/", func(r chi.Router) {
		r.With(sessionAuth(s.Web.API.GetUser)).Route("/", func(r chi.Router) {
			r.Get("/", s.Web.GetIndex)
			r.Get("/scoreboard", s.Web.GetScoreboard)
			r.Get("/commit/{login}:{commitHash:[0-9a-z]+}", s.Web.GetCommit)
		})
		r.Get("/signin", s.Web.GetSignin)
		r.Get("/logout", s.Web.Logout)
	})
	router.NotFound(s.Web.NotFound)

	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}

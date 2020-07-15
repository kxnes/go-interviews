package md5calc

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"github.com/kxnes/go-interviews/md5calc/services/md5calc/repo"
	"net/http"
	"time"

	"github.com/kxnes/go-interviews/md5calc/pkg/logging"
)

type Deps struct {
	DB *gorm.DB
}

type Opts struct {
	Timeout time.Duration
}

type resource struct {
	db  *gorm.DB
	cli *http.Client
}

type service struct {
	r   resource
	log *logging.Logger
	api chi.Router
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.api.ServeHTTP(w, r)
}

func (s *service) Submit(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'am submit"))
}

func (s *service) Check(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'am check"))
}

func (s *service) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'am delete"))
}

func (s *service) migrate() {
	s.r.db.AutoMigrate()
}

func New(deps Deps, opts Opts) *service {
	log := logging.New("md5calc")

	if deps.DB == nil {
		log.Fatal(errors.New("database not initialized"))
	}

	s := service{
		r: resource{
			db:  deps.DB,
			cli: &http.Client{Timeout: opts.Timeout * time.Second},
		},
		log: log,
		api: chi.NewMux(),
	}

	s.api.Post("/", s.Submit)
	s.api.Get("/{id:[0-9]+}", s.Check)
	s.api.Delete("/{id:[0-9]+}", s.Delete)

	s.r.db.AutoMigrate(&repo.Path{}, &repo.Checksum{})

	return &s
}

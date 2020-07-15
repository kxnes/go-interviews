package internal

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kxnes/go-interviews/md5calc/pkg/logging"
	"github.com/kxnes/go-interviews/md5calc/services/md5calc"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func ListenAndServe() {
	log := logging.New("server")

	env, err := setup()
	if err != nil {
		log.Fatal(err)
	}

	gorm.NowFunc = func() time.Time {
		return time.Now().UTC()
	}

	db, err := gorm.Open(env.database.dialect, env.databaseURI())
	if err != nil {
		log.Fatal(err)
	}

	db.DB().SetMaxOpenConns(env.database.maxConn)
	db.DB().SetMaxIdleConns(env.database.maxConn)
	db.DB().SetConnMaxLifetime(env.database.connLT)

	r := chi.NewRouter()
	r.Route("/api", func(api chi.Router) {
		api.Mount("/md5", md5calc.New(
			md5calc.Deps{DB: db},
			md5calc.Opts{Timeout: env.worker.timeout},
		))
	})

	s := http.Server{
		Addr:         env.serverURI(),
		WriteTimeout: env.server.rwTimeout,
		ReadTimeout:  env.server.rwTimeout,
		IdleTimeout:  env.server.idleTimeout,
		Handler:      r,
	}

	defer func() {
		_ = db.Close()
		ctx, _ := context.WithTimeout(context.Background(), env.server.shutTimeout)
		if err := s.Shutdown(ctx); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	graceful := make(chan os.Signal)
	signal.Notify(graceful, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-graceful
}

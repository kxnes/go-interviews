package internal

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
	"time"
)

const errFmt = "%q > %w"

var errMissing = errors.New("missing")

type environment struct {
	database struct {
		dialect  string
		host     string
		port     int
		username string
		password string
		name     string
		maxConn  int
		connLT   time.Duration
		sslMode  string
	}
	server struct {
		host        string
		port        int
		rwTimeout   time.Duration
		idleTimeout time.Duration
		shutTimeout time.Duration
	}
	worker struct {
		timeout time.Duration
	}
	errs []error
}

func (env *environment) Error() string {
	s := make([]string, len(env.errs))

	for i, e := range env.errs {
		s[i] = e.Error()
	}

	return fmt.Sprintf("environment errors:\n%s", strings.Join(s, "\n"))
}

func (env *environment) lookup(key string, conv func(string) (interface{}, error)) interface{} {
	e, ok := os.LookupEnv(key)
	val, err := conv(e)

	if !ok {
		env.errs = append(env.errs, fmt.Errorf(errFmt, key, errMissing))
		return val
	}

	if err != nil {
		env.errs = append(env.errs, fmt.Errorf(errFmt, key, err))
	}

	return val
}

func (env *environment) lookupString(key string) string {
	return env.lookup(key, func(s string) (interface{}, error) {
		return s, nil
	}).(string)
}

func (env *environment) lookupInt(key string) int {
	return env.lookup(key, func(s string) (interface{}, error) {
		return strconv.Atoi(s)
	}).(int)
}

func (env *environment) lookupDuration(key string) time.Duration {
	return env.lookup(key, func(s string) (interface{}, error) {
		return time.ParseDuration(s)
	}).(time.Duration)
}

func (env *environment) databaseURI() string {
	return strings.Join([]string{
		"host=" + env.database.host,
		"port=" + strconv.Itoa(env.database.port),
		"user=" + env.database.username,
		"password=" + env.database.password,
		"dbname=" + env.database.name,
		"sslmode=" + env.database.sslMode,
	}, " ")
}

func (env *environment) serverURI() string {
	return env.server.host + ":" + strconv.Itoa(env.server.port)
}

func setup() (*environment, error) {
	// ignore error for flexibility using both .env and os.Environ() directly
	_ = godotenv.Load()

	env := new(environment)

	env.database.dialect = env.lookupString("DB_DIALECT")
	env.database.host = env.lookupString("DB_HOST")
	env.database.port = env.lookupInt("DB_PORT")
	env.database.username = env.lookupString("DB_USERNAME")
	env.database.password = env.lookupString("DB_PASSWORD")
	env.database.name = env.lookupString("DB_NAME")
	env.database.maxConn = env.lookupInt("DB_MAX_CONN")
	env.database.connLT = env.lookupDuration("DB_CONN_LIFETIME")
	env.database.sslMode = env.lookupString("DB_SSL_MODE")

	env.server.host = env.lookupString("SERVER_HOST")
	env.server.port = env.lookupInt("SERVER_PORT")
	env.server.rwTimeout = env.lookupDuration("SERVER_RW_TIMEOUT")
	env.server.idleTimeout = env.lookupDuration("SERVER_IDLE_TIMEOUT")
	env.server.shutTimeout = env.lookupDuration("SERVER_SHUTDOWN_TIMEOUT")

	env.worker.timeout = env.lookupDuration("WORKER_TIMEOUT")

	if len(env.errs) != 0 {
		return nil, env
	}

	return env, nil
}

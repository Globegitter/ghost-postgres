package ghost_postgres

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/phayes/freeport"
)

var (
	postgresPortRegex  = regexp.MustCompile(`.*\(port (\d+)\).*`)
	defaultInitdbArgs  = []string{"-U", "postgres", "-A", "trust", "--lc-messages=C"}
	defaultTmpDir      = "/tmp"
	defaultHost        = "localhost"
	defaultUser        = "postgres"
	defaultDbname      = "test"
	defaultPrefix      = "testing-postgresql-"
	defaultInitDbGlobs = []string{
		"/usr/local/pgsql/bin/initdb",
		"/usr/local/initdb",
		"/usr/lib/postgresql/*/bin/initdb",      // for Debian/Ubuntu
		"/opt/local/lib/postgresql*/bin/initdb", // for MacPorts
		"/usr/local/bin/initdb",                 // for Homebrew
		"/usr/bin/initdb",                       // for Alpine
	}
	defaultPostgresGlobs = []string{
		"/usr/local/pgsql/bin/postgres",
		"/usr/local/postgres",
		"/usr/lib/postgresql/*/bin/postgres",      // for Debian/Ubuntu
		"/opt/local/lib/postgresql*/bin/postgres", // for MacPorts
		"/usr/local/bin/postgres",                 // for Homebrew
		"/usr/bin/postgres",                       // for Alpine
	}
	defaultConnOptions = []string{
		"sslmode=disable",
	}
	defaultLogWriter     = ioutil.Discard
	defaultLogPrefix     = "ghost_postgres"
	defaultLogFlags      = log.LstdFlags
	defaultDefaultDbname = "postgres"
)

type GhostPostgres struct {
	Host          string
	Port          int
	User          string
	Dbname        string
	InitdbArgs    []string
	InitdbGlobs   []string
	PostgresGlobs []string
	TmpDir        string
	Prefix        string
	Dir           string
	Postgres      *exec.Cmd
	ConnOptions   []string
	Logger        *log.Logger
	DefaultDbname string
}

func New() *GhostPostgres {
	return &GhostPostgres{
		Host:          defaultHost,
		User:          defaultUser,
		Dbname:        defaultDbname,
		InitdbGlobs:   defaultInitDbGlobs,
		InitdbArgs:    defaultInitdbArgs,
		PostgresGlobs: defaultPostgresGlobs,
		TmpDir:        defaultTmpDir,
		Prefix:        defaultPrefix,
		Port:          freeport.GetPort(),
		ConnOptions:   defaultConnOptions,
		Logger:        log.New(defaultLogWriter, defaultLogPrefix, defaultLogFlags),
		DefaultDbname: defaultDefaultDbname,
	}
}

// Init runs initdb to create the PostgreSQL database.
func (gp *GhostPostgres) Init() error {
	initdbs, err := findExecs(gp.InitdbGlobs)
	if err != nil {
		return err
	}
	initdb := initdbs[0]
	dir, err := ioutil.TempDir(gp.TmpDir, gp.Prefix)
	if err != nil {
		return err
	}
	gp.Dir = dir
	dataDir := dir + "/data"
	args := append([]string{"-D", dataDir}, gp.InitdbArgs...)
	cmd := exec.Command(initdb, args...)
	out, err := cmd.CombinedOutput()
	gp.Logger.Print(string(out))
	if err != nil {
		return err
	}
	return nil
}

// Start starts the postgresql server and waits for it to become available.
func (gp *GhostPostgres) Start() error {
	postgress, err := findExecs(gp.PostgresGlobs)
	if err != nil {
		return err
	}
	postgres := postgress[0]
	dataDir, err := gp.DataDir()
	if err != nil {
		return err
	}
	cmd := exec.Command(postgres, "-p", strconv.Itoa(gp.Port), "-D", dataDir, "-k", gp.TmpDir, "-h", gp.Host, "-F")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	output := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(output)
	go func(l *log.Logger, s *bufio.Scanner) {
		for s.Scan() {
			l.Printf(s.Text())
		}
		if err := s.Err(); err != nil {
			l.Printf("error: %s", err)
		}
	}(gp.Logger, scanner)
	err = cmd.Start()
	if err != nil {
		return err
	}
	db, err := gp.OpenFor(gp.DefaultDbname)
	if err != nil {
		return err
	}
	err = waitForService(db)
	if err != nil {
		return err
	}
	gp.Postgres = cmd
	return nil
}

// Create creates the database on the running postgres server.
func (gp *GhostPostgres) Create() error {
	db, err := gp.OpenFor(gp.DefaultDbname)
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", gp.Dbname))
	if err != nil {
		return err
	}
	return nil
}

// Prepare returns a functioning ephemeral database or an error.
func (gp *GhostPostgres) Prepare() error {
	if err := gp.Init(); err != nil {
		return err
	}
	if err := gp.Start(); err != nil {
		return err
	}
	if err := gp.Create(); err != nil {
		return err
	}
	return nil
}

// Stop stops the running postgresql server.
// It first tries to interrupt the process, and then kills it.
func (gp *GhostPostgres) Stop() error {
	if gp.Postgres != nil && gp.Postgres.Process != nil {
		return interruptThenKill(gp.Postgres, time.Second)
	}
	return nil
}

// Destroy removes the directory for the database.
func (gp *GhostPostgres) Destroy() error {
	// TODO cleanup sockets? - /tmp/.s.PGSQL.PID and /tmp/.s.PGSQL.PID.lock
	return os.RemoveAll(gp.Dir)
}

// Terminate stops and destroys the database.
func (gp *GhostPostgres) Terminate() error {
	gp.Stop()
	return gp.Destroy()
}

// DataDir returns the data directory for the database.
func (gp *GhostPostgres) DataDir() (string, error) {
	if gp.Dir == "" {
		return "", fmt.Errorf("Data directory has not be initialized")
	}
	return fmt.Sprintf("%s/data", gp.Dir), nil
}

// Open opens a connection to the testing database
func (gp *GhostPostgres) Open() (*sql.DB, error) {
	return gp.OpenFor(gp.Dbname)
}

// ConnString returns a connection string for the testing database
func (gp *GhostPostgres) ConnString() string {
	return gp.ConnStringFor(gp.Dbname)
}

// URL returns a URL string for the testing database
func (gp *GhostPostgres) URL() string {
	return gp.URLFor(gp.Dbname)
}

// OpenFor opens a connection to the database with the specified name
func (gp *GhostPostgres) OpenFor(dbname string) (*sql.DB, error) {
	return sql.Open("postgres", gp.ConnStringFor(dbname))
}

// ConnStringFor returns a connection string for the database with the specified name
func (gp *GhostPostgres) ConnStringFor(dbname string) string {
	options := strings.Join(gp.ConnOptions, " ")
	// TODO check that values aren't empty?
	return fmt.Sprintf("user=%s host=%s port=%d dbname=%s %s", gp.User, gp.Host, gp.Port, dbname, options)
}

// URLFor returns a URL string for the database with the specified name
func (gp *GhostPostgres) URLFor(dbname string) string {
	options := strings.Join(gp.ConnOptions, "&")
	// TODO check that values aren't empty?
	return fmt.Sprintf("postgres://%s@%s:%d/%s?%s", gp.User, gp.Host, gp.Port, dbname, options)
}

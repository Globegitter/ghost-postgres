package ghost_postgres

import (
	"fmt"
)

func ExampleUsage() {
	gp := New()
	defer gp.Terminate()
	err := gp.Prepare()
	if err != nil {
		fmt.Println(err)
		return
	}
	db, err := gp.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := db.Exec("CREATE TABLE test ( id SERIAL NOT NULL, text VARCHAR(100) NOT NULL, PRIMARY KEY (id));"); err != nil {
		fmt.Println(err)
		return
	}
	if _, err := db.Exec("INSERT INTO test (text) VALUES ('Hello, World');"); err != nil {
		fmt.Println(err)
		return
	}
	var s string
	if err := db.QueryRow("SELECT text FROM test").Scan(&s); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(s)
	// Output: Hello, World
}

func ExampleInit() {
	gp := New()
	defer gp.Terminate()
	// Init runs initdb, which creates the files and directories for the db
	err := gp.Init()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleStart() {
	gp := New()
	defer gp.Terminate()
	if err := gp.Init(); err != nil {
		fmt.Println(err)
		return
	}
	// Start executes the postgres process and waits until it is ready to accept connections
	err := gp.Start()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleCreate() {
	gp := New()
	defer gp.Terminate()
	if err := gp.Init(); err != nil {
		fmt.Println(err)
		return
	}
	if err := gp.Start(); err != nil {
		fmt.Println(err)
		return
	}
	// Create creates the testing database specified by the dbname
	err := gp.Create()
	fmt.Println(err == nil)
	// Output: true
}

func ExamplePrepare() {
	gp := New()
	defer gp.Terminate()
	// Prepare runs Init, Start, and Create;
	// when Prepare is finished, the DB is ready-to-use
	err := gp.Prepare()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleStop() {
	gp := New()
	defer gp.Terminate()
	if err := gp.Prepare(); err != nil {
		fmt.Println(err)
		return
	}
	// Stop stops the postgres process;
	// first it issues an interrupt signal, and if postgres doesn't exit,
	// it send a kill signal.
	err := gp.Stop()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleTerminate() {
	gp := New()
	defer gp.Terminate()
	if err := gp.Prepare(); err != nil {
		fmt.Println(err)
		return
	}
	// Terminate runs Stop and then Destroy
	err := gp.Terminate()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleDataDir() {
	gp := &GhostPostgres{Dir: "/tmp/ghost-postgres"}
	// DataDir returns the data directory for the db
	dataDir, err := gp.DataDir()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dataDir)
	// Output: /tmp/ghost-postgres/data
}

func ExampleOpen() {
	gp := New()
	defer gp.Terminate()
	if err := gp.Prepare(); err != nil {
		fmt.Println(err)
		return
	}
	// Open wraps sql.Open
	db, err := gp.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = db.Ping()
	fmt.Println(err == nil)
	// Output: true
}

func ExampleConnString() {
	gp := &GhostPostgres{User: "postgres", Host: "localhost", Port: 12345, Dbname: "test", ConnOptions: []string{"sslmode=disable"}}
	// ConnString creates the connection string used for sql.Open() or psql
	fmt.Println(gp.ConnString())
	// Output: user=postgres host=localhost port=12345 dbname=test sslmode=disable
}

func ExampleURL() {
	gp := &GhostPostgres{User: "postgres", Host: "localhost", Port: 12345, Dbname: "test", ConnOptions: []string{"sslmode=disable"}}
	// URL creates the connection URL used for sql.Open() or psql
	fmt.Println(gp.URL())
	// Output: postgres://postgres@localhost:12345/test?sslmode=disable
}

package ghost_postgres

import (
	// "fmt"
	"log"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInit(t *testing.T) {
	Convey("Given...", t, func() {
		Convey("If we can't find initdb", func() {
			gp := &GhostPostgres{
				Host:          defaultHost,
				User:          defaultUser,
				Dbname:        defaultDbname,
				InitdbGlobs:   []string{},
				InitdbArgs:    defaultInitdbArgs,
				TmpDir:        defaultTmpDir,
				Prefix:        defaultPrefix,
				Logger:        log.New(defaultLogWriter, defaultLogPrefix, defaultLogFlags),
				DefaultDbname: defaultDefaultDbname,
			}
			defer gp.Terminate()
			err := gp.Init()
			Convey("An error should be returned", func() { So(err, ShouldNotBeNil) })
			Convey("The error message should contain...", func() {
				So(err.Error(), ShouldContainSubstring, "Could not find any matching")
			})
		})
		Convey("If we can't create a tmp dir", func() {
			gp := &GhostPostgres{
				Host:          defaultHost,
				User:          defaultUser,
				Dbname:        defaultDbname,
				InitdbGlobs:   defaultInitDbGlobs,
				InitdbArgs:    defaultInitdbArgs,
				TmpDir:        "/fooBarBaz",
				Prefix:        defaultPrefix,
				Logger:        log.New(defaultLogWriter, defaultLogPrefix, defaultLogFlags),
				DefaultDbname: defaultDefaultDbname,
			}
			defer gp.Terminate()
			err := gp.Init()
			Convey("An error should be returned", func() { So(err, ShouldNotBeNil) })
			Convey("The error message should contain...", func() {
				So(err.Error(), ShouldContainSubstring, "no such file or directory")
			})
		})
		Convey("If we call initdb and it fails", func() {
			gp := &GhostPostgres{
				Host:          defaultHost,
				User:          defaultUser,
				Dbname:        defaultDbname,
				InitdbGlobs:   defaultInitDbGlobs,
				InitdbArgs:    []string{"--foo=bar"},
				TmpDir:        defaultTmpDir,
				Prefix:        defaultPrefix,
				Logger:        log.New(defaultLogWriter, defaultLogPrefix, defaultLogFlags),
				DefaultDbname: defaultDefaultDbname,
			}
			defer gp.Terminate()
			err := gp.Init()
			Convey("An error should be returned", func() { So(err, ShouldNotBeNil) })
			Convey("The error message should contain...", func() {
				So(err.Error(), ShouldContainSubstring, "exit status")
			})
		})
		Convey("If we call initdb and it's successful", func() {
			gp := New()
			defer gp.Terminate()
			err := gp.Init()
			Convey("No error should be returned", func() { So(err, ShouldBeNil) })
		})
	})
}

func TestStart(t *testing.T) {
	Convey("Given a GhostPostgres", t, func() {
		Convey("If we can't find a postgres executable", func() {
			gp := New()
			defer gp.Terminate()
			gp.PostgresGlobs = []string{}
			err := gp.Init()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Start()
			Convey("An error should be returned", func() { So(err, ShouldNotBeNil) })
			Convey("The error message should contain", func() {
				So(err.Error(), ShouldContainSubstring, "Could not find any matching paths for globs")
			})
		})
		Convey("If we can start postgres", func() {
			gp := New()
			defer gp.Terminate()
			err := gp.Init()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Start()
			Convey("No error should be returned", func() { So(err, ShouldBeNil) })
		})
	})
}

func TestCreate(t *testing.T) {
	Convey("Given a GhostPostgres", t, func() {
		Convey("If we can't connect to the default database", func() {
			gp := New()
			defer gp.Terminate()
			gp.DefaultDbname = "foo"
			err := gp.Init()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Start()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Create()
			Convey("An error should be returned", func() { So(err, ShouldNotBeNil) })
			Convey("The error message should contain", func() {
				So(err.Error(), ShouldContainSubstring, "does not exist")
			})
		})
		Convey("If we can create the database", func() {
			gp := New()
			defer gp.Terminate()
			err := gp.Init()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Start()
			if err != nil {
				t.Fatal(err)
			}
			err = gp.Create()
			Convey("No error should be returned", func() { So(err, ShouldBeNil) })
		})
	})
}

func TestPrepare(t *testing.T) {
	Convey("Given a GhostPostgres", t, func() {
		Convey("If we can connect", func() {
			gp := New()
			defer gp.Terminate()
			err := gp.Prepare()
			Convey("No error should be returned", func() { So(err, ShouldBeNil) })
		})
	})
}

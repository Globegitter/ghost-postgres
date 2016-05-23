# ghost-postgres
Create ephemeral postgresql databases, e.g. for testing

Inspired by [testing.postgresql](https://github.com/tk0miya/testing.postgresql) for python.

## Getting Started

This has been tested on ubuntu 14.04.
It will likely work on other ubuntu versions and possibly other \*NIX distros.

It's been tested the PostgreSQL 9.5, and should work with 9.x.

### Setup

Install PostgreSQL 9.x; on Ubuntu (14.04):
```
sudo apt-get install postgresql postgresql-contrib postgresql-common postgresql-client-common
```

### Example

```
package main

import (
    "github.com/recursionpharma/ghost-postgres"
)

func main() {
	gp := ghost_postgres.New()
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
```

## Directory Structure

```
code/ghost-postgres/
|-- examples_test.go
|   Examples
|-- file.go
|   File utilities
|-- .gitignore
|   Ignored git files
|-- LICENSE
|   Software license (MIT)
|-- postgresql.go
|   The main code and interface
|-- postgresql_test.go
|   Tests
|-- README.md
|   This file
|-- .travis.yml
|   Travis config file
`-- util.go
    Helper functions
```
The above file tree was generated with `tree -a -L 1 --charset ascii`.

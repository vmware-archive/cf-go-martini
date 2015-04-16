package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Language struct {
	Name    string
	Creator string
}

func main() {
	m := martini.Classic()

	db := initDB()

	m.Use(DB(db))
	m.Use(render.Renderer())

	m.Get("/", func(r render.Render) {
		appEnv, _ := cfenv.Current()

		r.HTML(200, "hello", appEnv)
	})

	m.Get("/languages", func(r render.Render, db *sql.DB) {
		languages, err := fetchLanguages(db)

		if err != nil {
			r.HTML(500, "error", err)
		} else {
			r.HTML(200, "languages", languages)
		}
	})

	m.Run()
}

func fetchLanguages(db *sql.DB) (languages []*Language, err error) {
	rs, err := db.Query("select name, creator FROM languages")
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	languages = make([]*Language, 0)

	for rs.Next() {
		language := new(Language)
		err = rs.Scan(&language.Name, &language.Creator)
		languages = append(languages, language)

		if err != nil {
			return nil, err
		}
	}
	err = rs.Err()
	if err != nil {
		return nil, err
	}

	return
}

func DB(db *sql.DB) martini.Handler {
	return func(c martini.Context) {
		c.Map(db)
		c.Next()
	}
}

func initDB() *sql.DB {
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxOpenConns(4) // for ClearDB free plan

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	if schemaIsNotCreated(db) {
		createSchema(db)
	}

	return db
}

func dsn() string {
	appEnv, _ := cfenv.Current()
	services := appEnv.Services
	var mysqlService cfenv.Service

	for _, instances := range services {
		for _, instance := range instances {
			if contains(instance.Tags, "mysql") {
				mysqlService = instance
			}
		}
	}

	credentials := mysqlService.Credentials
	return fmt.Sprintf("%v:%v@tcp(%v:3306)/%v",
		credentials["username"],
		credentials["password"],
		credentials["hostname"],
		credentials["name"])
}

func schemaIsNotCreated(db *sql.DB) bool {
	rs, err := db.Query("select * from languages limit 1")
	if err != nil {
		return true
	} else {
		rs.Close()
		return false
	}
}

func createSchema(db *sql.DB) {
	_, err := db.Exec(
		"CREATE TABLE languages (name varchar(45) NOT NULL, creator varchar(45) NOT NULL, PRIMARY KEY (name)) ENGINE=InnoDB DEFAULT CHARSET=utf8")

	if err != nil {
		panic(err.Error())
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err.Error())
	}

	insertRow(db, "INSERT INTO languages (name, creator) VALUES ('Go','Rob')")
	insertRow(db, "INSERT INTO languages (name, creator) VALUES ('Java','James')")
	insertRow(db, "INSERT INTO languages (name, creator) VALUES ('Clojure','Rich')")
	insertRow(db, "INSERT INTO languages (name, creator) VALUES ('Ruby','Matz')")
	insertRow(db, "INSERT INTO languages (name, creator) VALUES ('Python','Guido')")

	err = tx.Commit()
	if err != nil {
		panic(err.Error())
	}
}

func insertRow(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		panic(err.Error())
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

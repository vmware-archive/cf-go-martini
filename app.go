package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/go-martini/martini"
	"github.com/joefitzgerald/cfenv"
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
		appEnv := cfenv.Current()

		r.HTML(200, "hello", appEnv)
	})

	m.Get("/languages", func(r render.Render, db *sql.DB) {
		rows, err := db.Query("select name, creator FROM languages")
		defer rows.Close()

		if err != nil {
			r.HTML(500, "error", err)
		}

		languages, err := mapRowsToLanguages(rows)

		if err != nil {
			r.HTML(500, "error", err)
		}

		r.HTML(200, "languages", languages)
	})

	m.Run()
}

func mapRowsToLanguages(rs *sql.Rows) (languages []*Language, err error) {
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

	db.SetMaxOpenConns(4)

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return db
}

func dsn() string {
	services := cfenv.Current().Services
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

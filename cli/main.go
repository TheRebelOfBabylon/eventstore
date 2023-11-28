package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fiatjaf/eventstore"
	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/elasticsearch"
	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/fiatjaf/eventstore/mysql"
	"github.com/fiatjaf/eventstore/postgresql"
	"github.com/fiatjaf/eventstore/sqlite3"
	"github.com/urfave/cli/v2"
)

var db eventstore.Store

var app = &cli.App{
	Name:      "eventstore",
	Usage:     "a CLI for all the eventstore backends",
	UsageText: "eventstore -d ./data/sqlite <query|put|del> ...",
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "store",
			Aliases: []string{"d"},
			Usage:   "path to the database file or directory or database connection uri",
		},
		&cli.StringFlag{
			Name:    "type",
			Aliases: []string{"t"},
			Usage:   "store type ('sqlite', 'lmdb', 'badger', 'postgres', 'mysql', 'elasticsearch')",
		},
	},
	Before: func(c *cli.Context) error {
		path := c.Path("store")
		typ := c.String("type")
		if typ != "" {
			// bypass automatic detection
			// this also works for creating disk databases from scratch
		} else {
			// try to detect based on url scheme
			switch {
			case strings.HasPrefix(path, "postgres://"), strings.HasPrefix(path, "postgresql://"):
				typ = "postgres"
			case strings.HasPrefix(path, "mysql://"):
				typ = "mysql"
			case strings.HasPrefix(path, "https://"):
				// if we ever add something else that uses URLs we'll have to modify this
				typ = "elasticsearch"
			default:
				// try to detect based on the form and names of disk files
				dbname, err := detect(path)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf(
							"'%s' does not exist, to create a store there specify the --type argument", path)
					}
					return fmt.Errorf("failed to detect store type: %w", err)
				}
				typ = dbname
			}
		}

		switch typ {
		case "sqlite":
			db = &sqlite3.SQLite3Backend{DatabaseURL: path}
		case "lmdb":
			db = &lmdb.LMDBBackend{Path: path, MaxLimit: 5000}
		case "badger":
			db = &badger.BadgerBackend{Path: path, MaxLimit: 5000}
		case "postgres", "postgresql":
			db = &postgresql.PostgresBackend{DatabaseURL: path}
		case "mysql":
			db = &mysql.MySQLBackend{DatabaseURL: path}
		case "elasticsearch":
			db = &elasticsearch.ElasticsearchStorage{URL: path}
		case "":
			return fmt.Errorf("couldn't determine store type, you can use --type to specify it manually")
		default:
			return fmt.Errorf("'%s' store type is not supported by this CLI", typ)
		}

		return db.Init()
	},
	Commands: []*cli.Command{
		query,
		put,
		del,
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
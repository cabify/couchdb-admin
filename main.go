package main

import (
	"os"

	"github.com/cabify/couchdb-admin/cluster"
	"github.com/cabify/couchdb-admin/database"
	"github.com/cabify/couchdb-admin/httpUtils"
	"github.com/kr/pretty"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "The server to connect to",
			Value: "127.0.0.1",
			// TODO this should be required
		},
		cli.StringFlag{
			Name:  "admin",
			Usage: "Admin of the DB",
			Value: "admin",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "Password for the db's admin",
			Value: "password",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "describe_db",
			Action: func(c *cli.Context) error {
				db := database.LoadDB(c.String("db"), buildAuthHttpReq(c))
				pretty.Println(db)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "Database on which to operate",
					// TODO this should be required
				},
			},
		},
		{
			Name: "replicate",
			Action: func(c *cli.Context) error {
				ahr := buildAuthHttpReq(c)
				db := database.LoadDB(c.String("db"), ahr)
				return db.Replicate(c.String("shard"), c.String("replica"), ahr)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "shard",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "replica",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "db",
					// TODO this should be required
				},
			},
		},
		{
			Name: "add_node",
			Action: func(c *cli.Context) error {
				ahr := buildAuthHttpReq(c)
				cluster.LoadCluster(ahr).AddNode(c.String("node"), ahr)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "node",
					// TODO this should be required
				},
			},
		},
		{
			Name: "describe_cluster",
			Action: func(c *cli.Context) error {
				cluster.LoadCluster(buildAuthHttpReq(c))
				return nil
			},
		},
		{
			Name: "create_db",
			Action: func(c *cli.Context) error {
				database.CreateDatabase(c.String("db"), c.Int("replicas"), c.Int("shards"), buildAuthHttpReq(c))
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "db",
					// TODO this should be required
				},
				cli.IntFlag{
					Name:  "shards",
					Value: 8,
				},
				cli.IntFlag{
					Name:  "replicas",
					Value: 3,
				},
			},
		},
		{
			Name: "remove_replica",
			Action: func(c *cli.Context) error {
				ahr := buildAuthHttpReq(c)
				db := database.LoadDB(c.String("db"), ahr)
				return db.RemoveReplica(c.String("shard"), c.String("from"), ahr)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "db",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "shard",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "from",
					// TODO this should be required
				},
			},
		},
	}

	app.Run(os.Args)
}

func buildAuthHttpReq(c *cli.Context) *httpUtils.AuthenticatedHttpRequester {
	return httpUtils.NewAuthenticatedHttpRequester(c.GlobalString("username"), c.GlobalString("password"), c.GlobalString("server"))
}

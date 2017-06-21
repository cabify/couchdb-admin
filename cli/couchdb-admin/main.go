package main

import (
	"os"

	"github.com/cabify/couchdb-admin"
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
				db, err := couchdb_admin.LoadDB(c.String("db"), buildAuthHttpReq(c))
				if err != nil {
					return err
				} else {
					pretty.Println(db)
					return nil
				}
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
				db, err := couchdb_admin.LoadDB(c.String("db"), ahr)
				if err != nil {
					return err
				}
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
				couchdb_admin.LoadCluster(ahr).AddNode(c.String("node"), ahr)
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
				couchdb_admin.LoadCluster(buildAuthHttpReq(c))
				return nil
			},
		},
		{
			Name: "create_db",
			Action: func(c *cli.Context) error {
				couchdb_admin.CreateDatabase(c.String("db"), c.Int("replicas"), c.Int("shards"), buildAuthHttpReq(c))
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
				db, err := couchdb_admin.LoadDB(c.String("db"), ahr)
				if err != nil {
					return err
				}
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
		{
			Name: "disable_maintenance_mode",
			Action: func(c *cli.Context) error {
				ahr := buildAuthHttpReq(c)
				return couchdb_admin.NodeAt(c.String("node")).DisableMaintenance(ahr)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "node",
					// TODO this should be required
				},
			},
		},
		{
			Name: "set_config",
			Action: func(c *cli.Context) error {
				return couchdb_admin.NodeAt(c.String("node")).SetConfig(c.String("section"), c.String("key"), c.String("value"), buildAuthHttpReq(c))
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "node",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "section",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "key",
					// TODO this should be required
				},
				cli.StringFlag{
					Name: "value",
					// TODO this should be required
				},
			},
		},
		{
			Name: "remove_node",
			Action: func(c *cli.Context) error {
				ahr := buildAuthHttpReq(c)
				return couchdb_admin.LoadCluster(ahr).RemoveNode(couchdb_admin.NodeAt(c.String("node")), ahr)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "node",
					// TODO this should be required
				},
			},
		},
	}

	app.Run(os.Args)
}

func buildAuthHttpReq(c *cli.Context) *httpUtils.AuthenticatedHttpRequester {
	return httpUtils.NewAuthenticatedHttpRequester(c.GlobalString("admin"), c.GlobalString("password"), c.GlobalString("server"))
}

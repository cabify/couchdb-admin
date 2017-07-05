package main

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/cabify/couchdb-admin"
	"github.com/cabify/couchdb-admin/httpUtils"
	"github.com/kr/pretty"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "CouchDB 2 Admin tool"
	app.Usage = "Easily operate a CouchDB 2 cluster"
	app.UsageText = "$ couchdb-admin [COMMAND] [OPTIONS]"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "Carlos Alonso", Email: "carlos.alonso@cabify.com"}}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "The server to connect to",
			Value: "127.0.0.1",
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
			Name:  "describe_db",
			Usage: "Get a db's shards placement info",
			Action: func(c *cli.Context) {
				db_name := c.String("db")
				log.WithField("db", db_name).Info("Describing database...")
				db, err := couchdb_admin.LoadDB(db_name, buildAuthHttpReq(c))
				if err != nil {
					log.WithError(err).Error("Couldn't describe database!")
					return
				}
				pretty.Println(db)
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "Database on which to operate",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"db"}, c)
			},
		},
		{
			Name:  "replicate",
			Usage: "Replicate a database's shard into a node not containing it already",
			Action: func(c *cli.Context) {
				db_name := c.String("db")
				shard := c.String("shard")
				replica := c.String("replica")

				log.WithFields(log.Fields{"db": db_name, "shard": shard, "replica": replica}).Info("Replicating shard...")

				ahr := buildAuthHttpReq(c)
				db, err := couchdb_admin.LoadDB(db_name, ahr)
				if err != nil {
					log.WithError(err).WithField("db", db_name).Error("Couldn't load database")
					return
				}
				if err = db.Replicate(shard, replica, ahr); err != nil {
					log.WithError(err).WithFields(log.Fields{"db": db_name, "shard": shard, "replica": replica}).Error("Couldn't replicate shard!")
					return
				}
				log.WithFields(log.Fields{"db": db_name, "shard": shard, "replica": replica}).Info("Shard successfully replicated")
				log.WithField("node", replica).Warn("Node was sent into maintenance!. Remember to reenable it once it catches up with changes")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "shard",
					Usage: "Shard's identifier to replicate",
				},
				cli.StringFlag{
					Name:  "replica",
					Usage: "Node's address where to replicate the shard",
				},
				cli.StringFlag{
					Name:  "db",
					Usage: "Db to where the shard belongs to",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"db", "replica", "shard"}, c)
			},
		},
		{
			Name:  "add_node",
			Usage: "Join a node into the cluster",
			Action: func(c *cli.Context) {
				node := c.String("node")
				log.WithField("node", node).Info("Adding node to the cluster...")

				ahr := buildAuthHttpReq(c)
				cluster, err := couchdb_admin.LoadCluster(ahr)
				if err != nil {
					log.WithError(err).Error("Coulnd't load the cluster!")
					return
				}
				if err = cluster.AddNode(node, ahr); err != nil {
					log.WithField("node", node).WithError(err).Error("Couldn't add node!")
					return
				}
				log.WithField("node", node).Info("Successfully added node!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node",
					Usage: "Node's address",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"node"}, c)
			},
		},
		{
			Name:  "describe_cluster",
			Usage: "Get information of the cluster's nodes",
			Action: func(c *cli.Context) {
				log.WithField("server", c.GlobalString("server")).Info("Describing cluster layout...")
				cluster, err := couchdb_admin.LoadCluster(buildAuthHttpReq(c))
				if err != nil {
					log.WithError(err).Error("Couldn't describe cluster!")
					return
				}
				pretty.Println(cluster.NodesInfo)
			},
		},
		{
			Name:  "create_db",
			Usage: "Create a new database",
			Action: func(c *cli.Context) {
				db := c.String("db")
				replicas, shards := c.Int("replicas"), c.Int("shards")
				log.WithFields(log.Fields{"db": db, "replicas": replicas, "shards": shards}).Info("Creating database...")

				if _, err := couchdb_admin.CreateDatabase(db, replicas, shards, buildAuthHttpReq(c)); err != nil {
					log.WithFields(log.Fields{"db": db, "replicas": replicas, "shards": shards}).WithError(err).Error("Could not create database!")
					return
				}
				log.WithFields(log.Fields{"db": db, "replicas": replicas, "shards": shards}).Info("Database successfully created!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The database name",
				},
				cli.IntFlag{
					Name:  "shards",
					Usage: "Number of shards to distribute the database across",
					Value: 8,
				},
				cli.IntFlag{
					Name:  "replicas",
					Usage: "Number of replicas for each shard",
					Value: 3,
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"db"}, c)
			},
		},
		{
			Name:  "remove_replica",
			Usage: "Remove a shard from a particular node",
			Action: func(c *cli.Context) {
				db_name := c.String("db")
				shard := c.String("shard")
				replica := c.String("from")

				log.WithFields(log.Fields{"db": db_name, "shard": shard, "replica": replica}).Info("Removing shard ownership...")

				ahr := buildAuthHttpReq(c)
				db, err := couchdb_admin.LoadDB(db_name, ahr)
				if err != nil {
					log.WithField("db", db_name).WithError(err).Error("Couldn't load db config!")
					return
				}
				if err = db.RemoveReplica(shard, replica, ahr); err != nil {
					log.WithFields(log.Fields{"db": db_name, "shard": replica, "replica": replica}).WithError(err).Error("Replica could not be removed!")
					return
				}
				log.WithFields(log.Fields{"db": db_name, "shard": shard, "replica": replica}).Info("Replica shard successfully removed!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "The database to where the shard belongs to",
				},
				cli.StringFlag{
					Name:  "shard",
					Usage: "Name of the shard to remove from the node",
				},
				cli.StringFlag{
					Name:  "from",
					Usage: "Node's address from where to remove the shard's replica",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"db", "shard", "from"}, c)
			},
		},
		{
			Name:  "disable_maintenance_mode",
			Usage: "Disable a node's maintenance mode",
			Action: func(c *cli.Context) {
				node_name := c.String("node")
				log.WithField("node", node_name).Info("Removing maintenance flag...")

				ahr := buildAuthHttpReq(c)
				node, err := couchdb_admin.NodeAt(node_name)
				if err != nil {
					log.WithField("node", node_name).WithError(err).Error("Couldn't locate node!")
					return
				}
				if err = node.DisableMaintenance(ahr); err != nil {
					log.WithField("node", node_name).WithError(err).Error("Couldn't disable maintenance flag!")
					return
				}
				log.WithField("node", node_name).Info("Maintenance flag successfully removed!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node",
					Usage: "The node's address",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"node"}, c)
			},
		},
		{
			Name:  "set_config",
			Usage: "Set a config value of a particular node",
			Action: func(c *cli.Context) {
				node_name := c.String("node")
				section := c.String("section")
				key := c.String("key")
				value := c.String("value")
				log.WithFields(log.Fields{"node": node_name, "section": section, "key": key, "value": value}).Info("Setting config value...")

				node, err := couchdb_admin.NodeAt(node_name)
				if err != nil {
					log.WithError(err).WithField("node", node_name).Error("Couldn't locate node!")
					return
				}
				node.SetConfig(section, key, value, buildAuthHttpReq(c))
				log.WithFields(log.Fields{"node": node_name, "section": section, "key": key, "value": value}).Info("New config successfully applied!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node",
					Usage: "The node's address",
				},
				cli.StringFlag{
					Name:  "section",
					Usage: "CouchDB config's section where to apply the config",
				},
				cli.StringFlag{
					Name:  "key",
					Usage: "Key for the config value to apply",
				},
				cli.StringFlag{
					Name:  "value",
					Usage: "The actual config value to apply",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"node", "section", "key", "value"}, c)
			},
		},
		{
			Name:  "remove_node",
			Usage: "Remove a node from the cluster",
			Action: func(c *cli.Context) {
				node_name := c.String("node")
				log.WithField("node", node_name).Info("Removing node...")

				ahr := buildAuthHttpReq(c)
				cluster, err := couchdb_admin.LoadCluster(ahr)
				if err != nil {
					log.WithError(err).Error("Couldn't load cluster!")
					return
				}
				node, err := couchdb_admin.NodeAt(node_name)
				if err != nil {
					log.WithField("node", node_name).WithError(err).Error("Couldn't locate node!")
					return
				}
				if err = cluster.RemoveNode(node, ahr); err != nil {
					log.WithField("node", node_name).WithError(err).Error("Couldn't remove node!")
					return
				}
				log.WithField("node", node_name).Info("Node successfully removed!")
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node",
					Usage: "The node's address",
				},
			},
			Before: func(c *cli.Context) error {
				return requireFlags([]string{"node"}, c)
			},
		},
	}

	app.Run(os.Args)
}

func buildAuthHttpReq(c *cli.Context) *httpUtils.AuthenticatedHttpRequester {
	return httpUtils.NewAuthenticatedHttpRequester(c.GlobalString("admin"), c.GlobalString("password"), c.GlobalString("server"))
}

func requireFlags(names []string, c *cli.Context) error {
	for _, name := range names {
		if len(c.String(name)) == 0 {
			return fmt.Errorf("Missing %s parameter!", name)
		}
	}
	return nil
}

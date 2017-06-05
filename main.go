package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cabify/couchdb-admin/cluster"
	"github.com/cabify/couchdb-admin/http_utils"
	"github.com/kr/pretty"
	"github.com/urfave/cli"
)

type Data struct {
	Id        string              `json:"_id"`
	Rev       string              `json:"_rev"`
	Shards    []int               `json:"shard_suffix"`
	Changelog [][]string          `json:"changelog"`
	ByNode    map[string][]string `json:"by_node"`
	ByRange   map[string][]string `json:"by_range"`
}

var server string

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "server",
			Usage:       "The server to connect to",
			Value:       "127.0.0.1",
			Destination: &server,
			// TODO this should be required
		},
		cli.StringFlag{
			Name:        "admin",
			Usage:       "Admin of the DB",
			Value:       "admin",
			Destination: &http_utils.Username,
		},
		cli.StringFlag{
			Name:        "password",
			Usage:       "Password for the db's admin",
			Value:       "password",
			Destination: &http_utils.Password,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "describe_db",
			Action: describeDb,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "db",
					Usage: "Database on which to operate",
					// TODO this should be required
				},
			},
		},
		{
			Name:   "replicate",
			Action: replicate,
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
				cluster.LoadCluster(server).AddNode(server, c.String("node"))
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
				cluster.LoadCluster(server)
				return nil
			},
		},
		{
			Name:   "create_db",
			Action: createDatabase,
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
	}

	http_utils.HttpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	app.Run(os.Args)
}

func getDbConfig(db string) (data Data) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", server, db), nil)
	if err != nil {
		log.Fatal(err)
	}

	http_utils.RunRequest(req, &data)

	return
}

func describeDb(c *cli.Context) {
	data := getDbConfig(c.String("db"))
	pretty.Println(data)
}

func replicate(c *cli.Context) error {
	db := c.String("db")
	data := getDbConfig(db)

	replica := fmt.Sprintf("couchdb@%s", c.String("replica"))
	shard := c.String("shard")

	data.ByNode[replica] = append(data.ByNode[replica], shard)
	data.ByRange[shard] = append(data.ByRange[shard], replica)
	fmt.Printf("%+v\n", data)

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", server, db), bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	http_utils.RunRequest(req, nil)

	return nil
}

func createDatabase(c *cli.Context) error {
	replicas := c.Int("replicas")
	shards := c.Int("shards")
	db := c.String("db")

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/%s?n=%d&q=%d", server, db, replicas, shards), nil)
	if err != nil {
		log.Fatal(err)
	}

	http_utils.RunRequest(req, nil)

	return nil
}

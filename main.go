package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cabify/couchdb-admin/array_utils"
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
		{
			Name:   "remove_replica",
			Action: removeReplica,
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

func buildAuthHttpReq(c *cli.Context) *http_utils.AuthenticatedHttpRequester {
	return http_utils.NewAuthenticatedHttpRequester(c.GlobalString("username"), c.GlobalString("password"), c.GlobalString("server"))
}

func getDbConfig(db string, ahr *http_utils.AuthenticatedHttpRequester) (data Data) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db), nil)
	if err != nil {
		log.Fatal(err)
	}

	ahr.RunRequest(req, &data)

	return
}

func describeDb(c *cli.Context) {
	data := getDbConfig(c.String("db"), buildAuthHttpReq(c))
	pretty.Println(data)
}

func replicate(c *cli.Context) error {
	ahr := buildAuthHttpReq(c)
	db := c.String("db")
	data := getDbConfig(db, ahr)

	replica := fmt.Sprintf("couchdb@%s", c.String("replica"))
	shard := c.String("shard")

	data.ByNode[replica] = append(data.ByNode[replica], shard)
	data.ByRange[shard] = append(data.ByRange[shard], replica)
	// TODO add an entry to the changes section.
	fmt.Printf("%+v\n", data)

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db), bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	ahr.RunRequest(req, nil)

	return nil
}

func removeReplica(c *cli.Context) error {
	ahr := buildAuthHttpReq(c)
	db := c.String("db")
	shard := c.String("shard")
	replica := fmt.Sprintf("couchdb@%s", c.String("from"))

	data := getDbConfig(db, ahr)

	data.ByNode[replica] = array_utils.RemoveItem(data.ByNode[replica], shard)
	if len(data.ByNode[replica]) == 0 {
		delete(data.ByNode, replica)
	}

	data.ByRange[shard] = array_utils.RemoveItem(data.ByRange[shard], replica)
	// TODO add an entry to the changes section.
	fmt.Printf("%+v\n", data)

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db), bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	ahr.RunRequest(req, nil)

	return nil
}

func createDatabase(c *cli.Context) error {
	ahr := buildAuthHttpReq(c)
	replicas := c.Int("replicas")
	shards := c.Int("shards")
	db := c.String("db")

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/%s?n=%d&q=%d", ahr.GetServer(), db, replicas, shards), nil)
	if err != nil {
		log.Fatal(err)
	}

	ahr.RunRequest(req, nil)

	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

type ClusterLayout struct {
	AllNodes     []string `json:"all_nodes"`
	ClusterNodes []string `json:"cluster_nodes"`
}

var httpClient *http.Client
var server, username, password string

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
			Destination: &username,
		},
		cli.StringFlag{
			Name:        "password",
			Usage:       "Password for the db's admin",
			Value:       "password",
			Destination: &password,
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
			Name:   "add_node",
			Action: addNode,
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
				describeCluster()
				return nil
			},
		},
		{
			Name:   "create_database",
			Action: createDatabase,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name",
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

	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	app.Run(os.Args)
}

func runRequest(req *http.Request, dest interface{}) {
	req.SetBasicAuth(username, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode/100 != 2 {
		log.Printf("Received response %d for %s", resp.StatusCode, req.URL.String())
	}

	defer resp.Body.Close()

	if dest != nil {
		if err = json.NewDecoder(resp.Body).Decode(dest); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(resp)
	}
}

func getDbConfig(db string) (data Data) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", server, db), nil)
	if err != nil {
		log.Fatal(err)
	}

	runRequest(req, &data)

	return
}

func addNode(c *cli.Context) error {
	node := fmt.Sprintf("couchdb@%s", c.String("node"))

	describeCluster()

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_nodes/%s", server, node), strings.NewReader("{}"))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	runRequest(req, nil)

	describeCluster()

	return nil
}

func describeDb(c *cli.Context) {
	data := getDbConfig(c.String("db"))
	fmt.Printf("%# v", pretty.Formatter(data))
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

	runRequest(req, nil)

	return nil
}

func describeCluster() {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5984/_membership", server), nil)
	if err != nil {
		log.Fatal(err)
	}
	var cluster ClusterLayout

	runRequest(req, &cluster)

	fmt.Println("Current cluster layout")
	fmt.Printf("%+v\n", cluster)
}

func createDatabase(c *cli.Context) error {
	replicas := c.Int("replicas")
	shards := c.Int("shards")
	db := c.String("name")

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/%s?n=%d&q=%d", server, db, replicas, shards), nil)
	if err != nil {
		log.Fatal(err)
	}

	runRequest(req, nil)

	return nil
}

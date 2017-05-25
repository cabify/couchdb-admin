package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

var httpClient *http.Client

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server",
			Usage: "The server to connect to",
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
		cli.StringFlag{
			Name:  "db",
			Usage: "Database on which to operate",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "show",
			Action: show,
		},
		{
			Name:   "replicate",
			Action: replicate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "shard",
				},
				cli.StringFlag{
					Name: "replica",
				},
			},
		},
	}

	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	app.Run(os.Args)
}

func getDbConfig(server, db, username, password string) (data Data) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", server, db), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(username, password)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal(err)
	}

	return
}

func show(c *cli.Context) error {
	data := getDbConfig(c.GlobalString("server"), c.GlobalString("db"), c.GlobalString("admin"), c.GlobalString("password"))
	fmt.Printf("%+v\n", data)
	return nil
}

func replicate(c *cli.Context) error {
	server := c.GlobalString("server")
	db := c.GlobalString("db")
	username := c.GlobalString("admin")
	password := c.GlobalString("password")
	data := getDbConfig(server, db, username, password)

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
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	fmt.Println(resp)

	return nil
}

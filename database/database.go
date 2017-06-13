package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cabify/couchdb-admin/cluster"
	"github.com/cabify/couchdb-admin/http_utils"
	"github.com/cabify/couchdb-admin/sliceUtils"
)

type Database struct {
	name   string
	config Config
}

type Config struct {
	Id        string              `json:"_id"`
	Rev       string              `json:"_rev"`
	Shards    []int               `json:"shard_suffix"`
	Changelog [][]string          `json:"changelog"`
	ByNode    map[string][]string `json:"by_node"`
	ByRange   map[string][]string `json:"by_range"`
}

func LoadDB(name string, ahr *http_utils.AuthenticatedHttpRequester) *Database {
	db := &Database{
		name: name,
	}
	db.refreshDbConfig(ahr)
	return db
}

func CreateDatabase(name string, replicas, shards int, ahr *http_utils.AuthenticatedHttpRequester) *Database {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/%s?n=%d&q=%d", ahr.GetServer(), name, replicas, shards), nil)
	if err != nil {
		log.Fatal(err)
	}

	ahr.RunRequest(req, nil)

	return LoadDB(name, ahr)
}

func (db *Database) refreshDbConfig(ahr *http_utils.AuthenticatedHttpRequester) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), nil)
	if err != nil {
		log.Fatal(err)
	}

	ahr.RunRequest(req, &db.config)
}

func (db *Database) Replicate(shard, replica string, ahr *http_utils.AuthenticatedHttpRequester) error {
	replicaNode := fmt.Sprintf("couchdb@%s", replica)

	if sliceUtils.Contains(db.config.ByNode[replicaNode], shard) {
		return fmt.Errorf("%s is already replicating %s", replicaNode, shard)
	}

	if _, exists := db.config.ByRange[shard]; !exists {
		return fmt.Errorf("%s is not a %s's shard!", shard, db.name)
	}

	if !cluster.LoadCluster(ahr).IsNodeUpAndJoined(replicaNode) {
		return fmt.Errorf("%s is not part of the cluster!", replicaNode)
	}

	db.config.ByNode[replicaNode] = append(db.config.ByNode[replicaNode], shard)
	db.config.ByRange[shard] = append(db.config.ByRange[shard], replicaNode)
	// TODO add an entry to the changes section.

	b, err := json.Marshal(db.config)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	ahr.RunRequest(req, nil)

	return nil
}

func (db *Database) RemoveReplica(shard, from string, ahr *http_utils.AuthenticatedHttpRequester) error {
	replica := fmt.Sprintf("couchdb@%s", from)

	if _, exists := db.config.ByNode[replica]; !exists {
		return fmt.Errorf("%s does not have any replicas!", replica)
	}

	if !sliceUtils.Contains(db.config.ByNode[replica], shard) {
		return fmt.Errorf("Shard %s is not at %s", shard, replica)
	}

	newRange := sliceUtils.RemoveItem(db.config.ByRange[shard], replica)
	if len(newRange) == 0 {
		return fmt.Errorf("Aborting. Shard %s will be lost if deleted!!", shard)
	}
	db.config.ByRange[shard] = newRange

	newNode := sliceUtils.RemoveItem(db.config.ByNode[replica], shard)
	if len(newNode) > 0 {
		db.config.ByNode[replica] = newNode
	} else {
		delete(db.config.ByNode, replica)
	}
	// TODO add an entry to the changes section.

	b, err := json.Marshal(db.config)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	ahr.RunRequest(req, nil)

	return nil
}

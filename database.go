package couchdb_admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cabify/couchdb-admin/httpUtils"
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

func LoadDB(name string, ahr *httpUtils.AuthenticatedHttpRequester) (*Database, error) {
	db := &Database{
		name: name,
	}
	if err := db.refreshDbConfig(ahr); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

func CreateDatabase(name string, replicas, shards int, ahr *httpUtils.AuthenticatedHttpRequester) (*Database, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5984/%s?n=%d&q=%d", ahr.GetServer(), name, replicas, shards), nil)
	if err != nil {
		return nil, err
	}

	if err = ahr.RunRequest(req, nil); err != nil {
		return nil, err
	}

	return LoadDB(name, ahr)
}

func (db *Database) refreshDbConfig(ahr *httpUtils.AuthenticatedHttpRequester) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), nil)
	if err != nil {
		return err
	}

	if err = ahr.RunRequest(req, &db.config); err != nil {
		return err
	}

	if db.config.Id == "" {
		return fmt.Errorf("Could not retrieve config for db: %s", db.name)
	}
	return nil
}

func (db *Database) Replicate(shard, replica string, ahr *httpUtils.AuthenticatedHttpRequester) error {
	replicaNode := NodeAt(replica)

	if sliceUtils.Contains(db.config.ByNode[replicaNode.GetAddr()], shard) {
		return fmt.Errorf("%s is already replicating %s", replicaNode.GetAddr(), shard)
	}

	if _, exists := db.config.ByRange[shard]; !exists {
		return fmt.Errorf("%s is not a %s's shard!", shard, db.name)
	}

	if !LoadCluster(ahr).IsNodeUpAndJoined(replicaNode.GetAddr()) {
		return fmt.Errorf("%s is not part of the cluster!", replicaNode.GetAddr())
	}

	replicaNode.IntoMaintenance(ahr)

	db.config.ByNode[replicaNode.GetAddr()] = append(db.config.ByNode[replicaNode.GetAddr()], shard)
	db.config.ByRange[shard] = append(db.config.ByRange[shard], replicaNode.GetAddr())
	// TODO add an entry to the changes section.

	b, err := json.Marshal(db.config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return ahr.RunRequest(req, nil)
}

func (db *Database) RemoveReplica(shard, from string, ahr *httpUtils.AuthenticatedHttpRequester) error {
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
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:5986/_dbs/%s", ahr.GetServer(), db.name), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return ahr.RunRequest(req, nil)
}

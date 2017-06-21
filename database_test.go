package couchdb_admin

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cabify/couchdb-admin/httpUtils"
	"github.com/stretchr/testify/assert"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestLoadDBLoadsDB(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, db.name, "testdb")
	assert.Equal(t, db.config.Id, "testdb")
	assert.Equal(t, db.config.Rev, "1-5e2d10c29c70d3869fb7a1fd3a827a64")
	assert.Equal(t, db.config.Shards, []int{46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55})
	assert.Equal(t, db.config.Changelog, [][]string{[]string{"add", "00000000-7fffffff", "couchdb@127.0.0.1"}, []string{"add", "80000000-ffffffff", "couchdb@127.0.0.1"}})
	assert.Equal(t, db.config.ByNode, map[string][]string{"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"}})
	assert.Equal(t, db.config.ByRange, map[string][]string{"00000000-7fffffff": []string{"couchdb@127.0.0.1"}, "80000000-ffffffff": []string{"couchdb@127.0.0.1"}})
}

func TestCreateDatabaseWorks(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5984/testdb?n=1&q=2",
		httpmock.NewStringResponder(200, ""))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := CreateDatabase("testdb", 1, 2, ahr)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, db.name, "testdb")
	assert.Equal(t, db.config.Id, "testdb")
	assert.Equal(t, db.config.Rev, "1-5e2d10c29c70d3869fb7a1fd3a827a64")
	assert.Equal(t, db.config.Shards, []int{46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55})
	assert.Equal(t, db.config.Changelog, [][]string{[]string{"add", "00000000-7fffffff", "couchdb@127.0.0.1"}, []string{"add", "80000000-ffffffff", "couchdb@127.0.0.1"}})
	assert.Equal(t, db.config.ByNode, map[string][]string{"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"}})
	assert.Equal(t, db.config.ByRange, map[string][]string{"00000000-7fffffff": []string{"couchdb@127.0.0.1"}, "80000000-ffffffff": []string{"couchdb@127.0.0.1"}})
}

func TestReplicateAddsReplicaToNode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_dbs/testdb",
		func(req *http.Request) (*http.Response, error) {
			body := Config{}

			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
			}

			assert.Equal(t, body.ByNode, map[string][]string{
				"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
				"couchdb@127.0.0.2": []string{"00000000-7fffffff"},
			})
			assert.Equal(t, body.ByRange, map[string][]string{
				"00000000-7fffffff": []string{"couchdb@127.0.0.1", "couchdb@127.0.0.2"},
				"80000000-ffffffff": []string{"couchdb@127.0.0.1"},
			})

			return httpmock.NewStringResponse(200, ""), nil
		})

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]}`))

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5984/_node/couchdb@127.0.0.2/_config/couchdb/maintenance_mode",
		func(req *http.Request) (*http.Response, error) {
			defer req.Body.Close()

			bodyBytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, "\"true\"", string(bodyBytes))
			return httpmock.NewStringResponse(200, ""), nil
		})

	if err := db.Replicate("00000000-7fffffff", "127.0.0.2", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, db.name, "testdb")
	assert.Equal(t, db.config.Id, "testdb")
	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
		"couchdb@127.0.0.2": []string{"00000000-7fffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1", "couchdb@127.0.0.2"},
		"80000000-ffffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestReplicateFailsIfNodeIsAlreadyReplica(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	err = db.Replicate("00000000-7fffffff", "127.0.0.1", ahr)
	assert.Error(t, err, "Replica should have been rejected as 127.0.0.1 already contains a replica for the shard")

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
		"80000000-ffffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestReplicateFailsIfShardDoesNotExist(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	err = db.Replicate("dummy_shard", "127.0.0.1", ahr)
	assert.Error(t, err, "Replica should have been rejected as dummy_shard is not an existing DB shard")

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
		"80000000-ffffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestReplicateFailsIfNodeIsNotPartOfTheCluster(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"],
				"80000000-ffffffff": ["couchdb@127.0.0.1"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1"]}`))

	err = db.Replicate("00000000-7fffffff", "dummy_server", ahr)
	assert.Error(t, err, "Replica should have been rejected as dummy_server is not part of the cluster")

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
		"80000000-ffffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestRemoveReplicaWorks(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.2"
				],
				[
					"add",
					"80000000-ffffffff",
					"couchdb@127.0.0.2"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				],
				"couchdb@127.0.0.2": [
					"00000000-7fffffff",
					"80000000-ffffffff"
				]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"],
				"80000000-ffffffff": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]}}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_dbs/testdb",
		func(req *http.Request) (*http.Response, error) {
			body := Config{}

			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
			}

			assert.Equal(t, body.ByNode, map[string][]string{
				"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
				"couchdb@127.0.0.2": []string{"80000000-ffffffff"},
			})
			assert.Equal(t, body.ByRange, map[string][]string{
				"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
				"80000000-ffffffff": []string{"couchdb@127.0.0.1", "couchdb@127.0.0.2"},
			})

			return httpmock.NewStringResponse(200, ""), nil
		})

	if err := db.RemoveReplica("00000000-7fffffff", "127.0.0.2", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff", "80000000-ffffffff"},
		"couchdb@127.0.0.2": []string{"80000000-ffffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
		"80000000-ffffffff": []string{"couchdb@127.0.0.1", "couchdb@127.0.0.2"},
	})
}

func TestRemoveReplicaFailsIfNodeDoesNotContainShard(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": ["00000000-7fffffff"]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"]
			}
		}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	err = db.RemoveReplica("dummy_replica", "127.0.0.1", ahr)
	assert.Error(t, err, "Remove replica operation should have been rejected as the replica does not exist!")

	err = db.RemoveReplica("80000000-ffffffff", "dummy_server", ahr)
	assert.Error(t, err, "Remove replica operation should have been rejected as the server does not exist!")

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestRemoveReplicaFailsIfShardWillBeLost(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-7fffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": ["00000000-7fffffff"]
			},
			"by_range": {
				"00000000-7fffffff": ["couchdb@127.0.0.1"]
			}
		}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	err = db.RemoveReplica("00000000-7fffffff", "127.0.0.1", ahr)
	assert.Error(t, err, "Remove replica operation should have been rejected as the replica will be lost!")

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-7fffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-7fffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestRemoveReplicaRemovesNodeIfEmpty(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "1-5e2d10c29c70d3869fb7a1fd3a827a64",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-ffffffff",
					"couchdb@127.0.0.1"
				],
				[
					"add",
					"00000000-ffffffff",
					"couchdb@127.0.0.2"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": ["00000000-ffffffff"],
				"couchdb@127.0.0.2": ["00000000-ffffffff"]
			},
			"by_range": {
				"00000000-ffffffff": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]
			}
		}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	db, err := LoadDB("testdb", ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_dbs/testdb",
		func(req *http.Request) (*http.Response, error) {
			body := Config{}

			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
			}

			assert.Equal(t, body.ByNode, map[string][]string{
				"couchdb@127.0.0.1": []string{"00000000-ffffffff"},
			})
			assert.Equal(t, body.ByRange, map[string][]string{
				"00000000-ffffffff": []string{"couchdb@127.0.0.1"},
			})

			return httpmock.NewStringResponse(200, ""), nil
		})

	if err := db.RemoveReplica("00000000-ffffffff", "127.0.0.2", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, db.config.ByNode, map[string][]string{
		"couchdb@127.0.0.1": []string{"00000000-ffffffff"},
	})
	assert.Equal(t, db.config.ByRange, map[string][]string{
		"00000000-ffffffff": []string{"couchdb@127.0.0.1"},
	})
}

func TestFailureLoadingDBRaisesError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	_, err := LoadDB("testdb", ahr)
	assert.Error(t, err)
}

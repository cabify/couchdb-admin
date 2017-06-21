package couchdb_admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	httpmock "gopkg.in/jarcoal/httpmock.v1"

	"github.com/cabify/couchdb-admin/httpUtils"
	"github.com/stretchr/testify/assert"
)

func TestLoadClusterLoadsNodesInfo(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1","couchdb@127.0.0.1","couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1","couchdb@127.0.0.1","couchdb@127.0.0.1"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, cluster.NodesInfo.AllNodes, []string{"couchdb@127.0.0.1", "couchdb@127.0.0.1", "couchdb@127.0.0.1"})
	assert.Equal(t, cluster.NodesInfo.ClusterNodes, []string{"couchdb@127.0.0.1", "couchdb@127.0.0.1", "couchdb@127.0.0.1"})
}

func TestAddNodeAddsNode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_nodes/couchdb@111.222.333.444",
		httpmock.NewStringResponder(200, ""))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"]}`))

	if err = cluster.AddNode("111.222.333.444", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, cluster.NodesInfo.AllNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
	assert.Equal(t, cluster.NodesInfo.ClusterNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
}

func TestAddNodeRejectsToAddAlreadyAddedNode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	err = cluster.AddNode("127.0.0.1", ahr)
	assert.Error(t, err, "Node should be rejected as is already part of the cluster")
}

func TestAddNodeRejoinsNode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_nodes/couchdb@111.222.333.444",
		httpmock.NewStringResponder(200, `{
		"_id": "couchdb@111.222.333.444",
	  "_rev": "12345"}`))

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_nodes/couchdb@111.222.333.444",
		func(req *http.Request) (*http.Response, error) {
			body := struct {
				Rev string `json:"_rev"`
			}{}

			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				t.Error(err)
			}

			if body.Rev == "12345" {
				return httpmock.NewStringResponse(200, ""), nil
			}
			return nil, fmt.Errorf("Unexpected body sent to rejoin the node")
		})

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
			"all_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"]}`))

	if err = cluster.AddNode("111.222.333.444", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, cluster.NodesInfo.AllNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
	assert.Equal(t, cluster.NodesInfo.ClusterNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
}

func TestRemoveNodeRemovesNode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_all_dbs",
		httpmock.NewStringResponder(200, `["_global_changes", "testdb"]`))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/_global_changes",
		httpmock.NewStringResponder(200, `{
			"_id": "_global_changes",
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

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "41-asfefasdfw4543twf09uijkfg829438t",
			"shard_suffix": [ 46, 49, 52, 50, 53, 50, 48, 50, 53, 55, 55 ],
			"changelog": [
				[
					"add",
					"00000000-ffffffff",
					"couchdb@127.0.0.1"
				]
			],
			"by_node": {
				"couchdb@127.0.0.1": [ "00000000-ffffffff" ]
			},
			"by_range": {
				"00000000-ffffffff": ["couchdb@127.0.0.1"]
			}}`))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_nodes/couchdb@127.0.0.2",
		httpmock.NewStringResponder(200, `{"_id": "couchdb@127.0.0.2", "_rev": "1234567890asdfe"}`))

	httpmock.RegisterResponder("DELETE", "http://127.0.0.1:5986/_nodes/couchdb@127.0.0.2?rev=1234567890asdfe",
		httpmock.NewStringResponder(200, ""))

	if err = cluster.RemoveNode(NodeAt("127.0.0.2"), ahr); err != nil {
		t.Error(err)
	}
}

func TestRemoveNodeRejectsIfNodeHasReplica(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]}`))

	ahr := httpUtils.NewAuthenticatedHttpRequester("dummyuser", "dummypassword", "127.0.0.1")
	cluster, err := LoadCluster(ahr)
	if err != nil {
		t.Error(err)
	}

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_all_dbs",
		httpmock.NewStringResponder(200, `["_global_changes", "testdb"]`))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/_global_changes",
		httpmock.NewStringResponder(200, `{
			"_id": "_global_changes",
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

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5986/_dbs/testdb",
		httpmock.NewStringResponder(200, `{
			"_id": "testdb",
			"_rev": "41-asfefasdfw4543twf09uijkfg829438t",
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
				"couchdb@127.0.0.1": [ "00000000-ffffffff" ],
				"couchdb@127.0.0.2": [ "00000000-ffffffff" ]
			},
			"by_range": {
				"00000000-ffffffff": ["couchdb@127.0.0.1", "couchdb@127.0.0.2"]
			}}`))

	err = cluster.RemoveNode(NodeAt("127.0.0.2"), ahr)
	assert.Error(t, err)
}

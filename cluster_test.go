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
	cluster := LoadCluster(ahr)

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
	cluster := LoadCluster(ahr)

	httpmock.RegisterResponder("PUT", "http://127.0.0.1:5986/_nodes/couchdb@111.222.333.444",
		httpmock.NewStringResponder(200, ""))

	httpmock.RegisterResponder("GET", "http://127.0.0.1:5984/_membership",
		httpmock.NewStringResponder(200, `{
	"all_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"],
	"cluster_nodes": ["couchdb@127.0.0.1", "couchdb@111.222.333.444"]}`))

	cluster.AddNode("111.222.333.444", ahr)

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
	cluster := LoadCluster(ahr)

	err := cluster.AddNode("127.0.0.1", ahr)
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
	cluster := LoadCluster(ahr)

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

	if err := cluster.AddNode("111.222.333.444", ahr); err != nil {
		t.Error(err)
	}

	assert.Equal(t, cluster.NodesInfo.AllNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
	assert.Equal(t, cluster.NodesInfo.ClusterNodes, []string{"couchdb@127.0.0.1", "couchdb@111.222.333.444"})
}

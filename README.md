# CouchDB 2 cluster admin tool [![Build Status](https://travis-ci.org/cabify/couchdb-admin.svg?branch=master)](https://travis-ci.org/cabify/couchdb-admin)

Currently operating a CouchDB 2 clustered database is a highly manual and error prone task and that's something you definitely should avoid when it comes to databases.
`couchdb-admin` is a FOSS tool written in [Go](https://golang.org/) that operates your database via the [Oficial REST API](http://docs.couchdb.org/en/2.0.0/api/) in a command-line fashion. Performs sanity checks for you and hides all the REST communication complexities from you.

---

## Features

`couchdb-admin` works at three levels: Cluster, Node and Database:

* Cluster wide management:
  * Describe cluster: Get an overview of your cluster's current status (nodes joined, ...)
  * Add nodes: Join a node into the cluster.
  * Remove nodes: Remove a node from the cluster.
* Node management:
  * Set config values: Apply config values on your nodes. No need to restart.
  * Disable maintenance mode: Shortcut method to remove the `maintenance_mode` flag from a node.
* Database management:
  * Describe database: Get an overview of your database's shards distribution across nodes.
  * Create database: Create a database, configuring shards number and replication.
  * Replicate a shard: Design a new replica for a particular database's shard.
  * Remove a shard's replica: Free a node from holding a replica of a particular database's shard.

## Prerequisites

The `couchdb-admin` tool needs to be able to reach any of the nodes of the cluster to operate on both `5984` and `5986` ports. Additionally a configured admin role is required.

* The address where to contact the server has to be given in the `--server` argument (defaults to 127.0.0.1)
* The admin username is required into the `--admin` argument (defaults to admin)
* The admin's password is required into the `--password` argument (defaults to password)

## Examples

DISCLAIMER: The examples shown below will use the default values for `server`, `admin` and `password` arguments. Please update them accordingly when running your own.

### Cluster management

#### Describe cluster

Get an overview of your cluster's current status, which nodes form it and so on. Uses the [_membership](http://docs.couchdb.org/en/2.0.0/api/server/common.html#get--_membership) endpoint.

```
$ couchdb-admin describe_cluster

2017/06/29 13:46:26  info Describing cluster layout... server=127.0.0.1

couchdb_admin.Nodes{
    AllNodes:     {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
    ClusterNodes: {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
}
```

#### Add nodes

Following the procedure described [in the official docs](http://docs.couchdb.org/en/2.0.0/cluster/nodes.html#adding-a-node) joins a node into the cluster.
Uses the `_nodes` endpoint

After firing up a new node, located, for example, at `couch-3.couchdb2-replica-admin`:

```
$ couchdb-admin add_node --node=couch-3.couchdb2-replica-admin

2017/06/29 15:38:39  info Adding node to the cluster... node=couch-3.couchdb2-replica-admin

2017/06/29 15:38:40  info Successfully added node!  node=couch-3.couchdb2-replica-admin
```

And then, describing the cluster again...

```
$ couchdb-admin describe_cluster

2017/06/29 15:38:54  info Describing cluster layout... server=127.0.0.1

couchdb_admin.Nodes{
    AllNodes:     {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin", "couchdb@couch-3.couchdb2-replica-admin"},
    ClusterNodes: {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin", "couchdb@couch-3.couchdb2-replica-admin"},
}
```

#### Remove nodes

Again, following the procedure described [in the official docs](http://docs.couchdb.org/en/2.0.0/cluster/nodes.html#removing-a-node) it removes a node from the cluster.
Uses the `_nodes` endpoint

```
$ couchdb-admin remove_node --node=couch-3.couchdb2-replica-admin

2017/06/29 15:55:26  info Removing node...          node=couch-3.couchdb2-replica-admin

2017/06/29 15:55:26  info Checking that node does not own any shard... node=couchdb@couch-3.couchdb2-replica-admin

2017/06/29 15:55:26  info Node successfully removed! node=couch-3.couchdb2-replica-admin
```
And then, describing the cluster again...
```
$ couchdb-admin describe_cluster

2017/06/29 15:55:59  info Describing cluster layout... server=127.0.0.1

couchdb_admin.Nodes{
    AllNodes:     {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
    ClusterNodes: {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
}
```

### Node management

#### Set config values

Sets config values on a node by using the [_config](http://docs.couchdb.org/en/2.0.0/api/server/configuration.html#put--_config-section-key) endpoint.
You can check possible sections, keys and values to use on the [config reference](http://docs.couchdb.org/en/2.0.0/config-ref.html).
Changes will take effect on the fly, no need to restart!

Here we'll set the log level to debug in a particular node.

```
$ couchdb-admin set_config --section=log --key=level --value=debug --node=couch-0.couchdb2-replica-admin

2017/06/29 16:04:08  info Setting config value...   key=level node=couch-0.couchdb2-replica-admin section=log value=debug

2017/06/29 16:04:08  info New config successfully applied! key=level node=couch-0.couchdb2-replica-admin section=log value=debug
```

And in the logs you should see `[notice] 2017-06-29T12:42:21.558278Z couchdb@couch-0.couchdb2-replica-admin <0.89.0> -------- config: [log] level set to debug for reason nil`

#### Disable maintenance mode

This is a shortcut method to disable the [maintenance_mode flag](http://docs.couchdb.org/en/2.0.0/config/couchdb.html#couchdb/maintenance_mode). The reason for this method is that whenever a new node is configured to replicate a shard it is set into `maintenance_mode` to avoid unconsistent reads while it syncs.

```
$ couchdb-admin disasble_maintenance_mode --node=couch-1.couchdb2-replica-admin

2017/06/29 16:28:52  info Removing maintenance flag... node=couch-1.couchdb2-replica-admin

2017/06/29 16:28:52  info Maintenance flag successfully removed! node=couch-1.couchdb2-replica-admin
```

You should see something like this in the logs: `[notice] 2017-06-29T13:07:06.470213Z couchdb@couch-1.couchdb2-replica-admin <0.89.0> -------- config: [couchdb] maintenance_mode set to false for reason nil`

### Database management

#### Describe database

Gets a database shards ownership from the `_dbs` endpoint

```
$ couchdb-admin describe_database --db=testdb

2017/06/29 16:08:04  info Describing database...    db=testdb

&couchdb_admin.Database{
    name:   "testdb",
    config: couchdb_admin.Config{
        Id:        "testdb",
        Rev:       "1-898af88337774f013e7c14fafeb79f75",
        Shards:    {46, 49, 52, 57, 56, 50, 48, 56, 52, 50, 51},
        Changelog: {
            {"add", "00000000-1fffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "00000000-1fffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "00000000-1fffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "20000000-3fffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "20000000-3fffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "20000000-3fffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "40000000-5fffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "40000000-5fffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "40000000-5fffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "60000000-7fffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "60000000-7fffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "60000000-7fffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "80000000-9fffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "80000000-9fffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "80000000-9fffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "a0000000-bfffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "a0000000-bfffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "a0000000-bfffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "c0000000-dfffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "c0000000-dfffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "c0000000-dfffffff", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "e0000000-ffffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "e0000000-ffffffff", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "e0000000-ffffffff", "couchdb@couch-2.couchdb2-replica-admin"},
        },
        ByNode: {
            "couchdb@couch-0.couchdb2-replica-admin": {"00000000-1fffffff", "20000000-3fffffff", "40000000-5fffffff", "60000000-7fffffff", "80000000-9fffffff", "a0000000-bfffffff", "c0000000-dfffffff", "e0000000-ffffffff"},
            "couchdb@couch-1.couchdb2-replica-admin": {"00000000-1fffffff", "20000000-3fffffff", "40000000-5fffffff", "60000000-7fffffff", "80000000-9fffffff", "a0000000-bfffffff", "c0000000-dfffffff", "e0000000-ffffffff"},
            "couchdb@couch-2.couchdb2-replica-admin": {"00000000-1fffffff", "20000000-3fffffff", "40000000-5fffffff", "60000000-7fffffff", "80000000-9fffffff", "a0000000-bfffffff", "c0000000-dfffffff", "e0000000-ffffffff"},
        },
        ByRange: {
            "20000000-3fffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "40000000-5fffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "60000000-7fffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "80000000-9fffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "a0000000-bfffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "c0000000-dfffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "e0000000-ffffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "00000000-1fffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
        },
    },
}
```

Where we can see that our database has 8 shards and 3 replicas each shard. The `ByNode` key describes which shards each node is holding whilst `ByRange` describes which nodes contain which shard.
This is a special case as we have only 3 nodes, so each node has a complete copy of all the data.

#### Create a database

Creates a new database using the [PUT /{db}](http://docs.couchdb.org/en/2.0.0/api/database/common.html#put--db) endpoint.

```
$ couchdb-admin create_db --db=mydb --shards=3 --replicas=2

2017/06/29 16:15:42  info Creating database...      db=mydb replicas=2 shards=3

2017/06/29 16:15:43  info Database successfully created! db=mydb replicas=2 shards=3
```

Now we can see its layout across nodes.

```
$ couchdb-admin describe_db --db=mytdb

2017/06/29 16:16:43  info Describing database...    db=mydb

&couchdb_admin.Database{
    name:   "mydb",
    config: couchdb_admin.Config{
        Id:        "mydb",
        Rev:       "1-73e3eedeb4a3304842633c8f695e36c0",
        Shards:    {46, 49, 52, 57, 56, 55, 52, 48, 56, 51, 54},
        Changelog: {
            {"add", "00000000-55555554", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "00000000-55555554", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-1.couchdb2-replica-admin"},
        },
        ByNode: {
            "couchdb@couch-0.couchdb2-replica-admin": {"55555555-aaaaaaa9", "aaaaaaaa-ffffffff"},
            "couchdb@couch-1.couchdb2-replica-admin": {"00000000-55555554", "aaaaaaaa-ffffffff"},
            "couchdb@couch-2.couchdb2-replica-admin": {"00000000-55555554", "55555555-aaaaaaa9"},
        },
        ByRange: {
            "55555555-aaaaaaa9": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "aaaaaaaa-ffffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin"},
            "00000000-55555554": {"couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
        },
    },
}
```

#### Replicate a shard

Configures a node to also be a replica for a particular shard. It follows the procedure described [in the official docs](http://docs.couchdb.org/en/2.0.0/cluster/sharding.html?highlight=scaling%20out#scaling-out).

```
$ couchdb-admin replicate --shard=55555555-aaaaaaa9 --db=mydb --replica=couch-1.couchdb2-replica-admin

2017/06/29 16:21:16  info Replicating shard...      db=mydb replica=couch-1.couchdb2-replica-admin shard=55555555-aaaaaaa9

2017/06/29 16:21:16  info Shard successfully replicated db=mydb replica=couch-1.couchdb2-replica-admin shard=55555555-aaaaaaa9

2017/06/29 16:21:16  warn Node was sent into maintenance!. Remember to reenable it once it catches up with changes node=couch-1.couchdb2-replica-admin
```

And now describing the db again we can see that the shard 555... is on all nodes and taht couch-1 node holds all three shards!

```
$ couchdb-admin describe_db --db=mydb

2017/06/29 16:21:57  info Describing database...    db=mydb

&couchdb_admin.Database{
    name:   "mydb",
    config: couchdb_admin.Config{
        Id:        "mydb",
        Rev:       "2-5c965ef57142374eda39b6f4718b6298",
        Shards:    {46, 49, 52, 57, 56, 55, 52, 48, 56, 51, 54},
        Changelog: {
            {"add", "00000000-55555554", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "00000000-55555554", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-1.couchdb2-replica-admin"},
        },
        ByNode: {
            "couchdb@couch-0.couchdb2-replica-admin": {"55555555-aaaaaaa9", "aaaaaaaa-ffffffff"},
            "couchdb@couch-1.couchdb2-replica-admin": {"00000000-55555554", "aaaaaaaa-ffffffff", "55555555-aaaaaaa9"},
            "couchdb@couch-2.couchdb2-replica-admin": {"00000000-55555554", "55555555-aaaaaaa9"},
        },
        ByRange: {
            "55555555-aaaaaaa9": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin"},
            "aaaaaaaa-ffffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin"},
            "00000000-55555554": {"couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
        },
    },
}
```

BEWARE!!!: The node receiving the new replica is automatically set into [maintenance mode](http://docs.couchdb.org/en/2.0.0/config/couchdb.html#couchdb/maintenance_mode). You should check the logs for pending changes and once it finishes syncing [disable maintenance mode](https://github.com/cabify/couchdb-admin#disable-maintenance-mode) so that it participates in reads again.

#### Remove a shard's replica

Configures a node to stop being a replica for a particular shard. It follows the procedure described [in the official docs](http://docs.couchdb.org/en/2.0.0/cluster/sharding.html?highlight=scaling%20out#moving-shards).

```
$ couchdb-admin remove_replica --db=mydb --shard=55555555-aaaaaaa9 --from=couch-1.couchdb2-replica-admin

2017/06/29 16:34:47  info Removing shard ownership... db=mydb replica=couch-1.couchdb2-replica-admin shard=55555555-aaaaaaa9

2017/06/29 16:34:47  info Replica shard successfully removed! db=mydb replica=couch-1.couchdb2-replica-admin shard=55555555-aaaaaaa9
```

And describing the database layout again we see that couch-1 is no longer replicating shard 555...

```
$ couchdb-admin describe_db --db=mydb

2017/06/29 16:35:48  info Describing database...    db=mydb

&couchdb_admin.Database{
    name:   "mydb",
    config: couchdb_admin.Config{
        Id:        "mydb",
        Rev:       "3-64420ae508c9ad647f615d5d823ec9b8",
        Shards:    {46, 49, 52, 57, 56, 55, 52, 48, 56, 51, 54},
        Changelog: {
            {"add", "00000000-55555554", "couchdb@couch-1.couchdb2-replica-admin"},
            {"add", "00000000-55555554", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "55555555-aaaaaaa9", "couchdb@couch-2.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-0.couchdb2-replica-admin"},
            {"add", "aaaaaaaa-ffffffff", "couchdb@couch-1.couchdb2-replica-admin"},
        },
        ByNode: {
            "couchdb@couch-0.couchdb2-replica-admin": {"55555555-aaaaaaa9", "aaaaaaaa-ffffffff"},
            "couchdb@couch-1.couchdb2-replica-admin": {"00000000-55555554", "aaaaaaaa-ffffffff"},
            "couchdb@couch-2.couchdb2-replica-admin": {"00000000-55555554", "55555555-aaaaaaa9"},
        },
        ByRange: {
            "00000000-55555554": {"couchdb@couch-1.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "55555555-aaaaaaa9": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-2.couchdb2-replica-admin"},
            "aaaaaaaa-ffffffff": {"couchdb@couch-0.couchdb2-replica-admin", "couchdb@couch-1.couchdb2-replica-admin"},
        },
    },
}
```

## Vendoring

`couchdb-admin` currently uses [Glide](http://glide.sh/) for vendoring.

## Developing couchdb-admin

If you wish to work on couchdb-admin you'll first need Go installed (version 1.8+ is required). Make sure you have Go properly installed, including setting up your GOPATH.

Next, clone this repository into $GOPATH/src/github.com/cabify/couchdb-admin. Then enter into the directory `cli/couchdb-admin` and type:
```
$ make all
```

This will generate a binary file `couchdb-admin` which you can now play with. In case you are running on macOS type `make darwin` instead.

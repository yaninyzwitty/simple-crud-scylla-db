package main

import (
"fmt"

    "github.com/gocql/gocql"

)

func main() {
var cluster = gocql.NewCluster("node-0.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud", "node-1.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud", "node-2.aws-eu-central-1.c01a50d3bacf78706f92.clusters.scylla.cloud")
cluster.Authenticator = gocql.PasswordAuthenticator{Username: "scylla", Password: "t5IJAg9zdGYB6NL"}
cluster.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy("AWS_EU_CENTRAL_1")

    var session, err = cluster.CreateSession()
    if err != nil {
    	panic("Failed to connect to cluster")
    }

    defer session.Close()

    var query = session.Query("SELECT * FROM myKeyspace.monkey_species WHERE species='' ")

    if rows, err := query.Iter().SliceMap(); err == nil {
    	for _, row := range rows {
    		fmt.Printf("%v\n", row)
    	}
    } else {
    	panic("Query error: " + err.Error())
    }

}

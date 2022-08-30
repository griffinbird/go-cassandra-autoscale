package utils

import (
	"context"
	"crypto/tls"
	"log"
	"strconv"
	"time"

	"github.com/gocql/gocql"
)

// GetSession connects to Cassandra
func GetSession(cosmosCassandraContactPoint, cosmosCassandraPort, cosmosCassandraUser, cosmosCassandraPassword string) *gocql.Session {
	clusterConfig := gocql.NewCluster(cosmosCassandraContactPoint)
	port, err := strconv.Atoi(cosmosCassandraPort)
	if err != nil {
		log.Fatal(err)
	}
	clusterConfig.Port = port
	clusterConfig.ProtoVersion = 4
	clusterConfig.Authenticator = gocql.PasswordAuthenticator{Username: cosmosCassandraUser, Password: cosmosCassandraPassword}
	clusterConfig.SslOpts = &gocql.SslOptions{Config: &tls.Config{MinVersion: tls.VersionTLS12}}

	clusterConfig.ConnectTimeout = 10 * time.Second
	clusterConfig.Timeout = 10 * time.Second
	clusterConfig.DisableInitialHostLookup = true

	//If you setup geo-replication you need to set this property for the write region.
	clusterConfig.PoolConfig.HostSelectionPolicy = gocql.DCAwareRoundRobinPolicy("Australia East")

	// uncomment if you want to track time taken for individual queries
	//clusterConfig.QueryObserver = timer{}

	// uncomment if you want to track time taken for each connection to Cassandra
	//clusterConfig.ConnectObserver = timer{}

	session, err := clusterConfig.CreateSession()
	if err != nil {
		log.Fatal("Failed to connect to Azure Cosmos DB", err)
	}

	return session
}

// ExecuteQuery executes a query and returns an error if any
func ExecuteQuery(query string, session *gocql.Session) error {
	return session.Query(query).Exec()
}

type Timer struct {
	queryTime *time.Duration
}
func CreateTimer() Timer {
	t := new(time.Duration)
	return Timer{queryTime:t}
}

func (t Timer) ObserveQuery(ctx context.Context, oq gocql.ObservedQuery) {
	//log.Printf("Time taken for %v", time.Since(oq.Start))
	tmp := time.Since(oq.Start)
	*t.queryTime = tmp
}

func (t Timer) Duration()time.Duration{
	return *t.queryTime
}

/*func (t timer) ObserveConnect(oc gocql.ObservedConnect) {
	if oc.Err != nil {
		log.Println("Connection error: ", oc.Err)
	}
	log.Printf("Time taken for connection = %v ", time.Since(oc.Start))
}*/

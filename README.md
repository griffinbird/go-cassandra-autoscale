# Developing a Go app with Cassandra API using Azure Cosmos DB (`gocql` Driver)

[Azure Cosmos DB]((https://docs.microsoft.com/azure/cosmos-db/introduction?WT.mc_id=cassandrago-github-abhishgu)) is a globally distributed multi-model database. One of the supported APIs is the [Cassandra API](https://docs.microsoft.com/azure/cosmos-db/cassandra-introduction?WT.mc_id=cassandrago-github-abhishgu). 

The code included in this sample is intended to get you quickly started with a Go application that connects to Azure Cosmos DB with the Cassandra API. It walks you through creation of keyspace, table, inserting and querying the data.

## Creating load on Cosmso DB Cassandra API
This sample uses go-routines to create load on the Cosmos DB Cassandra API instance exhibiting 429's error/trottled request.

You can uncomment out line 14 and 15 in setup.go to move between 400 RU provisoned throughput and 1000 RU autocale. 

Autoscale will solve the 429 errors when set at 400 RU

The other way to remove the 429 errors is to enable service side retries on the Cosmos DB Cassandra API account.

```shell
export COSMOSDB_CASSANDRA_CONTACT_POINT=<Contact Point for Azure Cosmos DB Cassandra API>
export COSMOSDB_CASSANDRA_PORT=<Port for Azure Cosmos DB Cassandra API>
export COSMOSDB_CASSANDRA_USER=<Username for Azure Cosmos DB Cassandra API>
export COSMOSDB_CASSANDRA_PASSWORD=<password for Azure Cosmos DB Cassandra API>
```

## Prerequisites

Before you can run this sample, you must have the following prerequisites:

- An Azure account with an active subscription. [Create one for free](https://azure.microsoft.com/free/?WT.mc_id=cassandrago-github-abhishgu). Or [try Azure Cosmos DB for free](https://azure.microsoft.com/try/cosmosdb/?WT.mc_id=cassandrago-github-abhishgu) without an Azure subscription.
- [Go](https://golang.org/) installed on your computer, and a working knowledge of Go.
- [Git](https://git-scm.com/downloads).

## Running this sample

1. Clone this repository using `git clone https://github.com/griffinbird/go-cassandra-autoscale`

2. Change directories to the repo using `cd go-cassandra-autoscale`

3. Set environment variables. Either in the shell or via .env

### Enable SSR
```shell
az cosmosdb update --name accountname --resource-group resourcegroupname --capabilities EnableCassandra DisableRateLimitingResponses
```
### Checking if SSR is enabled
```shell
az cosmosdb show --name accountname --resource-group resourcegroupname
```
### Enable SSR
```shell
az cosmosdb update --name accountname --resource-group resourcegroupname --capabilities EnableCassandra DisableRateLimitingResponses
```

### Disable SSR
```shell
az cosmosdb update --name accountname --resource-group resourcegroupname --capabilities EnableCassandra <- disable SSR
```

4. Run the application

```shell
go run main.go
```

## More information

- [Azure Cosmos DB](https://docs.microsoft.com/azure/cosmos-db/introduction?WT.mc_id=cassandrago-github-abhishgu)
- [Azure Cosmos DB for Cassandra API](https://docs.microsoft.com/azure/cosmos-db/cassandra-introduction?WT.mc_id=cassandrago-github-abhishgu)
- [gocql - Cassandra Go driver](https://github.com/gocql/gocql)
- [gocql reference](https://godoc.org/github.com/gocql/gocql)
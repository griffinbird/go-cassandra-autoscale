package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	//"time"

	"github.com/joho/godotenv"
	"github.com/griffinbird/go-cassandra-autoscale/model"
	"github.com/griffinbird/go-cassandra-autoscale/operations"
	"github.com/griffinbird/go-cassandra-autoscale/utils"
)

var (
	cosmosCassandraContactPoint string
	cosmosCassandraPort         string
	cosmosCassandraUser         string
	cosmosCassandraPassword     string
)

var cities = []string{"New Delhi", "New York", "Bangalore", "Seattle"}

const (
	keyspace = "user_profile"
	table    = "user"
)

func init() {
	cosmosCassandraContactPoint = os.Getenv("COSMOSDB_CASSANDRA_CONTACT_POINT")
	cosmosCassandraPort = os.Getenv("COSMOSDB_CASSANDRA_PORT")
	cosmosCassandraUser = os.Getenv("COSMOSDB_CASSANDRA_USER")
	cosmosCassandraPassword = os.Getenv("COSMOSDB_CASSANDRA_PASSWORD")
	if cosmosCassandraContactPoint == "" || cosmosCassandraUser == "" || cosmosCassandraPassword == "" {
		//log.Printf("cannot find the mandatory environment variables..")
		err := godotenv.Load(".env")
		//log.Printf("checking for .env file..")
		if err != nil {
			log.Printf("error loading .env file")
		}
		cosmosCassandraContactPoint = os.Getenv("COSMOSDB_CASSANDRA_CONTACT_POINT")
		cosmosCassandraPort = os.Getenv("COSMOSDB_CASSANDRA_PORT")
		cosmosCassandraUser = os.Getenv("COSMOSDB_CASSANDRA_USER")
		cosmosCassandraPassword = os.Getenv("COSMOSDB_CASSANDRA_PASSWORD")	
	}
	if cosmosCassandraContactPoint == "" || cosmosCassandraUser == "" || cosmosCassandraPassword == "" {
		log.Fatal("missing mandatory environment variables")
	}
}

func main() {
	session := utils.GetSession(cosmosCassandraContactPoint, cosmosCassandraPort, cosmosCassandraUser, cosmosCassandraPassword)
	defer session.Close()

	//log.Println("Connected to Azure Cosmos DB")

	operations.DropKeySpaceIfExists(keyspace, session)
	operations.CreateKeySpace(keyspace, session)
	operations.CreateUserTable(keyspace, table, session)

	log.Println("*** Inserting users... ***")

	var parallel = true // 200 insertions across 10 go routines. False = 25 users
	//var itemsToInsert = 100000000

	if !parallel {
		for i := 1; i <= 25; i++ {
			name := "user-" + strconv.Itoa(i)
			insertRequestCharge, QueryDuration, insertError := operations.InsertUser(keyspace, table, session, model.User{ID: i, Name: name, City: cities[rand.Intn(len(cities))]})

			if insertError != nil {
				log.Fatal("Failed to create user: ", insertError)
			}

			log.Printf("User created. [Request Units: %v] [Duraton: %v]", insertRequestCharge, QueryDuration)
		}
	} else {
		var threads = 17
		var insertsPerThread = 10
		var wg sync.WaitGroup
		wg.Add(threads)
		for i := 0; i < threads; i++ {
			go func(i int) {
				for j := 0; j < insertsPerThread; j++ {
					var index = i*insertsPerThread + j

					name := "user-" + strconv.Itoa(index)
					insertRequestCharge, QueryDuartion, insertError := operations.InsertUser(keyspace, table, session, model.User{ID: index, Name: name, City: cities[rand.Intn(len(cities))]})

					if insertError != nil {
						log.Fatal("Failed to create user: ", insertError)
					}

					log.Printf("User %v created. [Request Units: %v] [Duraton: %v]", index, insertRequestCharge, QueryDuartion)
					// Sleep between each insert
					//time.Sleep(time.Second * 1)
				}
				wg.Done()
			}(i)
		}
		wg.Wait()
	}
}

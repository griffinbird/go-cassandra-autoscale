package operations

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/griffinbird/go-cassandra-autoscale/model"
	"github.com/griffinbird/go-cassandra-autoscale/utils"
	"github.com/griffinbird/go-cassandra-autoscale/utils/scopedRetry"
	"github.com/gocql/gocql"
)

const (
	createQuery       = "INSERT INTO %s.%s (user_id, user_name, user_bcity) VALUES (?,?,?)"
	selectQuery       = "SELECT * FROM %s.%s where user_id = ?"
	findAllUsersQuery = "SELECT * FROM %s.%s"
)

// InsertUser creates an entry(row) in a table
func InsertUser(keyspace, table string, session *gocql.Session, user model.User) (float64, time.Duration, error) {
	retryPolicy := scopedRetry.NewScopedCosmosRetryPolicy(10, fmt.Sprintf("User %d", user.ID))
	timer := utils.CreateTimer()
	query := session.Query(fmt.Sprintf(createQuery, keyspace, table)).Bind(user.ID, user.Name, user.City).RetryPolicy(retryPolicy).Observer(timer)
	charge, err := ExecWithRequestCharge(query)
	return charge,timer.Duration(),err
}

//FindUser with request charge
func FindUser(keyspace, table string, id int, session *gocql.Session) (model.User, float64) {
	var userid int
	var name, city string

	query := session.Query(fmt.Sprintf(selectQuery, keyspace, table)).Bind(id)

	requestCharge, err := ScanWithRequestCharge(query, &userid, &name, &city)

	if err != nil {
		if err == gocql.ErrNotFound {
			log.Printf("User with id %v does not exist\n", id)
		} else {
			log.Printf("Failed to find user with id %v - %v\n", id, err)
		}
	}
	return model.User{ID: userid, Name: name, City: city}, requestCharge
}

// FindAllUsers gets all users
func FindAllUsers(keyspace, table string, session *gocql.Session) ([]model.User, float64) {
	var users []model.User

	iter := session.Query(fmt.Sprintf(findAllUsersQuery, keyspace, table)).Iter()

	requestCharge := RequestChargeFromIter(iter)

	results, _ := iter.SliceMap()
	for _, u := range results {
		users = append(users, mapToUser(u))
	}

	iter.Close()

	return users, requestCharge
}

func mapToUser(m map[string]interface{}) model.User {
	id, _ := m["user_id"].(int)
	name, _ := m["user_name"].(string)
	city, _ := m["user_bcity"].(string)

	return model.User{ID: id, Name: name, City: city}
}

// RequestCharge Helpers
func ScanWithRequestCharge(q *gocql.Query, dest ...interface{}) (float64, error) {
	iter := q.Iter()
	if iter.NumRows() == 0 {
		return 0, gocql.ErrNotFound
	}
	if !iter.Scan(dest...) {
		return 0, iter.Close()
	}

	requestCharge := RequestChargeFromIter(iter)

	return requestCharge, iter.Close()
}

func ExecWithRequestCharge(q *gocql.Query) (float64, error) {
	iter := q.Iter()
	requestCharge := RequestChargeFromIter(iter)
	err := iter.Close()
	if err != nil {
		return -1, err
	}
	return requestCharge, err
}

func RequestChargeFromIter(iter *gocql.Iter) float64 {
	if iter == nil {
		return -3
	}
	customPayload := iter.GetCustomPayload()
	if customPayload == nil {
		return -2
	}
	requestCharge, hasRequestCharge := customPayload["RequestCharge"]
	if !hasRequestCharge || requestCharge == nil {
		return -1
	}
	return Float64frombytes(requestCharge)
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

package scopedRetry

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gocql/gocql"
)
// TODO - props to Abhiseek(?)
// ScopedCosmosRetryPolicy implements gcql.RetryPolicy. Retires only if query attempts are less than or equal to max retry config or max retry config is set to -1 (infinite retries). For RequestErrReadTimeout, RequestErrUnavailable, RequestErrWriteTimeout the request is retried immediately. For rate limited (429) errors, retries are eexecuted after waiting for a duration of RetryAfterMs. If not available, time between retries is increased as per GrowingBackOffTimeMs. If MaxRetryCount is -1 (inifinite) then retry back-off is as per FixedBackOffTimeMs
type ScopedCosmosRetryPolicy struct {
	MaxRetryCount        int
	FixedBackOffTimeMs   int
	GrowingBackOffTimeMs int
	Context              string
	numAttempts          int
}

const defaultGrowingBackOffTimeMs = 1000
const defaultFixedBackOffTimeMs = 5000

// NewScopedCosmosRetryPolicy returns a ScopedCosmosRetryPolicy with default values for growing and fixed back-off time (in ms)
func NewScopedCosmosRetryPolicy(maxRetryCount int, context string) *ScopedCosmosRetryPolicy {
	return &ScopedCosmosRetryPolicy{MaxRetryCount: maxRetryCount, FixedBackOffTimeMs: defaultFixedBackOffTimeMs, GrowingBackOffTimeMs: defaultGrowingBackOffTimeMs, Context: context}
}

// Attempt decides whether to retry or not. Retries only if query attempts are less than or equal to max retry config or max retry config is set to -1 (infinite retries)
func (crp *ScopedCosmosRetryPolicy) Attempt(rq gocql.RetryableQuery) bool {
	crp.numAttempts = rq.Attempts()
	return rq.Attempts() <= crp.MaxRetryCount || crp.MaxRetryCount == -1
}

// GetRetryType determines the RetryType. In case of rate limiting (429), it parses the error message to get RetryAfterMs
func (crp *ScopedCosmosRetryPolicy) GetRetryType(err error) gocql.RetryType {

	switch err.(type) {
	default:
		retryAfterMs := crp.getRetryAfterMs(err.Error())
		if retryAfterMs == -1 {
			return gocql.Rethrow
		}
		log.Printf("%s - 429 TooManyRequests - Waiting %v", crp.Context, retryAfterMs)
		time.Sleep(retryAfterMs)
		return gocql.Retry
	case *gocql.RequestErrReadTimeout:
		return gocql.Retry
	case *gocql.RequestErrUnavailable:
		return gocql.Retry
	case *gocql.RequestErrWriteTimeout:
		return gocql.Retry
	}
}

const rateLimitingErrPart = "TooManyRequests (429)"
const retryAfterKey = "RetryAfterMs"

const growingBackOffSaltMillis = 2000

/*
		Request rate is large: ActivityID=c268afb6-7367-4ff8-b06b-b7e2d1269f55, RetryAfterMs=304, Additional details='Response status code does not indicate success: TooManyRequests (429); Substatus: 3200; ActivityId: c268afb6-7367-4ff8-b06b-b7e2d1269f55; Reason: ({
	  "Errors": [
	    "Request rate is large. More Request Units may be needed, so no changes were made. Please retry this request later. Learn more: http://aka.ms/cosmosdb-error-429"
	  ]
	});
*/
func (crp *ScopedCosmosRetryPolicy) getRetryAfterMs(errMsg string) time.Duration {
	// if rate limiting error
	if strings.Contains(errMsg, rateLimitingErrPart) {
		parts := strings.Split(errMsg, ",")
		retryPart := parts[1]
		retryAfterMs := strings.Split(retryPart, "=")

		// should be RetryAfterMs
		if strings.TrimSpace(retryAfterMs[0]) == retryAfterKey {
			r, _ := strconv.Atoi(retryAfterMs[1])
			return time.Duration(r) * time.Millisecond
		}
		//if RetryAfterMs is not available

		// finite max retry count - use fix backoff retry time
		if crp.MaxRetryCount > -1 {
			return time.Duration(crp.FixedBackOffTimeMs) * time.Millisecond
		}

		// in case of infinite max retry count - use exponentially growing backoff retry time
		return time.Duration((crp.GrowingBackOffTimeMs*crp.numAttempts + rand.Intn(growingBackOffSaltMillis))) * time.Millisecond
	}

	return -1
}

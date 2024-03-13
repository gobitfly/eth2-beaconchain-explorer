package ratelimit

import (
	"context"
	"database/sql"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
	"eth2-exporter/utils"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type TimeWindow string

const (
	SecondTimeWindow = "second"
	HourTimeWindow   = "hour"
	MonthTimeWindow  = "month"
)

const (
	HeaderRateLimitLimit     = "ratelimit-limit"     // the rate limit ceiling that is applicable for the current request
	HeaderRateLimitRemaining = "ratelimit-remaining" // the number of requests left for the current rate-limit window
	HeaderRateLimitReset     = "ratelimit-reset"     // the number of seconds until the quota resets
	HeaderRateLimitWindow    = "ratelimit-window"    // what window the ratelimit represents
	HeaderRetryAfter         = "retry-after"         // the number of seconds until the quota resets, same as HeaderRateLimitReset, RFC 7231, 7.1.3

	HeaderRateLimitRemainingSecond = "x-ratelimit-remaining-second" // the number of requests left for the current rate-limit window
	HeaderRateLimitRemainingMinute = "x-ratelimit-remaining-minute" // the number of requests left for the current rate-limit window
	HeaderRateLimitRemainingHour   = "x-ratelimit-remaining-hour"   // the number of requests left for the current rate-limit window
	HeaderRateLimitRemainingDay    = "x-ratelimit-remaining-day"    // the number of requests left for the current rate-limit window
	HeaderRateLimitRemainingMonth  = "x-ratelimit-remaining-month"  // the number of requests left for the current rate-limit window

	HeaderRateLimitLimitSecond = "x-ratelimit-limit-second" // the rate limit ceiling that is applicable for the current user
	HeaderRateLimitLimitMinute = "x-ratelimit-limit-minute" // the rate limit ceiling that is applicable for the current user
	HeaderRateLimitLimitHour   = "x-ratelimit-limit-hour"   // the rate limit ceiling that is applicable for the current user
	HeaderRateLimitLimitDay    = "x-ratelimit-limit-day"    // the rate limit ceiling that is applicable for the current user
	HeaderRateLimitLimitMonth  = "x-ratelimit-limit-month"  // the rate limit ceiling that is applicable for the current user

	DefaultRateLimitSecond = 2   // RateLimit per second if no ratelimits are set in database
	DefaultRateLimitHour   = 500 // RateLimit per second if no ratelimits are set in database
	DefaultRateLimitMonth  = 0   // RateLimit per second if no ratelimits are set in database

	FallbackRateLimitSecond = 20 // RateLimit per second for when redis is offline
	FallbackRateLimitBurst  = 20 // RateLimit burst for when redis is offline

	defaultBucket = "default" // if no bucket is set for a route, use this one

	statsTruncateDuration = time.Hour * 1 // ratelimit-stats are truncated to this duration
)

var NoKeyRateLimit = &RateLimit{
	Second: DefaultRateLimitSecond,
	Hour:   DefaultRateLimitHour,
	Month:  DefaultRateLimitMonth,
}

var updateInterval = time.Second * 60 // how often to update ratelimits, weights and stats

var FreeRatelimit = &RateLimit{
	Second: DefaultRateLimitSecond,
	Hour:   DefaultRateLimitHour,
	Month:  DefaultRateLimitMonth,
}

var redisClient *redis.Client
var redisIsHealthy atomic.Bool

var lastRateLimitUpdateKeys = time.Unix(0, 0)       // guarded by lastRateLimitUpdateMu
var lastRateLimitUpdateRateLimits = time.Unix(0, 0) // guarded by lastRateLimitUpdateMu
var lastRateLimitUpdateMu = &sync.Mutex{}

var fallbackRateLimiter = NewFallbackRateLimiter() // if redis is offline, use this rate limiter

var initializedWg = &sync.WaitGroup{} // wait for everything to be initialized before serving requests

var rateLimitsMu = &sync.RWMutex{}
var rateLimits = map[string]*RateLimit{}        // guarded by rateLimitsMu
var rateLimitsByUserId = map[int64]*RateLimit{} // guarded by rateLimitsMu
var userIdByApiKey = map[string]int64{}         // guarded by rateLimitsMu

var weightsMu = &sync.RWMutex{}
var weights = map[string]int64{}  // guarded by weightsMu
var buckets = map[string]string{} // guarded by weightsMu

var logger = logrus.StandardLogger().WithField("module", "ratelimit")

type DbEntry struct {
	Date     time.Time
	UserId   int64
	ApiKey   string
	Endpoint string
	Count    int64
}

type RateLimit struct {
	Second int64
	Hour   int64
	Month  int64
}

type RateLimitResult struct {
	BlockRequest  bool
	Time          time.Time
	Weight        int64
	Route         string
	IP            string
	Key           string
	IsValidKey    bool
	UserId        int64
	RedisKeys     []RedisKey
	RedisStatsKey string
	RateLimit     *RateLimit

	Limit       int64
	LimitSecond int64
	LimitMinute int64
	LimitHour   int64
	LimitDay    int64
	LimitMonth  int64

	Remaining       int64
	RemainingSecond int64
	RemainingMinute int64
	RemainingHour   int64
	RemainingDay    int64
	RemainingMonth  int64

	Reset  int64
	Bucket string
	Window TimeWindow
}

type RedisKey struct {
	Key      string
	ExpireAt time.Time
}

type ApiProduct struct {
	Name          string    `db:"name"`
	StripePriceID string    `db:"stripe_price_id"`
	Second        int64     `db:"second"`
	Hour          int64     `db:"hour"`
	Month         int64     `db:"month"`
	ValidFrom     time.Time `db:"valid_from"`
}

type responseWriterDelegator struct {
	http.ResponseWriter
	written     int64
	status      int
	wroteHeader bool
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
	r.wroteHeader = true
}

func (r *responseWriterDelegator) Status() int {
	return r.status
}

var DefaultRequestFilter = func(req *http.Request) bool {
	if req.URL == nil || !strings.HasPrefix(req.URL.Path, "/api") || strings.HasPrefix(req.URL.Path, "/api/i/") || strings.HasPrefix(req.URL.Path, "/api/v1/docs/") || strings.HasPrefix(req.URL.Path, "/api/v2/docs/") {
		return false
	}
	return true
}

var requestFilter = DefaultRequestFilter
var requestFilterMu = &sync.RWMutex{}

func SetRequestFilter(filter func(req *http.Request) bool) {
	requestFilterMu.Lock()
	defer requestFilterMu.Unlock()
	requestFilter = filter
}

func GetRequestFilter() func(req *http.Request) bool {
	requestFilterMu.RLock()
	defer requestFilterMu.RUnlock()
	return requestFilter
}

var maxBadRequestWeight int64 = 1

func SetMaxBadRquestWeight(weight int64) {
	atomic.StoreInt64(&maxBadRequestWeight, weight)
}

func GetMaxBadRquestWeight() int64 {
	return atomic.LoadInt64(&maxBadRequestWeight)
}

// Init initializes the RateLimiting middleware, the rateLimiting middleware will not work without calling Init first. The second parameter is a function the will get called on every request, it will only apply ratelimiting to requests when this func returns true.
func Init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisSessionStoreEndpoint,
		ReadTimeout: time.Second * 3,
	})

	updateInterval = utils.Config.Frontend.RatelimitUpdateInterval
	if updateInterval < time.Second {
		logger.Warnf("updateInterval is below 1s, setting to 60s")
		updateInterval = time.Second * 60
	}

	initializedWg.Add(3)

	go func() {
		firstRun := true
		for {
			err := updateWeights(firstRun)
			if err != nil {
				logger.WithError(err).Errorf("error updating weights")
				time.Sleep(time.Second * 2)
				continue
			}
			if firstRun {
				initializedWg.Done()
				firstRun = false
			}
			time.Sleep(updateInterval)
		}
	}()
	go func() {
		firstRun := true
		for {
			err := updateRateLimits()
			if err != nil {
				logger.WithError(err).Errorf("error updating ratelimits")
				time.Sleep(time.Second * 2)
				continue
			}
			if firstRun {
				initializedWg.Done()
				firstRun = false
			}
			time.Sleep(updateInterval)
		}
	}()
	go func() {
		firstRun := true
		for {
			err := updateRedisStatus()
			if err != nil {
				logger.WithError(err).Errorf("error checking redis")
				time.Sleep(time.Second * 1)
				continue
			}
			if firstRun {
				initializedWg.Done()
				firstRun = false
			}
			time.Sleep(time.Second * 1)
		}
	}()

	initializedWg.Wait()
}

// HttpMiddleware returns an http.Handler that can be used as middleware to RateLimit requests. If redis is offline, it will use a fallback rate limiter.
func HttpMiddleware(next http.Handler) http.Handler {
	initializedWg.Wait()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f := GetRequestFilter()
		if !f(r) {
			next.ServeHTTP(w, r)
			return
		}

		if !redisIsHealthy.Load() {
			fallbackRateLimiter.Handle(w, r, next.ServeHTTP)
			return
		}

		rl, err := rateLimitRequest(r)
		if err != nil {
			// just serve the request if there is a problem with getting the rate limit
			logger.WithFields(logrus.Fields{"error": err}).Errorf("error getting rate limit")
			next.ServeHTTP(w, r)
			return
		}

		// logrus.WithFields(logrus.Fields{"route": rl.Route, "key": rl.Key, "limit": rl.Limit, "remaining": rl.Remaining, "reset": rl.Reset, "window": rl.Window, "validKey": rl.IsValidKey}).Infof("rateLimiting")

		w.Header().Set(HeaderRateLimitLimit, strconv.FormatInt(rl.Limit, 10))
		w.Header().Set(HeaderRateLimitRemaining, strconv.FormatInt(rl.Remaining, 10))
		w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(rl.Reset, 10))

		w.Header().Set(HeaderRateLimitWindow, string(rl.Window))

		w.Header().Set(HeaderRateLimitLimitMonth, strconv.FormatInt(rl.LimitMonth, 10))
		w.Header().Set(HeaderRateLimitLimitDay, strconv.FormatInt(rl.LimitDay, 10))
		w.Header().Set(HeaderRateLimitLimitHour, strconv.FormatInt(rl.LimitHour, 10))
		w.Header().Set(HeaderRateLimitLimitMinute, strconv.FormatInt(rl.LimitMinute, 10))
		w.Header().Set(HeaderRateLimitLimitSecond, strconv.FormatInt(rl.LimitSecond, 10))

		w.Header().Set(HeaderRateLimitRemainingMonth, strconv.FormatInt(rl.RemainingMonth, 10))
		w.Header().Set(HeaderRateLimitRemainingDay, strconv.FormatInt(rl.RemainingDay, 10))
		w.Header().Set(HeaderRateLimitRemainingHour, strconv.FormatInt(rl.RemainingHour, 10))
		w.Header().Set(HeaderRateLimitRemainingMinute, strconv.FormatInt(rl.RemainingMinute, 10))
		w.Header().Set(HeaderRateLimitRemainingSecond, strconv.FormatInt(rl.RemainingSecond, 10))

		if rl.BlockRequest {
			w.Header().Set(HeaderRetryAfter, strconv.FormatInt(rl.Reset, 10))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			err = postRateLimit(rl, http.StatusTooManyRequests)
			if err != nil {
				logger.WithFields(logrus.Fields{"error": err}).Errorf("error calling postRateLimit")
			}
			return
		}

		d := &responseWriterDelegator{ResponseWriter: w}
		next.ServeHTTP(d, r)
		err = postRateLimit(rl, d.Status())
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Errorf("error calling postRateLimit")
		}
	})
}

// updateWeights gets the weights and buckets from postgres and updates the weights and buckets maps.
func updateWeights(firstRun bool) error {
	start := time.Now()
	defer func() {
		logger.WithField("duration", time.Since(start)).Infof("updateWeights")
		metrics.TaskDuration.WithLabelValues("ratelimit_updateWeights").Observe(time.Since(start).Seconds())
	}()

	dbWeights := []struct {
		Endpoint  string    `db:"endpoint"`
		Weight    int64     `db:"weight"`
		Bucket    string    `db:"bucket"`
		ValidFrom time.Time `db:"valid_from"`
	}{}
	err := db.FrontendWriterDB.Select(&dbWeights, "SELECT DISTINCT ON (endpoint) endpoint, bucket, weight, valid_from FROM api_weights WHERE valid_from <= NOW() ORDER BY endpoint, valid_from DESC")
	if err != nil {
		return err
	}
	weightsMu.Lock()
	defer weightsMu.Unlock()
	oldWeights := weights
	oldBuckets := buckets
	weights = make(map[string]int64, len(dbWeights))
	for _, w := range dbWeights {
		weights[w.Endpoint] = w.Weight
		if !firstRun && oldWeights[w.Endpoint] != weights[w.Endpoint] {
			logger.WithFields(logrus.Fields{"endpoint": w.Endpoint, "weight": w.Weight, "oldWeight": oldWeights[w.Endpoint]}).Infof("weight changed")
		}
		buckets[w.Endpoint] = strings.ReplaceAll(w.Bucket, ":", "_")
		if buckets[w.Endpoint] == "" {
			buckets[w.Endpoint] = defaultBucket
		}
		if !firstRun && oldBuckets[w.Endpoint] != buckets[w.Endpoint] {
			logger.WithFields(logrus.Fields{"endpoint": w.Endpoint, "bucket": w.Weight, "oldBucket": oldBuckets[w.Endpoint]}).Infof("bucket changed")
		}
	}
	return nil
}

// updateRedisStatus checks if redis is healthy and updates redisIsHealthy accordingly.
func updateRedisStatus() error {
	oldStatus := redisIsHealthy.Load()
	newStatus := true
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*1))
	defer cancel()
	err := redisClient.Ping(ctx).Err()
	if err != nil {
		logger.WithError(err).Errorf("error pinging redis")
		newStatus = false
	}
	if oldStatus != newStatus {
		logger.WithFields(logrus.Fields{"oldStatus": oldStatus, "newStatus": newStatus}).Infof("redis status changed")
	}
	redisIsHealthy.Store(newStatus)
	return nil
}

// updateStats scans redis for ratelimit:stats:* keys and inserts them into postgres, if the key's truncated date is older than specified stats-truncation it will also delete the key in redis.
func updateStats(redisClient *redis.Client) error {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ratelimit_updateStats").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	var err error
	startTruncated := start.Truncate(statsTruncateDuration)

	allKeys := []string{}
	cursor := uint64(0)

	for {
		// rl:s:<year>-<month>-<day>-<hour>:<userId>:<apikey>:<route>
		cmd := redisClient.Scan(ctx, cursor, "rl:s:*:*:*:*", 1000)
		if cmd.Err() != nil {
			return cmd.Err()
		}
		keys, nextCursor, err := cmd.Result()
		if err != nil {
			return err
		}
		cursor = nextCursor
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}

	batchSize := 10000
	for i := 0; i <= len(allKeys); i += batchSize {
		start := i
		end := i + batchSize
		if end > len(allKeys) {
			end = len(allKeys)
		}

		if start == end {
			break
		}

		keysToDelete := []string{}
		keys := allKeys[start:end]
		entries := make([]DbEntry, len(keys))
		for i, k := range keys {
			ks := strings.Split(k, ":")
			if len(ks) != 6 {
				return fmt.Errorf("error parsing key %s: split-len != 5", k)
			}
			dateString := ks[2]
			date, err := time.Parse("2006-01-02-15", dateString)
			if err != nil {
				return fmt.Errorf("error parsing date in key %s: %v", k, err)
			}
			dateTruncated := date.Truncate(statsTruncateDuration)
			if dateTruncated.Before(startTruncated) {
				keysToDelete = append(keysToDelete, k)
			}
			userIdStr := ks[3]
			userId, err := strconv.ParseInt(userIdStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing userId in key %s: %v", k, err)
			}
			entries[i] = DbEntry{
				Date:     dateTruncated,
				UserId:   userId,
				ApiKey:   ks[4],
				Endpoint: ks[5],
			}
		}

		mgetSize := 500
		for j := 0; j < len(keys); j += mgetSize {
			mgetStart := j
			mgetEnd := j + mgetSize
			if mgetEnd > len(keys) {
				mgetEnd = len(keys)
			}
			mgetRes, err := redisClient.MGet(ctx, keys[mgetStart:mgetEnd]...).Result()
			if err != nil {
				return fmt.Errorf("error getting stats-count from redis (%v-%v/%v): %w", mgetStart, mgetEnd, len(keys), err)
			}
			for k, v := range mgetRes {
				vStr, ok := v.(string)
				if !ok {
					return fmt.Errorf("error parsing stats-count from redis: value is not string: %v: %v: %w", k, v, err)
				}
				entries[mgetStart+k].Count, err = strconv.ParseInt(vStr, 10, 64)
				if err != nil {
					return fmt.Errorf("error parsing stats-count from redis: value is not int64: %v: %v: %w", k, v, err)
				}
			}
		}

		err = updateStatsEntries(entries)
		if err != nil {
			return fmt.Errorf("error updating stats entries: %w", err)
		}

		if len(keysToDelete) > 0 {
			delSize := 500
			for j := 0; j < len(keysToDelete); j += delSize {
				delStart := j
				delEnd := j + delSize
				if delEnd > len(keysToDelete) {
					delEnd = len(keysToDelete)
				}
				_, err = redisClient.Del(ctx, keysToDelete[delStart:delEnd]...).Result()
				if err != nil {
					logger.Errorf("error deleting stats-keys from redis: %v", err)
				}
			}
		}
	}

	return nil
}

func updateStatsEntries(entries []DbEntry) error {
	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	numArgs := 4
	batchSize := 65535 / numArgs // max 65535 params per batch, since postgres uses int16 for binding input params
	valueArgs := make([]interface{}, 0, batchSize*numArgs)
	valueStrings := make([]string, 0, batchSize)
	valueStringArr := make([]string, numArgs)
	batchIdx, allIdx := 0, 0
	for _, entry := range entries {
		for u := 0; u < numArgs; u++ {
			valueStringArr[u] = fmt.Sprintf("$%d", batchIdx*numArgs+1+u)
		}

		valueStrings = append(valueStrings, "("+strings.Join(valueStringArr, ",")+")")
		valueArgs = append(valueArgs, entry.Date)
		valueArgs = append(valueArgs, entry.ApiKey)
		valueArgs = append(valueArgs, entry.Endpoint)
		valueArgs = append(valueArgs, entry.Count)

		// logger.WithFields(logger.Fields{"count": entry.Count, "apikey": entry.ApiKey, "path": entry.Path, "date": entry.Date}).Infof("inserting stats entry %v/%v", allIdx+1, len(entries))

		batchIdx++
		allIdx++

		if batchIdx >= batchSize || allIdx >= len(entries) {
			stmt := fmt.Sprintf(`INSERT INTO api_statistics (ts, apikey, endpoint, count) VALUES %s ON CONFLICT (ts, apikey, endpoint) DO UPDATE SET count = EXCLUDED.count`, strings.Join(valueStrings, ","))
			_, err := tx.Exec(stmt, valueArgs...)
			if err != nil {
				return err
			}
			batchIdx = 0
			valueArgs = valueArgs[:0]
			valueStrings = valueStrings[:0]
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// updateRateLimits updates the maps rateLimits, rateLimitsByUserId and userIdByApiKey with data from postgres-tables api_keys and api_ratelimits.
func updateRateLimits() error {
	start := time.Now()
	defer func() {
		logger.WithField("duration", time.Since(start)).Infof("updateRateLimits")
		metrics.TaskDuration.WithLabelValues("ratelimit_updateRateLimits").Observe(time.Since(start).Seconds())
	}()

	lastRateLimitUpdateMu.Lock()
	lastTKeys := lastRateLimitUpdateKeys
	lastTRateLimits := lastRateLimitUpdateRateLimits
	lastRateLimitUpdateMu.Unlock()

	tx, err := db.FrontendWriterDB.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	dbApiKeys := []struct {
		UserID     int64     `db:"user_id"`
		ApiKey     string    `db:"api_key"`
		ValidUntil time.Time `db:"valid_until"`
		ChangedAt  time.Time `db:"changed_at"`
	}{}

	err = tx.Select(&dbApiKeys, `SELECT user_id, api_key, valid_until, changed_at FROM api_keys WHERE changed_at > $1 OR valid_until < NOW()`, lastTKeys)
	if err != nil {
		return fmt.Errorf("error getting api_keys: %w", err)
	}

	dbRateLimits := []struct {
		UserID     int64     `db:"user_id"`
		Second     int64     `db:"second"`
		Hour       int64     `db:"hour"`
		Month      int64     `db:"month"`
		ValidUntil time.Time `db:"valid_until"`
		ChangedAt  time.Time `db:"changed_at"`
	}{}

	err = tx.Select(&dbRateLimits, `SELECT user_id, second, hour, month, valid_until, changed_at FROM api_ratelimits WHERE changed_at > $1 OR valid_until < NOW()`, lastTRateLimits)
	if err != nil {
		return fmt.Errorf("error getting api_ratelimits: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	dbApiProducts, err := DBGetCurrentApiProducts()
	if err != nil {
		return err
	}

	rateLimitsMu.Lock()
	now := time.Now()

	for _, dbApiProduct := range dbApiProducts {
		if dbApiProduct.Name == "nokey" {
			NoKeyRateLimit.Second = dbApiProduct.Second
			NoKeyRateLimit.Hour = dbApiProduct.Hour
			NoKeyRateLimit.Month = dbApiProduct.Month
		}
		if dbApiProduct.Name == "free" {
			FreeRatelimit.Second = dbApiProduct.Second
			FreeRatelimit.Hour = dbApiProduct.Hour
			FreeRatelimit.Month = dbApiProduct.Month
		}
	}

	for _, dbKey := range dbApiKeys {
		if dbKey.ChangedAt.After(lastTKeys) {
			lastTKeys = dbKey.ChangedAt
		}
		if dbKey.ValidUntil.Before(now) {
			delete(userIdByApiKey, dbKey.ApiKey)
			continue
		}
		userIdByApiKey[dbKey.ApiKey] = dbKey.UserID
	}

	for _, dbRl := range dbRateLimits {
		if dbRl.ChangedAt.After(lastTRateLimits) {
			lastTRateLimits = dbRl.ChangedAt
		}
		if dbRl.ValidUntil.Before(now) {
			delete(rateLimitsByUserId, dbRl.UserID)
			continue
		}
		rlStr := fmt.Sprintf("%d/%d/%d", dbRl.Second, dbRl.Hour, dbRl.Month)
		rl, exists := rateLimits[rlStr]
		if !exists {
			rl = &RateLimit{
				Second: dbRl.Second,
				Hour:   dbRl.Hour,
				Month:  dbRl.Month,
			}
			rateLimits[rlStr] = rl
		}
		rateLimitsByUserId[dbRl.UserID] = rl
	}
	rateLimitsMu.Unlock()
	metrics.TaskDuration.WithLabelValues("ratelimit_updateRateLimits_lock").Observe(time.Since(now).Seconds())

	lastRateLimitUpdateMu.Lock()
	lastRateLimitUpdateKeys = lastTKeys
	lastRateLimitUpdateRateLimits = lastTRateLimits
	lastRateLimitUpdateMu.Unlock()

	return nil
}

// postRateLimit decrements the rate limit keys in redis if the status is not 200.
func postRateLimit(rl *RateLimitResult, status int) error {
	// if status == http.StatusOK {
	if !(status >= 500 && status <= 599) {
		// anything other than 5xx is considered successful and counts towards the rate limit
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pipe := redisClient.Pipeline()

	decrByWeight := rl.Weight
	mbrw := GetMaxBadRquestWeight()
	if decrByWeight > mbrw {
		decrByWeight = mbrw
	}

	for _, k := range rl.RedisKeys {
		pipe.DecrBy(ctx, k.Key, decrByWeight)
		pipe.ExpireAt(ctx, k.Key, k.ExpireAt) // make sure all keys have a TTL
	}
	pipe.DecrBy(ctx, rl.RedisStatsKey, 1)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// rateLimitRequest is the main function for rate limiting, it will check the rate limits for the request and update the rate limits in redis.
func rateLimitRequest(r *http.Request) (*RateLimitResult, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ratelimit_rateLimitRequest").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*1000)
	defer cancel()

	res := &RateLimitResult{}
	// defer func() { logger.Infof("rateLimitRequest: %+v", *res) }()

	key, ip := getKey(r)
	res.Key = key
	res.IP = ip

	rateLimitsMu.RLock()
	userId, ok := userIdByApiKey[key]
	if !ok {
		res.UserId = -1
		res.IsValidKey = false
		res.RateLimit = NoKeyRateLimit
	} else {
		res.UserId = userId
		res.IsValidKey = true
		limit, ok := rateLimitsByUserId[userId]
		if ok {
			res.RateLimit = limit
		} else {
			res.RateLimit = FreeRatelimit
		}
	}
	rateLimitsMu.RUnlock()

	weight, route, bucket := getWeight(r)
	res.Weight = weight
	res.Route = route
	res.Bucket = bucket

	startUtc := start.UTC()
	res.Time = startUtc

	nextHourUtc := time.Now().Truncate(time.Hour).Add(time.Hour)
	nextMonthUtc := time.Date(startUtc.Year(), startUtc.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	timeUntilNextHourUtc := nextHourUtc.Sub(startUtc)
	timeUntilNextMonthUtc := nextMonthUtc.Sub(startUtc)

	rateLimitSecondKey := fmt.Sprintf("rl:c:s:%s:%d", res.Bucket, res.UserId)
	rateLimitHourKey := fmt.Sprintf("rl:c:h:%04d-%02d-%02d-%02d:%s:%d", startUtc.Year(), startUtc.Month(), startUtc.Day(), startUtc.Hour(), res.Bucket, res.UserId)
	rateLimitMonthKey := fmt.Sprintf("rl:c:m:%04d-%02d:%s:%d", startUtc.Year(), startUtc.Month(), res.Bucket, res.UserId)
	statsKey := fmt.Sprintf("rl:s:%04d-%02d-%02d-%02d:%d:%s:%s", startUtc.Year(), startUtc.Month(), startUtc.Day(), startUtc.Hour(), res.UserId, res.Key, res.Route)
	if !res.IsValidKey {
		rateLimitSecondKey = fmt.Sprintf("rl:c:s:%s:%s", res.Bucket, res.IP)
		rateLimitHourKey = fmt.Sprintf("rl:c:h:%04d-%02d-%02d-%02d:%s:%s", startUtc.Year(), startUtc.Month(), startUtc.Day(), startUtc.Hour(), res.Bucket, res.IP)
		rateLimitMonthKey = fmt.Sprintf("rl:c:m:%04d-%02d:%s:%s", startUtc.Year(), startUtc.Month(), res.Bucket, res.IP)
		statsKey = fmt.Sprintf("rl:s:%04d-%02d-%02d-%02d:%d:%s:%s", startUtc.Year(), startUtc.Month(), startUtc.Day(), startUtc.Hour(), res.UserId, "nokey", res.Route)
	}
	res.RedisStatsKey = statsKey

	pipe := redisClient.Pipeline()

	var rateLimitSecond, rateLimitHour, rateLimitMonth *redis.IntCmd

	if res.RateLimit.Second > 0 {
		rateLimitSecond = pipe.IncrBy(ctx, rateLimitSecondKey, weight)
		pipe.ExpireNX(ctx, rateLimitSecondKey, time.Second)
	}

	if res.RateLimit.Hour > 0 {
		rateLimitHour = pipe.IncrBy(ctx, rateLimitHourKey, weight)
		pipe.ExpireAt(ctx, rateLimitHourKey, nextHourUtc.Add(time.Second*60)) // expire 1 minute after the window to make sure we do not miss any requests due to time-sync
		res.RedisKeys = append(res.RedisKeys, RedisKey{rateLimitHourKey, nextHourUtc.Add(time.Second * 60)})
	}

	if res.RateLimit.Month > 0 {
		rateLimitMonth = pipe.IncrBy(ctx, rateLimitMonthKey, weight)
		pipe.ExpireAt(ctx, rateLimitMonthKey, nextMonthUtc.Add(time.Second*60)) // expire 1 minute after the window to make sure we do not miss any requests due to time-sync
		res.RedisKeys = append(res.RedisKeys, RedisKey{rateLimitMonthKey, nextMonthUtc.Add(time.Second * 60)})
	}

	pipe.Incr(ctx, statsKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if res.RateLimit.Month > 0 && rateLimitMonth.Val() > res.RateLimit.Month {
		res.Limit = res.RateLimit.Month
		res.Remaining = 0
		res.Reset = int64(timeUntilNextMonthUtc.Seconds())
		res.Window = MonthTimeWindow
		res.BlockRequest = true
	} else if res.RateLimit.Hour > 0 && rateLimitHour.Val() > res.RateLimit.Hour {
		res.Limit = res.RateLimit.Hour
		res.Remaining = 0
		res.Reset = int64(timeUntilNextHourUtc.Seconds())
		res.Window = HourTimeWindow
		res.BlockRequest = true
	} else if res.RateLimit.Second > 0 && rateLimitSecond.Val() > res.RateLimit.Second {
		res.Limit = res.RateLimit.Second
		res.Remaining = 0
		res.Reset = int64(1)
		res.Window = SecondTimeWindow
		res.BlockRequest = true
	} else {
		res.Limit = res.RateLimit.Second
		res.Remaining = res.RateLimit.Second - rateLimitSecond.Val()
		res.Reset = int64(1)
		res.Window = SecondTimeWindow
	}

	if res.RateLimit.Second > 0 {
		res.RemainingSecond = res.RateLimit.Second - rateLimitSecond.Val()
		if res.RemainingSecond < 0 {
			res.RemainingSecond = 0
		}
	}
	if res.RateLimit.Hour > 0 {
		res.RemainingHour = res.RateLimit.Hour - rateLimitHour.Val()
		if res.RemainingHour < 0 {
			res.RemainingHour = 0
		}
	}
	if res.RateLimit.Month > 0 {
		res.RemainingMonth = res.RateLimit.Month - rateLimitMonth.Val()
		if res.RemainingMonth < 0 {
			res.RemainingMonth = 0
		}
	}

	// normalize limit-headers to keep them consistent with previous versions
	if res.RateLimit.Month > 0 {
		res.LimitMonth = res.RateLimit.Month
	} else {
		res.LimitMonth = max(res.RateLimit.Month, res.RateLimit.Hour, res.RateLimit.Second)
		res.RemainingMonth = max(res.RemainingMonth, res.RemainingHour, res.RemainingSecond)
	}
	res.LimitDay = res.LimitMonth
	res.RemainingDay = res.RemainingMonth

	if res.RateLimit.Hour > 0 {
		res.LimitHour = res.RateLimit.Hour
	} else {
		res.LimitHour = res.LimitMonth
		res.RemainingHour = res.RemainingMonth
	}
	res.LimitMinute = res.LimitHour
	res.RemainingMinute = res.RemainingHour

	if res.RateLimit.Second > 0 {
		res.LimitSecond = res.RateLimit.Second
	} else {
		res.LimitSecond = res.LimitHour
	}

	return res, nil
}

func max(vals ...int64) int64 {
	max := vals[0]
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	return max
}

// getKey returns the key used for RateLimiting. It first checks the query params, then the header and finally the ip address.
func getKey(r *http.Request) (key, ip string) {
	ip = getIP(r)
	key = r.URL.Query().Get("apikey")
	if key != "" {
		return key, ip
	}
	key = r.Header.Get("apikey")
	if key != "" {
		return key, ip
	}
	key = r.Header.Get("X-API-KEY")
	if key != "" {
		return key, ip
	}
	return "ip_" + strings.ReplaceAll(ip, ":", "_"), ip
}

// getWeight returns the weight of an endpoint. if the weight of the endpoint is not defined, it returns 1.
func getWeight(r *http.Request) (cost int64, identifier, bucket string) {
	route := getRoute(r)
	weightsMu.RLock()
	weight, weightOk := weights[route]
	bucket, bucketOk := buckets[route]
	weightsMu.RUnlock()
	if !weightOk {
		weight = 1
	}
	if !bucketOk {
		bucket = defaultBucket
	}
	return weight, route, bucket
}

func getRoute(r *http.Request) string {
	route := mux.CurrentRoute(r)
	pathTpl, err := route.GetPathTemplate()
	if err != nil {
		return "UNDEFINED"
	}
	return pathTpl
}

// getIP returns the ip address from the http request
func getIP(r *http.Request) string {
	ips := r.Header.Get("CF-Connecting-IP")
	if ips == "" {
		ips = r.Header.Get("X-Forwarded-For")
	}
	splitIps := strings.Split(ips, ",")

	if len(splitIps) > 0 {
		// get last IP in list since ELB prepends other user defined IPs, meaning the last one is the actual client IP.
		netIP := net.ParseIP(splitIps[len(splitIps)-1])
		if netIP != nil {
			return netIP.String()
		}
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "INVALID"
	}

	netIP := net.ParseIP(ip)
	if netIP != nil {
		ip := netIP.String()
		if ip == "::1" {
			return "127.0.0.1"
		}
		return ip
	}

	return "INVALID"
}

type FallbackRateLimiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type FallbackRateLimiter struct {
	clients map[string]*FallbackRateLimiterClient
	mu      sync.Mutex
}

func NewFallbackRateLimiter() *FallbackRateLimiter {
	rl := &FallbackRateLimiter{
		clients: make(map[string]*FallbackRateLimiterClient),
	}
	go func() {
		for {
			time.Sleep(time.Minute)
			rl.mu.Lock()
			for ip, client := range rl.clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *FallbackRateLimiter) Handle(w http.ResponseWriter, r *http.Request, next func(writer http.ResponseWriter, request *http.Request)) {
	key, _ := getKey(r)
	rl.mu.Lock()
	if _, found := rl.clients[key]; !found {
		rl.clients[key] = &FallbackRateLimiterClient{limiter: rate.NewLimiter(FallbackRateLimitSecond, FallbackRateLimitBurst)}
	}
	rl.clients[key].lastSeen = time.Now()
	if !rl.clients[key].limiter.Allow() {
		rl.mu.Unlock()
		w.Header().Set(HeaderRateLimitLimit, strconv.FormatInt(FallbackRateLimitSecond, 10))
		w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(1, 10))
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}
	rl.mu.Unlock()
	next(w, r)
}

func DBGetUserApiRateLimit(userId int64) (*RateLimit, error) {
	rl := &RateLimit{}
	err := db.FrontendWriterDB.Get(rl, `
        select second, hour, month
        from api_ratelimits
        where user_id = $1`, userId)
	if err != nil && err == sql.ErrNoRows {
		rl.Second = FreeRatelimit.Second
		rl.Hour = FreeRatelimit.Hour
		rl.Month = FreeRatelimit.Month
		return rl, nil
	}
	return rl, err
}

func DBGetCurrentApiProducts() ([]*ApiProduct, error) {
	apiProducts := []*ApiProduct{}
	err := db.FrontendWriterDB.Select(&apiProducts, `
        select distinct on (name) name, stripe_price_id, second, hour, month, valid_from 
        from api_products 
        where valid_from <= now()
        order by name, valid_from desc`)
	return apiProducts, err
}

func DBUpdater() {
	iv := utils.Config.Frontend.RatelimitUpdateInterval
	if iv < time.Second {
		logger.Warnf("updateInterval is below 1s, setting to 60s")
		iv = time.Second * 60
	}
	logger.WithField("redis", utils.Config.RedisSessionStoreEndpoint).Infof("starting db updater")
	redisClient = redis.NewClient(&redis.Options{
		Addr:        utils.Config.RedisSessionStoreEndpoint,
		ReadTimeout: time.Second * 3,
	})
	for {
		DBUpdate(redisClient)
		time.Sleep(iv)
	}
}

func DBUpdate(redisClient *redis.Client) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		start := time.Now()
		err := updateStats(redisClient)
		if err != nil {
			logger.WithError(err).Errorf("error updating stats")
			return
		}
		logger.WithField("duration", time.Since(start)).Infof("updated stats")
	}()
	go func() {
		defer wg.Done()
		start := time.Now()
		res, err := DBUpdateApiKeys()
		if err != nil {
			logger.WithError(err).Errorf("error updating api_keys")
			return
		}
		ra, err := res.RowsAffected()
		if err != nil {
			logger.WithError(err).Errorf("error getting rows affected")
			return
		}
		logger.WithField("duration", time.Since(start)).WithField("updates", ra).Infof("updated api_keys")

		start = time.Now()
		res, err = DBUpdateApiRatelimits()
		if err != nil {
			logger.WithError(err).Errorf("error updating api_ratelimit")
			return
		}
		ra, err = res.RowsAffected()
		if err != nil {
			logger.WithError(err).Errorf("error getting rows affected")
			return
		}
		logger.WithField("duration", time.Since(start)).WithField("updates", ra).Infof("updated api_ratelimits")

		start = time.Now()
		res, err = DBInvalidateApiKeys()
		if err != nil {
			logger.WithError(err).Errorf("error invalidating api_keys")
			return
		}
		ra, err = res.RowsAffected()
		if err != nil {
			logger.WithError(err).Errorf("error getting rows affected")
			return
		}
		logger.WithField("duration", time.Since(start)).WithField("updates", ra).Infof("invalidated api_keys")
	}()
	wg.Wait()
}

// DBInvalidateApiKeys invalidates api_keys that are not associated with a user. This func is only needed until api-key-mgmt is fully implemented - where users.apikey column is not used anymore.
func DBInvalidateApiKeys() (sql.Result, error) {
	return db.FrontendWriterDB.Exec(`
        update api_keys 
        set changed_at = now(), valid_until = now() 
        where valid_until > now() and not exists (select id from users where id = api_keys.user_id)`)
}

// DBUpdateApiKeys updates the api_keys table with the api_keys from the users table. This func is only needed until api-key-mgmt is fully implemented - where users.apikey column is not used anymore.
func DBUpdateApiKeys() (sql.Result, error) {
	return db.FrontendWriterDB.Exec(
		`insert into api_keys (user_id, api_key, valid_until, changed_at)
        select 
            id as user_id, 
            api_key,
            to_timestamp('9999-12-31 23:59:59', 'YYYY-MM-DD HH24:MI:SS') as valid_until,
            now() as changed_at
        from users 
        where api_key is not null and not exists (select user_id from api_keys where api_keys.user_id = users.id)
        on conflict (api_key) do update set
			user_id = excluded.user_id,
            valid_until = excluded.valid_until,
            changed_at = excluded.changed_at
        where api_keys.valid_until != excluded.valid_until`,
	)
}

func DBUpdateApiRatelimits() (sql.Result, error) {
	return db.FrontendWriterDB.Exec(
		`with 
			current_api_products as (
				select distinct on (name) name, stripe_price_id, second, hour, month, valid_from 
				from api_products 
				where valid_from <= now()
				order by name, valid_from desc
			)
		insert into api_ratelimits (user_id, second, hour, month, valid_until, changed_at)
		select 
			user_id,
			case when min(second) = 0 then 0 else max(second) end as second,
			case when min(hour) = 0 then 0 else max(hour) end as hour,
			case when min(month) = 0 then 0 else max(month) end as month,
			to_timestamp('9999-12-31 23:59:59', 'YYYY-MM-DD HH24:MI:SS') as valid_until,
			now() as changed_at
		from (
			-- set all current ratelimits to free
			select user_id, cap.second, cap.hour, cap.month
			from api_ratelimits
			left join current_api_products cap on cap.name = 'free'
		union
			-- set ratelimits for stripe subscriptions
			select u.id as user_id, cap.second, cap.hour, cap.month
			from users_stripe_subscriptions uss
			left join users u on u.stripe_customer_id = uss.customer_id
			left join current_api_products cap on cap.stripe_price_id = uss.price_id
			where uss.active = true and u.id is not null
		union
			-- set ratelimits for app subscriptions
			select asv.user_id, cap.second, cap.hour, cap.month
			from app_subs_view asv
			left join current_api_products cap on cap.name = asv.product_id
			where asv.active = true
		union
			-- set ratelimits for admins to unlimited
			select u.id as user_id, cap.second, cap.hour, cap.month
			from users u
			left join current_api_products cap on cap.name = 'unlimited'
			where u.user_group = 'ADMIN' and cap.second is not null
		) a
		group by user_id
		on conflict (user_id) do update set
			second = excluded.second,
			hour = excluded.hour,
			month = excluded.month,
			valid_until = excluded.valid_until,
			changed_at = now()
		where
			api_ratelimits.second != excluded.second 
			or api_ratelimits.hour != excluded.hour 
			or api_ratelimits.month != excluded.month`)
}

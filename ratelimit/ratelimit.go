package RateLimit

import (
	"context"
	"eth2-exporter/db"
	"eth2-exporter/metrics"
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
	HeaderRateLimitLimit     = "X-RateLimit-Limit"     // the rate limit ceiling that is applicable for the current request
	HeaderRateLimitRemaining = "X-RateLimit-Remaining" // the number of requests left for the current rate-limit window
	HeaderRateLimitReset     = "X-RateLimit-Reset"     // the number of seconds until the quota resets
	HeaderRetryAfter         = "Retry-After"           // the number of seconds until the quota resets, same as HeaderRateLimitReset, RFC 7231, 7.1.3

	NokeyRateLimitSecond = 5   // RateLimit for requests without or with invalid apikey
	NokeyRateLimitHour   = 500 // RateLimit for requests without or with invalid apikey
	NokeyRateLimitMonth  = 0   // RateLimit for requests without or with invalid apikey

	FallbackRateLimitSecond = 20 // RateLimit for when redis is offline
	FallbackRateLimitBurst  = 20 // RateLimit for when redis is offline

	SecondTimeWindow = "second"
	HourTimeWindow   = "hour"
	MonthTimeWindow  = "month"
)

var NoKeyRateLimit = &RateLimit{
	Second: NokeyRateLimitSecond,
	Hour:   NokeyRateLimitHour,
	Month:  NokeyRateLimitMonth,
}

var redisClient *redis.Client
var redisIsHealthy atomic.Bool

var fallbackRateLimiter = NewFallbackRateLimiter() // if redis is offline, use this rate limiter

var initializedWg = &sync.WaitGroup{} // wait for everything to be initialized before serving requests

var rateLimitsMu = &sync.RWMutex{}
var rateLimits = make(map[string]*RateLimit)      // guarded by RateLimitsMu
var rateLimitsByKey = make(map[string]*RateLimit) // guarded by RateLimitsMu

var weightsMu = &sync.RWMutex{}
var weights = map[string]int64{} // guarded by weightsMu

var pathPrefix = "" // only requests with this prefix will be RateLimited

var logger = logrus.StandardLogger().WithField("module", "ratelimit")

type dbEntry struct {
	Date  time.Time
	Key   string
	Path  string
	Count int64
}

type RateLimit struct {
	Second int64
	Hour   int64
	Month  int64
}

type RateLimitResult struct {
	Time          time.Time
	Weight        int64
	Route         string
	IP            string
	Key           string
	IsValidKey    bool
	RedisKeys     []RedisKey
	RedisStatsKey string
	RateLimit     *RateLimit
	Limit         int64
	Remaining     int64
	Reset         int64
	Window        TimeWindow
}

type RedisKey struct {
	Key      string
	ExpireAt time.Time
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
}

func (r *responseWriterDelegator) Status() int {
	return r.status
}

// Init initializes the RateLimiting middleware, the RateLimiting middleware will not work without calling Init first.
func Init(redisAddress, pathPrefixOpt string) {
	pathPrefix = pathPrefixOpt

	redisClient = redis.NewClient(&redis.Options{
		Addr:        redisAddress,
		ReadTimeout: time.Second * 3,
	})

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
			time.Sleep(time.Second * 60)
		}
	}()
	go func() {
		firstRun := true
		lastRunTime := time.Unix(0, 0)
		for {
			t, err := updateRateLimits(lastRunTime)
			if err != nil {
				logger.WithError(err).Errorf("error updating RateLimits")
				time.Sleep(time.Second * 2)
				continue
			}
			lastRunTime = t
			if firstRun {
				initializedWg.Done()
				firstRun = false
			}
			time.Sleep(time.Second * 60)
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
	go func() {
		for {
			err := updateStats()
			if err != nil {
				logger.WithError(err).Errorf("error updating stats")
			}
			time.Sleep(time.Second * 60)
		}
	}()

	initializedWg.Wait()
}

// HttpMiddleware returns an http.Handler that can be used as middleware to RateLimit requests. If redis is offline, it will use a fallback rate limiter.
func HttpMiddleware(next http.Handler) http.Handler {
	initializedWg.Wait()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, pathPrefix) {
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
		// logrus.WithFields(logrus.Fields{"route": rl.Route, "key": rl.Key, "limit": rl.Limit, "remaining": rl.Remaining, "reset": rl.Reset, "window": rl.Window}).Infof("RateLimiting")

		w.Header().Set(HeaderRateLimitLimit, strconv.FormatInt(rl.Limit, 10))
		w.Header().Set(HeaderRateLimitRemaining, strconv.FormatInt(rl.Remaining, 10))
		w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(rl.Reset, 10))
		if rl.Weight > rl.Remaining {
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

// updateWeights gets the weights from postgres and updates the weights map.
func updateWeights(firstRun bool) error {
	dbWeights := []struct {
		Endpoint  string    `db:"endpoint"`
		Weight    int64     `db:"weight"`
		ValidFrom time.Time `db:"valid_from"`
	}{}
	err := db.WriterDb.Select(&dbWeights, "SELECT DISTINCT ON (endpoint) endpoint, weight, valid_from FROM api_weights WHERE valid_from <= NOW() ORDER BY endpoint, valid_from DESC")
	if err != nil {
		return err
	}
	weightsMu.Lock()
	defer weightsMu.Unlock()
	oldWeights := weights
	weights = make(map[string]int64, len(dbWeights))
	for _, w := range dbWeights {
		weights[w.Endpoint] = w.Weight
		if !firstRun && oldWeights[w.Endpoint] != w.Weight {
			logger.WithFields(logrus.Fields{"endpoint": w.Endpoint, "weight": w.Weight, "oldWeight": oldWeights[w.Endpoint]}).Infof("weight changed")
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

// updateStats scans redis for ratelimit:stats:* keys and inserts them into postgres, if the key's date is in the past it will also delete the key in redis.
func updateStats() error {
	allKeys := []string{}
	cursor := uint64(0)
	ctx := context.Background()
	for {
		cmd := redisClient.Scan(ctx, cursor, "ratelimit:stats:*:*:*", 1000)
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
		keys := allKeys[start:end]
		entries := make([]dbEntry, len(keys))
		values := make([]*redis.StringCmd, len(keys))
		cmds, err := redisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for i, k := range keys {
				ks := strings.Split(k, ":")
				if len(ks) != 5 {
					return fmt.Errorf("error parsing key %s: split-len != 5", k)
				}
				dateString := ks[2]
				date, err := time.Parse("2006-01-02", dateString)
				if err != nil {
					return fmt.Errorf("error parsing date in key %s: %v", k, err)
				}
				key := ks[3]
				path := ks[4]
				values[i] = pipe.Get(ctx, k)
				entries[i] = dbEntry{
					Date: date,
					Key:  key,
					Path: path,
				}
			}
			return nil
		})
		for i := range cmds {
			entries[i].Count, err = values[i].Int64()
			if err != nil {
				return fmt.Errorf("error parsing count of key %s: %v: %w", entries[i].Key, entries[i].Count, err)
			}
		}
		if err != nil {
			return err
		}
		err = updateStatsEntries(entries)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateStatsEntries(entries []dbEntry) error {
	tx, err := db.WriterDb.Beginx()
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
		valueArgs = append(valueArgs, entry.Key)
		valueArgs = append(valueArgs, entry.Path)
		valueArgs = append(valueArgs, entry.Count)

		logger.WithFields(logrus.Fields{"count": entry.Count, "key": entry.Key}).Infof("inserting stats entry %v/%v", allIdx, len(entries))

		batchIdx++
		allIdx++

		if batchIdx >= batchSize || allIdx >= len(entries) {
			stmt := fmt.Sprintf(`INSERT INTO api_statistics (ts, apikey, call, count) VALUES %s ON CONFLICT (ts, apikey, call) DO UPDATE SET count = excluded.count`, strings.Join(valueStrings, ","))
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

// updateRateLimits gets the ratelimits from postgres and updates the ratelimits map. it will delete expired ratelimits and assumes that no other process deletes entries in the table api_ratelimits.
func updateRateLimits(lastUpdate time.Time) (time.Time, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ratelimit_updateRateLimits").Observe(time.Since(start).Seconds())
	}()
	dbRateLimits := []struct {
		UserID     int64     `db:"user_id"`
		ApiKey     string    `db:"apikey"`
		Second     int64     `db:"second"`
		Hour       int64     `db:"hour"`
		Month      int64     `db:"month"`
		ValidUntil time.Time `db:"valid_until"`
		ChangedAt  time.Time `db:"changed_at"`
	}{}
	err := db.WriterDb.Select(&dbRateLimits, "SELECT ar.user_id, ak.apikey, ar.second, ar.hour, ar.month, ar.valid_until, ar.changed_at FROM api_ratelimits ar LEFT JOIN users u ON u.id = ar.user_id LEFT JOIN api_keys ak ON ak.user_id = u.id WHERE ar.changed_at > $1", lastUpdate)
	if err != nil {
		return lastUpdate, fmt.Errorf("error getting ratelimits: %w", err)
	}

	rateLimitsMu.Lock()
	now := time.Now()
	newestChange := time.Unix(0, 0)
	for _, dbRl := range dbRateLimits {
		if dbRl.ChangedAt.After(newestChange) {
			newestChange = dbRl.ChangedAt
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
		_, exists = rateLimitsByKey[dbRl.ApiKey]
		if !exists {
			rateLimitsByKey[dbRl.ApiKey] = rl
		}
		if dbRl.ValidUntil.Before(now) {
			delete(rateLimitsByKey, dbRl.ApiKey)
		}
	}
	rateLimitsMu.Unlock()
	metrics.TaskDuration.WithLabelValues("ratelimit_updateRateLimits_lock").Observe(time.Since(now).Seconds())

	return newestChange, nil
}

func postRateLimit(rl *RateLimitResult, status int) error {
	if status == 200 {
		return nil
	}
	// logger.WithFields(logrus.Fields{"key": rl.Key, "status": status}).Infof("decreasing key")
	// if status is not 200 decrement keys since we do not count unsuccessful requests
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pipe := redisClient.Pipeline()
	for _, k := range rl.RedisKeys {
		pipe.DecrBy(ctx, k.Key, rl.Weight)
		pipe.ExpireAt(ctx, k.Key, k.ExpireAt) // make sure all keys have a TTL
	}
	pipe.DecrBy(ctx, rl.RedisStatsKey, 1)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func rateLimitRequest(r *http.Request) (*RateLimitResult, error) {
	start := time.Now()
	defer func() {
		metrics.TaskDuration.WithLabelValues("ratelimit_total").Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*1000)
	defer cancel()

	res := &RateLimitResult{}

	key, ip := getKey(r)
	res.Key = key
	res.IP = ip

	rateLimitsMu.RLock()
	limit, ok := rateLimits[key]
	rateLimitsMu.RUnlock()
	if !ok {
		res.IsValidKey = false
		res.RateLimit = &RateLimit{
			Second: NokeyRateLimitSecond,
			Hour:   NokeyRateLimitHour,
			Month:  NokeyRateLimitMonth,
		}
	} else {
		res.IsValidKey = true
		res.RateLimit = limit
	}

	weight, path := getWeight(r)
	res.Weight = weight
	res.Route = path

	startUtc := start.UTC()
	res.Time = startUtc
	t := startUtc.AddDate(0, 1, -startUtc.Day())
	endOfMonthUtc := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	timeUntilEndOfMonthUtc := endOfMonthUtc.Sub(startUtc)
	endOfHourUtc := time.Now().Truncate(time.Hour).Add(time.Hour)
	timeUntilEndOfHourUtc := endOfHourUtc.Sub(startUtc)

	RateLimitSecondKey := "ratelimit:second:" + res.Key
	RateLimitMonthKey := fmt.Sprintf("ratelimit:month:%04d-%02d:%s", startUtc.Year(), startUtc.Month(), res.Key)
	RateLimitHourKey := fmt.Sprintf("ratelimit:hour:%04d-%02d-%02d:%s", startUtc.Year(), startUtc.Month(), startUtc.Hour(), res.Key)

	statsKey := fmt.Sprintf("ratelimit:stats:%04d-%02d-%02d:%s:%s", startUtc.Year(), startUtc.Month(), startUtc.Day(), res.Key, path)
	if !res.IsValidKey {
		statsKey = fmt.Sprintf("ratelimit:stats:%04d-%02d-%02d:%s:%s", startUtc.Year(), startUtc.Month(), startUtc.Day(), "nokey", path)
	}
	res.RedisStatsKey = statsKey

	pipe := redisClient.Pipeline()

	var RateLimitSecond, RateLimitHour, RateLimitMonth *redis.IntCmd

	if res.RateLimit.Second > 0 {
		RateLimitSecond = pipe.IncrBy(ctx, RateLimitSecondKey, weight)
		pipe.ExpireNX(ctx, RateLimitSecondKey, time.Second)
	}

	if res.RateLimit.Hour > 0 {
		RateLimitHour = pipe.IncrBy(ctx, RateLimitHourKey, weight)
		pipe.ExpireAt(ctx, RateLimitHourKey, endOfHourUtc)
		res.RedisKeys = append(res.RedisKeys, RedisKey{RateLimitHourKey, endOfHourUtc})
	}

	if res.RateLimit.Month > 0 {
		RateLimitMonth = pipe.IncrBy(ctx, RateLimitMonthKey, weight)
		pipe.ExpireAt(ctx, RateLimitMonthKey, endOfMonthUtc)
		res.RedisKeys = append(res.RedisKeys, RedisKey{RateLimitMonthKey, endOfMonthUtc})
	}

	pipe.Incr(ctx, statsKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	if res.RateLimit.Second > 0 {
		if RateLimitSecond.Val() > res.RateLimit.Second {
			res.Limit = res.RateLimit.Second
			res.Remaining = 0
			res.Reset = int64(1)
			res.Window = SecondTimeWindow
			return res, nil
		} else if res.RateLimit.Second-RateLimitSecond.Val() > res.Limit {
			res.Limit = res.RateLimit.Second - RateLimitSecond.Val()
			res.Remaining = res.RateLimit.Second - RateLimitSecond.Val()
			res.Reset = int64(1)
			res.Window = SecondTimeWindow
		}
	}

	if res.RateLimit.Hour > 0 {
		if RateLimitSecond.Val() > res.RateLimit.Hour {
			res.Limit = res.RateLimit.Hour
			res.Remaining = 0
			res.Reset = int64(timeUntilEndOfHourUtc.Seconds())
			res.Window = HourTimeWindow
			return res, nil
		} else if res.RateLimit.Hour-RateLimitHour.Val() > res.Limit {
			res.Limit = res.RateLimit.Hour - RateLimitHour.Val()
			res.Remaining = res.RateLimit.Hour - RateLimitHour.Val()
			res.Reset = int64(timeUntilEndOfHourUtc.Seconds())
			res.Window = HourTimeWindow
		}
	}

	if res.RateLimit.Month > 0 {
		if RateLimitSecond.Val() > res.RateLimit.Month {
			res.Limit = res.RateLimit.Month
			res.Remaining = 0
			res.Reset = int64(timeUntilEndOfMonthUtc.Seconds())
			res.Window = MonthTimeWindow
			return res, nil
		} else if res.RateLimit.Month-RateLimitMonth.Val() > res.Limit {
			res.Limit = res.RateLimit.Month - RateLimitMonth.Val()
			res.Remaining = res.RateLimit.Month - RateLimitMonth.Val()
			res.Reset = int64(timeUntilEndOfMonthUtc.Seconds())
			res.Window = MonthTimeWindow
		}
	}

	return res, nil
}

// getKey returns the key used for RateLimiting. It first checks the query params, then the header and finally the ip address.
func getKey(r *http.Request) (key, ip string) {
	ip = getIP(r)
	key = r.URL.Query().Get("apikey")
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
func getWeight(r *http.Request) (cost int64, identifier string) {
	route := getRoute(r)
	weightsMu.RLock()
	weight, ok := weights[route]
	weightsMu.RUnlock()
	if ok {
		return weight, route
	}
	return 1, route
}

func getRoute(r *http.Request) string {
	route := mux.CurrentRoute(r)
	path, err := route.GetPathTemplate()
	if err != nil {
		path = "UNDEFINED"
	}
	return path
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

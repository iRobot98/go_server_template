package server

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kataras/iris/v12"
	"template.com/service/src/base"
	"template.com/service/src/logger"
)

var RateLimiter RateLimiterStruct

const RateLimitPeriod = 200 * time.Millisecond

var BannedIps base.MutexedMap[time.Time]

func Ban(ip string, reasons ...string) {
	funcname := base.NewStrF("Ban(%s)", ip)
	now := time.Now()
	logger.LoggerEngine.LogErrF(funcname.String(), "banned at %s\nreason: %s", base.FormatTime(&now), strings.Join(reasons, "\n"))
	// BannedIps.Set(ip, now)
}

type RateLimiterStruct struct {
	IPmap base.MutexedMap[*ConnectionAttemptStruct]
}

type ConnectionAttemptStruct struct {
	sync.RWMutex
	// limit connections using channel
	// channel size of 3
	Connect chan time.Time
	// check if connection is too soon
	// check mime type?
	LastWarningTime time.Time

	// URI visited at what time
	UrlMap base.MutexedMap[time.Time]

	// ticker
	RateLimitLocks base.CounterStruct
	Ticker         time.Ticker
	// counter
	Counter base.CounterStruct
}

func init() {
	RateLimiter = RateLimiterStruct{
		IPmap: base.MutexedMap[*ConnectionAttemptStruct]{
			Map: make(map[string]*ConnectionAttemptStruct),
		},
	}
	BannedIps = base.MutexedMap[time.Time]{
		Map: make(map[string]time.Time),
	}
}

func NewConnectStruct() *ConnectionAttemptStruct {
	CAS := &ConnectionAttemptStruct{
		Connect: make(chan time.Time, 4),
		Ticker:  *time.NewTicker(RateLimitPeriod),
		UrlMap: base.MutexedMap[time.Time]{
			Map: make(map[string]time.Time),
		},
	}
	now := time.Now()
	// 1
	CAS.Connect <- now
	// 2
	CAS.Connect <- now
	// 3
	CAS.Connect <- now
	return CAS
}

func BannedIpHandler(ctx iris.Context) {
	funcname := "BannedIpHandler(iris.Context)"
	ips := strings.Split(ctx.Request().RemoteAddr, ":")
	ip := ips[0]
	_, ok := BannedIps.Get(ip)
	handle := func() {
		ctx.StopWithStatus(iris.StatusNoContent)
	}
	if ok {
		handle()
		return
	}

	str := base.NewStr(ip)
	if !str.StartsWith("192.") && !str.StartsWith("127.") { // local net or localhost
		Ban(ip, "invalid ip source")
		handle()
		return
	}
	url := ctx.Request().URL.Path
	validurl := "^\\/[a-z\\/0-9?=+%]+(|.html|.css|.js|.ico)"

	str.Swap(url)
	err := str.LowerCase().Validate(validurl)
	if err != nil {
		logger.LoggerEngine.LogErrF(funcname, "regex test url: %s\n%v", url, err)
	}

	// max url length is 256 characters
	// (revise later)
	if len(url) > int(512) {
		Ban(ip, "url too long", url)
		ctx.StopWithStatus(iris.StatusNoContent)
		return
	}

	ctx.Next()
}

func RateLimiterHandler(ctx iris.Context) {
	funcname := "ratelimiterhandler"
	ips := strings.Split(ctx.Request().RemoteAddr, ":")
	ip := ips[0]

	RateLimiter.IPmap.RLock()
	connattempt, ok := RateLimiter.IPmap.Map[ip]
	RateLimiter.IPmap.RUnlock()

	if !ok {
		logger.LoggerEngine.Log(funcname, fmt.Sprintf("adding %s to map", ip))
		defer logger.LoggerEngine.Log(funcname, fmt.Sprintf("done adding %s to map", ip))
		RateLimiter.IPmap.Lock()
		RateLimiter.IPmap.Map[ip] = NewConnectStruct()
		RateLimiter.IPmap.Unlock()

		RateLimiterHandler(ctx)
		return
	}

	defer logger.LoggerEngine.Log(funcname, fmt.Sprintf("done serving ip: %s", ip))

	now := time.Now()
	lastping := <-connattempt.Connect
	ctr := connattempt.Counter.CurrentCount()
	if ctr > 0 && (ctr%3) == 0 {

		// ## this section handles overall access to all urls
		logger.LoggerEngine.Log(funcname, fmt.Sprintf("rate limiting ip: %s", ip))
		<-connattempt.Ticker.C
		diff := now.Sub(lastping)

		// if difference is too small, probably a ddos from a specific ip
		// even after ticker
		// just ban (? indefinitely)
		if diff.Microseconds() < 1000 { // microseconds == 1000th of a millisecond
			logger.LoggerEngine.LogErrF(funcname, "sub 1000microsecond ping %s", ip)
			// BannedIps.Set(ip, now)

			connattempt.Connect <- time.Now()
			return
		}

		// ## this section checks the last access of current url
		url := ctx.Request().URL.Path
		lastfetched, ok := connattempt.UrlMap.Get(url) // ?? PUTTING A RAW URL INTO MY MAP
		if ok {
			diff2 := now.Sub(lastfetched)
			if diff2.Microseconds() < 1000 {
				err := logger.LoggerEngine.LogErrF(funcname, "sub 1000microsecond ping %s %s %d", ip, url, diff2.Microseconds())
				// BannedIps.Set(ip, now)
				Ban(ip, err.Error())

				connattempt.Connect <- time.Now()
				return
			}
		}
		connattempt.UrlMap.Set(url, now)
	}
	ctx.Next()
	connattempt.Counter.Increment()
	connattempt.Connect <- time.Now()
}

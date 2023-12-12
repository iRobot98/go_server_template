package logger

import (
	"context"
	"fmt"

	"strings"

	"time"

	// "log"
	"net/http"
	// "strings"
	// "sync"

	"github.com/kataras/iris/v12"
	"template.com/service/src/base"
)

type LoggerEngineStruct struct {
	ctx      context.Context
	log_chan chan LogItem
}

func (l *LoggerEngineStruct) Log(Loc string, msg string) {
	L := NewLog(Loc, msg)

	go func() {
		l.log_chan <- L
	}()

}

func (l *LoggerEngineStruct) LogErr(Loc string, err error) (Lerr error) {
	L := NewLogErr(Loc, err.Error())
	go func() {
		l.log_chan <- L
	}()
	return &L
}

func (l *LoggerEngineStruct) LogError(Loc string, msg string) (Lerr error) {
	L := NewLogErr(Loc, msg)
	go func() {
		l.log_chan <- L
	}()

	return &L
}

func (l *LoggerEngineStruct) LogErrorln(Loc string, msg ...any) (Lerr error) {
	L := NewLogErr(Loc, fmt.Sprintln(msg...))

	go func() {
		l.log_chan <- L
	}()

	return &L
}

func (l *LoggerEngineStruct) NewLogThread() (err error) {
	var queue []LogItem

	general_log := PersistentLogFile{}

	if err = general_log.OpenLog("server_log"); err != nil {
		return
	}
	defer general_log.F.Close()

	todayslog := PersistentLogFile{}
	t := time.Now()
	filenametoday := func() string { return fmt.Sprintf("daylog/%d_%d_%d", t.Year(), t.Month(), t.Day()) }
	todayfilename := filenametoday()
	if err = todayslog.OpenLog(todayfilename); err != nil {
		return
	}
	defer todayslog.F.Close()

	tm := time.NewTicker(time.Millisecond)

	for loopok := true; loopok; {

		if todayfilename != filenametoday() {
			if err = todayslog.F.Close(); err != nil {
				return
			}
			todayfilename = filenametoday()
			if err = todayslog.OpenLog(todayfilename); err != nil {
				return
			}
		}
		select {
		case <-l.ctx.Done():
			loopok = false
		case item, ok := <-l.log_chan:
			{
				loopok = ok && loopok
				queue = append(queue, item)
			}
			if len(l.log_chan) > 0 {
				continue
			}

		case _, loopok = <-tm.C:

		}

		if len(queue) > 0 {
			for _, item := range queue {
				if len(item.Message) < 1 {
					continue
				}
				general_log.Writeln(item.String(), "\n", End)
				if err != nil {
					return
				}
				todayslog.Writeln(item.String(), "\n", End)
				if err != nil {
					return
				}
				fmt.Println(item.String(), "\n", End)
			}
			clear(queue)
		}
	}

	return nil
}

func NewLoggerEngine(ctx context.Context) *LoggerEngineStruct {
	LoggerEngine.ctx = ctx
	LoggerEngine.log_chan = make(chan LogItem, 200)

	return &LoggerEngine
}

func LogRequest(ctx iris.Context, dur *base.CustomDuration) string {
	var mylog string
	if dur != nil {
		mylog = fmt.Sprintf("req time:\n%s", dur.String())
	}
	mylog = fmt.Sprintf("%s\n%s\n%s %s %s %d", mylog,
		ctx.Request().RemoteAddr,
		ctx.Request().Method,
		ctx.Request().RequestURI,
		ctx.Request().Proto,
		ctx.ResponseWriter().Clone().StatusCode(),
	)

	stringifyHeaders := func(req *http.Request) string {
		logHeaders := ""

		for header, str := range req.Header {
			logHeaders = fmt.Sprintf("%s\n%s: [%s]", logHeaders, header, strings.Join(str, " "))

		}
		return logHeaders
	}
	mylog = fmt.Sprintf("%s\n%s", mylog, stringifyHeaders(ctx.Request()))
	return mylog
}

func (l *LoggerEngineStruct) LogRequest(ctx iris.Context, dur base.CustomDuration) {

	l.Log("LogRequest", LogRequest(ctx, &dur))
}

func (l *LoggerEngineStruct) LogF(loc string, s string, m ...any) {
	l.Log(loc, fmt.Sprintf(s, m...))
}

func (l *LoggerEngineStruct) LogErrF(loc string, s string, m ...any) error {
	return l.LogError(loc, fmt.Sprintf(s, m...))
}

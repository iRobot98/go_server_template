package logger

import (
	// "context"
	"fmt"
	// "os"
	// "strings"
	// "sync"
	"time"

	"template.com/service/src/base"
	// // "net/http"
	// // "strings"
	// // "sync"
	// "github.com/kataras/iris/v12"
)

type LogItem struct {
	time.Time `json:"time"`
	Loc       string `json:"loc"`
	Message   string `json:"message"`
	err       bool
}

func LogError(Loc string, err error, Message string) (msg string) {
	t := time.Now()
	tme := base.FormatTime(&t)
	msg = fmt.Sprintln("\n[ERR]: ", tme, "\nloc: ", Loc, "\nerror:", err.Error(), "\nmessage: ", Message)
	return

}

func (l *LogItem) Error() string {

	if l.err {
		return fmt.Sprintf("[ERR] %s:\n%s\t%s", base.FormatTime(&l.Time), l.Loc, l.Message)
	}
	return ""
}

func (l *LogItem) String() string {
	if l.err {
		return l.Error()
	}

	return fmt.Sprintf("[LOG] %s:\n%s\t%s", base.FormatTime(&l.Time), l.Loc, l.Message)
}

func NewLog(Loc string, msg string) (L LogItem) {
	L.Time = time.Now()
	L.Loc = Loc
	L.Message = msg
	return
}

func NewLogErr(Loc string, msg string) (L LogItem) {
	L = NewLog(Loc, msg)
	L.err = true
	return
}

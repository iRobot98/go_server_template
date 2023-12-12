package logger

import (
	// "context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"template.com/service/src/base"
	// "net/http"
	// "strings"
	// "sync"
	// "github.com/kataras/iris/v12"
)

type PersistentLogFile struct {
	log.Logger
	F *os.File
	sync.RWMutex
	filename string
	opened   time.Time
}

func (p *PersistentLogFile) OpenLog(file_name string) (err error) {
	//	return nil
	p.Lock()
	defer p.Unlock()
	p.filename = file_name
	log_filename := fmt.Sprintf("./%s/%s.txt", os.Getenv("LOG_FOLDER"), file_name)
	fmt.Println("opening ", file_name, "\nname: ", log_filename)
	p.F, err = os.OpenFile(log_filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModeAppend)
	if err != nil {
		return
	}

	p.opened = time.Now()

	fmt.Fprintf(p.F, "%s server restart...", base.FormatTime(&p.opened))

	return
}

func (p *PersistentLogFile) Writeln(m ...string) (err error) {
	err = p.ReopenFile()
	if err != nil {
		return
	}
	p.RLock()
	defer p.RUnlock()

	_, err = fmt.Fprintf(p.F, "%s\n", strings.Join(m, " "))

	return
}

func (p *PersistentLogFile) ReopenFile() (err error) {
	p.RLock()
	tm := p.opened
	f := p.filename
	p.RUnlock()
	if time.Since(tm) > time.Hour {
		err = p.OpenLog(f)
	}

	return
}

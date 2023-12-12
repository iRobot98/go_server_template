package renderengine

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"
	"sync"
	"time"

	"template.com/service/src/logger"
)

var RenderEngine RenderEngineStruct

type RenderEngineStruct struct {
	sync.RWMutex
	Template *template.Template
	buf      chan *bytes.Buffer

	ctx context.Context
}

type FileContent struct {
	file_name string
	contents  string
	dir_name  string
}

var printloaded bool = true

func (rs *RenderEngineStruct) readDirSync(dir_name string) (output []FileContent, err error) {
	funcname := "(RenderEngineStruct)readDirSync(" + dir_name + ")"
	dir, err := os.ReadDir("./views/" + dir_name)
	if err != nil {
		return
	}

	for _, file := range dir {

		if !file.IsDir() && strings.Contains(file.Name(), ".html") {

			file_name := file.Name()
			fpath := "./views/" + dir_name + "/" + file_name
			if printloaded {
				logger.LoggerEngine.LogF(funcname, "[DBUG] loading %s %s", fpath, "...")
			}
			outtmp, err := os.ReadFile(fpath)
			if err != nil {
				return nil, err
			}
			contents := strings.Trim(string(outtmp), " \n\t")
			if len(contents) < 5 {
				continue
			}

			output = append(output, FileContent{
				file_name: file_name,
				contents:  contents,
				dir_name:  dir_name,
			})

		}
	}

	if len(output) < 1 {
		err = logger.LoggerEngine.LogErrF(funcname, "[Error] Unable to load %s", dir_name)
	}

	return
}

func (rs *RenderEngineStruct) UpdateTemplates() (err error) {
	funcname := "UpdateTemplates() (error)"
	conts, err := rs.readDirSync("components")
	if err != nil {
		logger.LogError("UpdateTemplates() ", err, "no output")
		return
	}
	partials, err := rs.readDirSync("partials")
	if err != nil {
		logger.LogError("UpdateTemplates() ", err, "no output")
		return
	}
	conts = append(conts, partials...)

	pages, err := rs.readDirSync("pages")
	if err != nil {
		logger.LogError("UpdateTemplates() ", err, "no output")
		return
	}
	conts = append(conts, pages...)
	output := ""
	for _, val := range conts {
		switch val.dir_name {
		case "pages":
			output = fmt.Sprintf("%s{{define \"%s\"}} %s {{end}}\n", output, val.file_name, val.contents)
		default:
			output = fmt.Sprintf("%s%s\n", output, val.contents)
		}
	}

	rs.Lock()
	defer rs.Unlock()

	if rs.Template != nil {
		rs.Template = nil
	}
	rs.Template, err = template.New("renderengine").Parse(output)
	if printloaded {
		fmt.Println("loaded dirs  templates")
		printloaded = false
	}
	if err != nil {
		err = logger.LoggerEngine.LogErrF(funcname, "problem generating template: %v\ntemplate:\n%s", err, output)
		<-time.After(5 * time.Second)
	}
	//
	return
}

func NewRenderer(ctx context.Context) (*RenderEngineStruct, error) {
	funcname := "NewRenderer(ctx)"
	rs := &RenderEngine
	err := rs.UpdateTemplates()
	if err != nil {
		err = logger.LoggerEngine.LogErrF(funcname, "reading templates %v\n", err)
		return nil, err
	}
	// create buffer pool
	rs.buf = make(chan *bytes.Buffer, 100)
	// populate buffer
	for i := 100; i > 1; i-- {
		rs.buf <- &bytes.Buffer{}
	}

	rs.ctx = ctx

	tm := time.NewTicker(5 * time.Second)
	go func() {
		funcname := "UpdateTemplates()"

		for {
			select {
			case <-tm.C:
				if err := rs.UpdateTemplates(); err != nil {
					logger.LoggerEngine.LogErrF(funcname, "error %v", err)
					return
				}

			case <-ctx.Done():
				logger.LoggerEngine.LogErrF(funcname, "ctx cancelled %v", ctx.Err())

				return

			}
		}
	}()

	return rs, nil

}

func (rs *RenderEngineStruct) Render(name string, data any) (output string, err error) {

	funcname := fmt.Sprintf("(rs).Render(%s,%v)", name, data)

	canceltimer := time.NewTicker(10 * time.Millisecond)
	if rs.Template == nil {
		err = rs.UpdateTemplates()
		if err != nil {
			err = logger.LoggerEngine.LogErr(funcname, err)
			return
		}
	}

	for i := 0; !rs.TryRLock(); i++ {
		if i > 3 {
			err = fmt.Errorf(logger.LogError(funcname, fmt.Errorf("too many lock tries %d", i), "too many lock tries"))
			return
		}
		select {

		case <-rs.ctx.Done():
			err = fmt.Errorf(logger.LogError(funcname, rs.ctx.Err(), "context cancel called"))
			return

		case <-canceltimer.C:
		}
	}
	defer rs.RUnlock()
	buf, ok := <-rs.buf
	if !ok {
		err = fmt.Errorf("buffer channel closed")
		return
	}
	defer func() {
		buf.Reset()
		go func() {
			rs.buf <- buf
		}()
	}()

	err = rs.Template.ExecuteTemplate(buf, name, data)
	if err != nil {
		fmt.Println("(rs).Render(", name, ",", data, ")\nerror:\n", err)
		return

	}
	output = buf.String()
	return

}

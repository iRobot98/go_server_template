package logger

var LoggerEngine LoggerEngineStruct

const (
	End = "----------------"
)

func init() {

	LoggerEngine = LoggerEngineStruct{}
}

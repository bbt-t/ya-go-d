package pkg

import (
	"log"
	"os"
)

//var Log = NewLog(LogFileNames{}.InitDefaultValues())

type LogInterface interface {
	Info(msg string)
	Warn(msg string)
	Err(err error)
	Fatal(err error)
	Close()
}

type logger struct {
	infoLogger, warningLogger, errorLogger, fatalLogger *log.Logger
	files                                               []*os.File
}

func NewLog(logFileNames *LogFileNames) LogInterface {
	return newLogger(logFileNames)
}

type LogFileNames struct {
	Warn, Err, Fatal string
}

func (n LogFileNames) InitDefaultValues() *LogFileNames {
	if err := os.Mkdir("log", os.ModePerm); err != nil && !os.IsExist(err) {
		panic(err)
	}
	return &LogFileNames{
		Warn:  "log/warning_log",
		Err:   "log/error_log",
		Fatal: "log/fatal_log",
	}
}

func newLogger(logFileNames *LogFileNames) *logger {
	fileW, err := os.OpenFile(logFileNames.Warn, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fileE, err := os.OpenFile(logFileNames.Err, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	fileF, err := os.OpenFile(logFileNames.Fatal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return &logger{
		infoLogger:    log.New(os.Stdout, "[INFO]: ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger: log.New(fileW, "[WARNING]: ", log.Ldate|log.Lmicroseconds|log.Llongfile),
		errorLogger:   log.New(fileE, "[ERROR]:", log.Ldate|log.Lmicroseconds|log.Llongfile),
		fatalLogger:   log.New(fileF, "[FATAL]:", log.Ldate|log.Lmicroseconds|log.Llongfile),

		files: []*os.File{
			fileW, fileE,
		},
	}
}

func (l logger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l logger) Warn(msg string) {
	l.warningLogger.Println(msg)
}
func (l logger) Err(err error) {
	l.errorLogger.Printf("%+v\n", err)
}

func (l logger) Fatal(err error) {
	l.fatalLogger.Fatalf("%+v\n", err)
}

func (l logger) Close() {
	for _, f := range l.files {
		if err := f.Close(); err != nil {
			l.Err(err)
		}
	}
}

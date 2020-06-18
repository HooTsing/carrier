package src

import (
	"io"
	"log"
	"os"
	"runtime"
)

var logger *log.Logger

func _print(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

//Log carrier log
func Log(format string, a ...interface{}) {
	_print(format, a...)
}

//Trace carrier trace
func Trace(format string, a ...interface{}) {
	if LogLevel > 3 {
		_print(format, a...)
	}
}

//Debug carrier debug
func Debug(format string, a ...interface{}) {
	if LogLevel > 2 {
		_print(format, a...)
	}
}

//Info carrier info
func Info(format string, a ...interface{}) {
	if LogLevel > 1 {
		_print(format, a...)
	}
}

//Error carrier error
func Error(format string, a ...interface{}) {
	if LogLevel > 0 {
		_print(format, a...)
	}
}

//LogStack carrier log stack
func LogStack(format string, a ...interface{}) {
	_print(format, a...)

	buf := make([]byte, StackBufferMaxCap)
	runtime.Stack(buf, true)
	_print("==============stack==============")
	_print("%s", buf)
	_print("==============stack==============")
}

//Panic carrier panic
func Panic(format string, a ...interface{}) {
	LogStack(format, a...)
	panic("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
}

func init() {
	logger = log.New(io.Writer(os.Stderr), "", log.Ldate|log.Lmicroseconds)
}

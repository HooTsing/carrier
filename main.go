package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"./src"
)

//IService service interface
type IService interface {
	Start() error
	Status()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func handleSignal(app IService) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)

	for sig := range ch {
		switch sig {
		case syscall.SIGHUP:
			app.Status()
			src.Log("total goroutines: %d", runtime.NumGoroutine())
			os.Exit(1)
		default:
			src.Log("catch signal: %v, exit", sig)
			os.Exit(1)
		}
	}
}

func main() {
	laddr := flag.String("listen", ":8001", "listen address")
	baddr := flag.String("addr", "127.0.0.1:1234", "carrier server address")
	secret := flag.String("secret", "the answer to life, the universe and everything", "carrier secret")
	number := flag.Uint("number", 0, "0 if work as server")
	flag.IntVar(&src.Heartbeat, "heartbeat", 10, "tunnel heartbeat interval")
	flag.UintVar(&src.LogLevel, "log", 1, "log level")

	flag.Usage = usage
	flag.Parse()

	var app IService
	var err error
	if *number == 0 {
		app, err := src.NewServer(*laddr, *baddr, *secret)
		if err != nil {
			src.Log("carrier server start failed, %v, err: %s", app.Start(), err.Error())
		} else {
			src.Log("carrier server start success, %v, err: %s", app.Start(), err.Error())
		}
	} else {
		app, err := src.NewClient(*laddr, *baddr, *secret, *number)
		if err != nil {
			src.Log("carrier client start failed, %v, err: %s", app.Start(), err.Error())
		} else {
			src.Log("carrier client start success, %v, err: %s", app.Start(), err.Error())
		}
	}

	if err != nil {
		src.Log("create service failed: %s\n", err.Error())
		return
	}

	// waiting for signal
	go handleSignal(app)

	// start app
	src.Log("exit: %v", app.Start())
}

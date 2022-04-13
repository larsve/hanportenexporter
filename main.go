package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type signalHandlers map[os.Signal]func()

var data promData

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Specify the serial bridge server(s) to connect to:\nExample\n\thanportenexporter tasmota-D84B8F-2959:8232")
		return
	}

	prometheus.MustRegister(&data)

	appCtx, stopAppCtx := context.WithCancel(context.Background())

	for _, svr := range os.Args[1:] {
		startSerialBridgeClient(appCtx, svr, &data)
	}

	startMetricsServer()

	run(appCtx, signalHandlers{
		syscall.SIGINT:  stopAppCtx,
		syscall.SIGQUIT: stopAppCtx,
		syscall.SIGABRT: stopAppCtx,
		syscall.SIGKILL: stopAppCtx,
		syscall.SIGTERM: stopAppCtx,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	stopSerialBridgeClients()
	stopMetricsServer(ctx)
	cancel()

	log.Println("All good things must come to an end...")
}

func run(ctx context.Context, handlers signalHandlers) {
	log.Println("Application started")
	c := make(chan os.Signal, 1)
	var signals = make([]os.Signal, len(handlers))
	for s := range handlers {
		signals = append(signals, s)
	}
	signal.Notify(c, signals...)
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down application...")
			return
		case sig := <-c:
			f, ok := handlers[sig]
			if ok {
				f()
			} else {
				log.Printf("Unknown/unhandled signal: %s (%v)", sig.String(), sig)
			}
		}
	}
}

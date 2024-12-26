package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	run()
}

func run() {
	log.Println("Start server ...")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()
	log.Println("Got stop signal")
}

package main

import (
	"flag"
	"os"
	"os/signal"
	"routine-server/server"
	"routine-server/spammer"
	"syscall"
)

func main() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGHUP,
	)

	var amountOfSpammers int
	flag.IntVar(&amountOfSpammers, "spammers", 10, "The amount of spammers")

	flag.Parse()

	httpServer := server.NewServer("localhost", "8080")

	spammers := make([]*spammer.Spammer, amountOfSpammers)

	for i := 0; i < amountOfSpammers; i++ {
		spammers[i] = spammer.NewSpammer(httpServer.Addr)
	}

	go func() {
		<-sigChan

		for _, spam := range spammers {
			spam.Stop()
		}

		if err := httpServer.GracefulShutdown(); err != nil {
			panic(err)
		}

		os.Exit(0)
	}()

	for _, spam := range spammers {
		if err := spam.Start(); err != nil {
			panic(err)
		}
	}

	if err := httpServer.Listen(); err != nil {
		panic(err)
	}
}

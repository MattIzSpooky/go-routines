package main

import (
	"os"
	"os/signal"
	"routine-server/server"
	"routine-server/spammer"
	"syscall"
)

const amountOfSpammers = 10

func main() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGKILL,
		syscall.SIGHUP,
	)

	httpServer := server.NewServer("localhost", "8080")

	var spammers []*spammer.Spammer

	for i := 0; i < amountOfSpammers; i++ {
		spammers = append(spammers, spammer.NewSpammer(httpServer.Addr))
	}

	go func() {
		<-sigChan

		for _, spam := range spammers {
			spam.Stop()
		}

		httpServer.GracefulShutdown()
		os.Exit(0)
	}()

	for _, spam := range spammers {
		spam.Start()
	}

	if err := httpServer.Listen(); err != nil {
		panic(err)
	}
}

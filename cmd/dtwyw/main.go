package main

import (
	"log"
	"os"
	"sync"

	"github.com/fernando8franco/dtwyw/internal/config"
)

type state struct {
	cfg *config.Config
	mu  *sync.RWMutex
}

func main() {
	conf, err := config.Read()
	mu := &sync.RWMutex{}
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	programState := state{
		cfg: &conf,
		mu:  mu,
	}

	commands := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	commands.Register("compress", HandlerCompress)

	if len(os.Args) < 2 {
		log.Fatal("not enough arguments were provided")
	}

	cmd := command{
		Name:      os.Args[1],
		Arguments: os.Args[2:],
	}

	err = commands.Run(&programState, cmd)
	if err != nil {
		log.Fatalf("error running the command:\n%v", err)
	}
}

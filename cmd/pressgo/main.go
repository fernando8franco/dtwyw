package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/fernando8franco/pressgo/internal/config"
)

type state struct {
	cfg    *config.Config
	mu     *sync.RWMutex
	client *http.Client
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	log.Fatalf("error getting user home dir: %v", err)
	// }

	programState := state{
		cfg:    &conf,
		mu:     &sync.RWMutex{},
		client: &http.Client{},
	}

	commands := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	commands.Register(credentialsCmd, HandlerCredentials)
	// commands.Register(initCmd, HandlerInit)
	// commands.Register(compressCmd, HandlerCompress)

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

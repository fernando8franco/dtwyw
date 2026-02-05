package main

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fernando8franco/dtwyw/internal/config"
)

type state struct {
	cfg   *config.Config
	paths paths
	mu    *sync.RWMutex
}

type paths struct {
	dtwywDir        string
	pdfsDir         string
	compressPdfsDir string
	configPDFsFile  string
}

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting user home dir: %v", err)
	}
	appPath := filepath.Join(homeDir, dtwywDir)
	pdfsDirPath := filepath.Join(appPath, pdfsDir)
	configPDFsFilePath := filepath.Join(appPath, pdfsDir, cnfPDFsFilename)

	mu := &sync.RWMutex{}

	programState := state{
		cfg: &conf,
		paths: paths{
			pdfsDir:        pdfsDirPath,
			configPDFsFile: configPDFsFilePath,
		},
		mu: mu,
	}

	commands := commands{
		registeredCommands: make(map[string]func(*state, command) error),
	}

	commands.Register("init", HandlerInit)
	// commands.Register("compress", HandlerCompress)

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

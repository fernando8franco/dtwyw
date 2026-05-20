package main

import (
	"context"
	"flag"
	"fmt"

	iloveapi "github.com/fernando8franco/i-love-api-golang"
	"github.com/fernando8franco/pressgo/internal/config"
)

func HandlerCredentials(s *state, cmd command) error {
	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)

	var (
		help = fs.Bool(initHelpFlag, false, "Show help message")
		add  = fs.Bool("add", false, "Add new credentials (id, key)")
	)
	fs.Parse(cmd.Arguments)

	if *help {
		fs.Usage()
		return nil
	}

	if *add {
		cmd.Arguments = fs.Args()
		if len(cmd.Arguments) != 2 {
			fmt.Printf("Only two arguments are valid with -add flag (id, key)")
			return nil
		}
		return addCredential(s, cmd)
	}

	return nil
}

func addCredential(s *state, cmd command) error {
	email := cmd.Arguments[0]
	key := cmd.Arguments[1]

	token, credits, err := validateCredential(s, key)
	if err != nil {
		return fmt.Errorf("Error in validating the credential")
	}

	return s.cfg.AddCredential(email, config.CreateCredential(key, token, credits))
}

func validateCredential(s *state, key string) (string, int, error) {
	api := iloveapi.NewClient(s.client)
	err := api.GenerateToken(context.Background(), key)
	if err != nil {
		return "", 0, err
	}

	start, err := api.Start(context.Background(), iloveapi.StartParams{Tool: toolCompress, Region: region})
	if err != nil {
		return "", 0, err
	}

	return api.GetToken(), start.RemainingCredits, nil
}

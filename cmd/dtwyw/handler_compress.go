package main

import (
	"github.com/fernando8franco/dtwyw/pkg/api"
)

func HandlerCompress(s *state, cmd command) error {
	key := s.cfg.GetActiveKey()
	api.GetToken(key)
	return nil
}

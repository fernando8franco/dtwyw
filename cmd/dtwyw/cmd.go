package main

type command struct {
	Name      string
	Arguments []string
}

type commands struct {
	registeredCommands map[string]func(*state, command) error
}

func (c *commands) Run(s *state, cmd command) error {
	return c.registeredCommands[cmd.Name](s, cmd)
}

func (c *commands) Register(name string, f func(*state, command) error) {
	c.registeredCommands[name] = f
}

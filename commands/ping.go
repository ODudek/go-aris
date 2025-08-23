package commands

import "github.com/ODudek/go-aris/resp"

type PingCommand struct{}

func (p *PingCommand) Execute() (resp.Value, error) {
	return resp.NewStrValue("PONG"), nil
}

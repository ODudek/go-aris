package commands

import "github.com/ODudek/go-aris/resp"

type Command interface {
	Execute() (resp.Value, error)
}

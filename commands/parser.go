package commands

import (
	"errors"
	"github.com/ODudek/go-aris/resp"
	"github.com/ODudek/go-aris/storage"
	"strings"
)

type Parser struct {
	storage *storage.Storage
}

func NewParser(s *storage.Storage) *Parser {
	return &Parser{
		storage: s,
	}
}

func (p *Parser) Parse(value resp.Value) (Command, error) {
	arr, ok := value.(*resp.ArrValue)
	if !ok {
		return nil, errors.New("invalid command format")
	}

	if len(arr.GetArray()) == 0 {
		return nil, errors.New("empty command")
	}

	command := strings.ToUpper(p.extractString(arr.GetArray()[0]))
	args := make([]string, 0, len(arr.GetArray())-1)
	
	for i := 1; i < len(arr.GetArray()); i++ {
		args = append(args, p.extractString(arr.GetArray()[i]))
	}

	switch command {
	case "PING":
		return &PingCommand{}, nil
	case "SET":
		if len(args) < 2 {
			return nil, errors.New("wrong number of arguments for SET")
		}
		return &SetCommand{
			storage: p.storage,
			key:     args[0],
			value:   args[1],
		}, nil
	case "GET":
		if len(args) < 1 {
			return nil, errors.New("wrong number of arguments for GET")
		}
		return &GetCommand{
			storage: p.storage,
			key:     args[0],
		}, nil
	case "HSET":
		if len(args) < 3 {
			return nil, errors.New("wrong number of arguments for HSET")
		}
		return &HSetCommand{
			storage: p.storage,
			key:     args[0],
			field:   args[1],
			value:   args[2],
		}, nil
	case "HGET":
		if len(args) < 2 {
			return nil, errors.New("wrong number of arguments for HGET")
		}
		return &HGetCommand{
			storage: p.storage,
			key:     args[0],
			field:   args[1],
		}, nil
	case "HGETALL":
		if len(args) < 1 {
			return nil, errors.New("wrong number of arguments for HGETALL")
		}
		return &HGetAllCommand{
			storage: p.storage,
			key:     args[0],
		}, nil
	default:
		return nil, errors.New("unknown command: " + command)
	}
}

func (p *Parser) extractString(value resp.Value) string {
	switch v := value.(type) {
	case *resp.BulkValue:
		return string(v.GetBulkString())
	case *resp.StrValue:
		return v.GetString()
	default:
		return ""
	}
}

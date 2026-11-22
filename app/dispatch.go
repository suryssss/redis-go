package main

import (
	"fmt"
	"strings"

	"github.com/suryssss/my_redis_go/internal"
)

func DispatchCommand(RESP internal.RESP) ([]byte, error) {
	result := make([]byte, 0)
	values := []string{}

	RESP.ForEach(func(r internal.RESP) bool {
		values = append(values, r.String())
		return true
	})

	if len(values) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	command := values[0]
	args := values[1:]

	command = strings.ToUpper(command)

	switch command {
	case "PING":
		if len(args) > 0 {
			result = internal.AppendString(result, args[0])
		} else {
			result = append(result, "+PONG\r\n"...)
		}
	case "ECHO":
		if len(args) > 0 {
			result = internal.AppendString(result, args[0])
		} else {
			result = internal.AppendString(result, "")
		}
	case "DOCS":
		result = internal.AppendString(result, "DOCS")
	case "COMMAND":
		result = internal.AppendString(result, "COMMAND")
	default:
		return nil, fmt.Errorf("unkown command %s", command)
	}
	return result, nil

}

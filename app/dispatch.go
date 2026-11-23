package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/suryssss/my_redis_go/internal"
)

type storeItem struct {
	value     string
	expiresAt time.Time // zero time means no expiry
}

var (
	storeMu sync.RWMutex
	store   = make(map[string]storeItem)
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
	case "SET":
		if len(args) < 2 {
			return nil, fmt.Errorf("wrong number of arguments for 'SET'")
		}
		key := args[0]
		val := args[1]
		var expiresAt time.Time

		// check for optional EX / PX expiry flags
		for i := 2; i < len(args); i++ {
			opt := strings.ToUpper(args[i])
			if opt == "PX" && i+1 < len(args) {
				if ms, err := strconv.Atoi(args[i+1]); err == nil {
					expiresAt = time.Now().Add(time.Duration(ms) * time.Millisecond)
				}
				i++
			} else if opt == "EX" && i+1 < len(args) {
				if sec, err := strconv.Atoi(args[i+1]); err == nil {
					expiresAt = time.Now().Add(time.Duration(sec) * time.Second)
				}
				i++
			}
		}

		storeMu.Lock()
		store[key] = storeItem{value: val, expiresAt: expiresAt}
		storeMu.Unlock()

		result = append(result, "+OK\r\n"...)

	case "GET":
		if len(args) < 1 {
			return nil, fmt.Errorf("wrong number of arguments for 'GET'")
		}
		key := args[0]

		storeMu.RLock()
		item, exists := store[key]
		storeMu.RUnlock()

		// key does not exist or has expired
		if !exists || (!item.expiresAt.IsZero() && time.Now().After(item.expiresAt)) {
			if exists {
				storeMu.Lock()
				delete(store, key)
				storeMu.Unlock()
			}
			result = append(result, "$-1\r\n"...)
		} else {
			result = internal.AppendString(result, item.value)
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


package internal

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RESP represents a parsed RESP data type (simple string, bulk string, array, etc.)
type RESP struct {
	Value    string
	Elements []RESP
}

// ForEach iterates over child elements (if this is an array)
func (r RESP) ForEach(fn func(RESP) bool) {
	for _, elem := range r.Elements {
		if !fn(elem) {
			break
		}
	}
}

// String returns the string representation of this RESP element
func (r RESP) String() string {
	return r.Value
}

// AppendString appends a RESP bulk string ($<len>\r\n<value>\r\n) to a byte slice
func AppendString(b []byte, s string) []byte {
	return append(b, fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)...)
}

// Parse reads and decodes a RESP value from a buffered reader
func Parse(r *bufio.Reader) (RESP, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return RESP{}, err
	}
	line = strings.TrimSuffix(line, "\r\n")
	line = strings.TrimSuffix(line, "\n")
	if len(line) == 0 {
		return RESP{}, io.EOF
	}

	// RESP Array: *<count>\r\n
	if line[0] == '*' {
		count, err := strconv.Atoi(line[1:])
		if err != nil {
			return RESP{}, err
		}
		elements := make([]RESP, 0, count)
		for i := 0; i < count; i++ {
			elem, err := Parse(r)
			if err != nil {
				return RESP{}, err
			}
			elements = append(elements, elem)
		}
		return RESP{Elements: elements}, nil
	}

	// RESP Bulk String: $<length>\r\n<value>\r\n
	if line[0] == '$' {
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return RESP{}, err
		}
		if length == -1 {
			return RESP{Value: ""}, nil
		}
		buf := make([]byte, length+2) // data + \r\n
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return RESP{}, err
		}
		return RESP{Value: string(buf[:length])}, nil
	}

	// RESP Simple String: +<value>\r\n
	if line[0] == '+' {
		return RESP{Value: line[1:]}, nil
	}

	// Fallback for inline plain commands (e.g. "PING\r\n")
	parts := strings.Fields(line)
	elements := make([]RESP, 0, len(parts))
	for _, part := range parts {
		elements = append(elements, RESP{Value: part})
	}
	return RESP{Elements: elements}, nil
}

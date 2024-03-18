package jl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Stringer transforms a field returned by the FieldFinder into a string.
type Stringer func(ctx *Context, v interface{}) string

var _ = Stringer(DefaultStringer)
var _ = Stringer(ErrorStringer)

// DefaultStringer attempts to turn a field into string by attempting the following in order
// 1. casting it to a string
// 2. unmarshalling it as a json.RawMessage
// 3. using fmt.Sprintf("%v", input)
func DefaultStringer(ctx *Context, v interface{}) string {
	var s string
	if tmp, ok := v.(string); ok {
		s = tmp
	} else if rawMsg, ok := v.(json.RawMessage); ok {
		var unmarshaled interface{}
		if err := json.Unmarshal(rawMsg, &unmarshaled); err != nil {
			s = string(rawMsg)
		} else {
			s = fmt.Sprintf("%v", unmarshaled)
		}
	} else {
		s = fmt.Sprintf("%v", v)
	}
	return s
}

// ErrorStringer stringifies LogrusError to a multiline string. If the field is not a LogrusError, it falls back
// to the DefaultStringer.
func ErrorStringer(ctx *Context, v interface{}) string {
	w := &bytes.Buffer{}
	if logrusErr, ok := v.(LogrusError); ok {
		w.WriteString("\n  ")
		w.WriteString(logrusErr.Error)
		w.WriteRune('\n')
		// left pad with a tab
		lines := strings.Split(logrusErr.Stack, "\n")
		stackStr := "\t" + strings.Join(lines, "\n\t")
		w.WriteString(stackStr)
		return w.String()
	} else {
		return DefaultStringer(ctx, v)
	}
}

func TraceStringer(ctx *Context, v interface{}) string {
	w := &bytes.Buffer{}
	if rawMsg, ok := v.(json.RawMessage); ok {
		var arr []string
		_ = json.Unmarshal(rawMsg, &arr)
		for _, e := range arr {
			w.WriteRune('\n')
			w.WriteString(e)
		}
	}
	return w.String()
}

type LogException struct {
	File  string   `json:"file"`
	Trace []string `json:"trace"`
}

func ExceptionStringer(ctx *Context, v interface{}) string {
	w := &bytes.Buffer{}
	if rawMsg, ok := v.(json.RawMessage); ok {
		var log LogException
		_ = json.Unmarshal(rawMsg, &log)
		w.WriteString(log.File)
		for _, e := range log.Trace {
			w.WriteRune('\n')
			w.WriteString(e)
		}
	}
	return w.String()
}

type LogExtra struct {
	Class string `json:"class"`
	Line  int    `json:"line"`
}

func ExtraStringer(ctx *Context, v interface{}) string {
	if rawMsg, ok := v.(json.RawMessage); ok {
		var extra LogExtra
		err := json.Unmarshal(rawMsg, &extra)
		if err == nil {
			return extra.Class + ":" + strconv.Itoa(extra.Line)
		}
		var c string
		_ = json.Unmarshal(rawMsg, &c)
		return c
	}
	return ""
}

func LevelStringer(ctx *Context, v interface{}) string {
	if rawMsg, ok := v.(json.RawMessage); ok {
		var str string
		_ = json.Unmarshal(rawMsg, &str)
		switch str {
		case "WARNING":
			return "WARN"
		case "CRITICAL":
			return "CRIT"
		default:
			return str
		}
	}
	return ""
}

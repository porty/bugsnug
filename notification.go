package bugsnug

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type severity string

const (
	severityInfo    severity = "info"
	severityWarning severity = "warning"
	severityError   severity = "error"
)

type notifier struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

type stacktrace struct {
	File         string `json:"file"`
	LineNumber   int    `json:"lineNumber"`
	ColumnNumber int    `json:"columnNumber"`
	Method       string `json:"method"`
	InProject    bool   `json:"inProject"`
	// Code         struct {
	// 	Num1231 string `json:"1231"`
	// 	Num1232 string `json:"1232"`
	// 	Num1233 string `json:"1233"`
	// 	Num1234 string `json:"1234"`
	// 	Num1235 string `json:"1235"`
	// 	Num1236 string `json:"1236"`
	// 	Num1237 string `json:"1237"`
	// } `json:"code"`
}

type exception struct {
	ErrorClass string       `json:"errorClass"`
	Message    string       `json:"message"`
	Stacktrace []stacktrace `json:"stacktrace"`
}

type app struct {
	Version      string `json:"version"`
	ReleaseStage string `json:"releaseStage"`
	Type         string `json:"type"`
}

type event struct {
	PayloadVersion string      `json:"payloadVersion"`
	Exceptions     []exception `json:"exceptions"`
	// Threads []struct {
	// 	ID         string `json:"id"`
	// 	Name       string `json:"name"`
	// 	Stacktrace []struct {
	// 	} `json:"stacktrace"`
	// } `json:"threads"`
	// Context is useful for searching and grouping
	Context string `json:"context"`
	// GroupingHash is really really useful for grouping
	GroupingHash string   `json:"groupingHash"`
	Severity     severity `json:"severity"`
	// User         struct {
	// 	ID    string `json:"id"`
	// 	Name  string `json:"name"`
	// 	Email string `json:"email"`
	// } `json:"user"`
	App app `json:"app"`
	// Device struct {
	// 	OSVersion string `json:"osVersion"`
	// 	Hostname  string `json:"hostname"`
	// } `json:"device"`
	MetaData map[string]map[string]interface{} `json:"metaData"`
}

type Notification struct {
	APIKey   string   `json:"apiKey"`
	Notifier notifier `json:"notifier"`
	Events   []event  `json:"events"`
}

type causer interface {
	Cause() error
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func Notify(err error, apiKey string) error {
	// exceptions := []exception{
	// 	{
	// 		ErrorClass: reflect.TypeOf(err).String(),
	// 		Message:    err.Error(),
	// 		Stacktrace: []stacktrace{},
	// 	},
	// }
	var exceptions []exception

	for ; err != nil; err = nextError(err) {
		exceptions = append(exceptions, exception{
			ErrorClass: reflect.TypeOf(err).String(),
			Message:    err.Error(),
			Stacktrace: getStack(err),
		})
	}

	grouping := []string{}
	for _, e := range exceptions {
		for _, s := range e.Stacktrace {
			grouping = append(grouping, s.Method)
		}
		if len(grouping) == 0 {
			grouping = []string{e.Message, e.ErrorClass}
		}
	}
	groupingHash := strings.Join(grouping, ",")
	n := Notification{
		APIKey: apiKey,
		Notifier: notifier{
			Name:    "Porty's Notifier",
			URL:     "https://7x.io/",
			Version: "0.0.1",
		},
		Events: []event{
			{
				App: app{
					ReleaseStage: "dev",
					Type:         "type",
					Version:      "0.0.1",
				},
				Context:      "main",
				Exceptions:   exceptions,
				GroupingHash: groupingHash,
				MetaData: map[string]map[string]interface{}{
					"request": map[string]interface{}{
						"method": "GET",
						"URL":    "https://www.something.com/",
					},
				},
				PayloadVersion: "2",
				Severity:       severityInfo,
			},
		},
	}

	b, err := json.Marshal(n)
	if err != nil {
		return err
	}
	resp, err := http.Post("https://notify.bugsnag.com/", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("bad response from bugsnag, received " + resp.Status)
	}

	return nil
}

func nextError(err error) error {
	if c, ok := err.(causer); ok {
		return c.Cause()
	}
	return nil
}

func getStack(err error) []stacktrace {
	st := []stacktrace{}

	if s, ok := err.(stackTracer); ok {
		for _, trace := range s.StackTrace() {
			//for i := 0; i < max(0, len(s.StackTrace())-2); i++ {
			//	trace := s.StackTrace()[i]
			// line, _ := strconv.Atoi(fmt.Sprintf("%d", trace))
			// file := fmt.Sprintf("%+s", trace)
			// function := fmt.Sprintf("%n", trace)
			file, line, function := trace.FileLineFunc()
			st = append(st, stacktrace{
				File:       file,
				InProject:  true,
				LineNumber: line,
				Method:     function,
			})
			if function == "main.main" {
				break
			}
		}
	}
	return st
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

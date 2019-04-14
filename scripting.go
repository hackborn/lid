package dlock

import (
	"encoding/json"
	"errors"
	"github.com/hackborn/sqi"
	"io"
	"strings"
)

type ScriptResponse struct {
	History [][]interface{}
}

// RunScript() ingests a string in our script format and returns
// a response. The response will contain all outputs from every
// command found in the script. Answer an error if anything
// goes wrong with preparing a script (but note that I do not
// answer an error from running a command, all output is captured
// by the script response).
func runScript(_script interface{}, s Service) (ScriptResponse, error) {
	resp := ScriptResponse{}
	script, ok := _script.(string)
	if !ok {
		return resp, badRequestErr
	}

	dec := json.NewDecoder(strings.NewReader(script))
	for {
		m := make(map[string]interface{})
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return resp, err
		}
		for k, v := range m {
			r, e := runScriptCommand(k, v, s)
			if e != nil {
				return resp, e
			}
			resp.History = append(resp.History, r)
		}
	}
	return resp, nil
}

func runScriptCommand(command string, script interface{}, s Service) ([]interface{}, error) {
	switch command {
	case lockCmd:
		return runScriptLock(script, s)
	case unlockCmd:
		return runScriptUnlock(script, s)
	}
	return nil, errors.New("Unknown script command (" + command + ")")
}

func runScriptLock(script interface{}, s Service) ([]interface{}, error) {
	req := LockRequest{}
	opts := &LockOpts{}
	err := readScriptJson(script, "/req", &req)
	if err != nil {
		return nil, err
	}
	err = readScriptJson(script, "/opts", opts)
	if err != nil {
		return nil, err
	}
	resp, err := s.Lock(req, opts)
	return []interface{}{resp, err}, nil
}

func runScriptUnlock(script interface{}, s Service) ([]interface{}, error) {
	return nil, nil
}

func readScriptJson(src interface{}, path string, dst interface{}) error {
	v, err := sqi.Eval(path, src, nil)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// ------------------------------------------------------------
// CONST and VAR

const (
	lockCmd   = "l"
	unlockCmd = "u"
)

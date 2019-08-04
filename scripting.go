package lid

import (
	"encoding/json"
	"errors"
	"github.com/hackborn/sqi"
	"io"
	"strings"
	"time"
)

type scriptResponse struct {
	History [][]interface{}
}

// runScript ingests a string in our script format and returns
// a response. The response will contain all outputs from every
// command found in the script. Answer an error if anything
// goes wrong with preparing the script, but note that I do not
// answer an error from running a command, all output is captured
// by the script response.
func runScript(_script interface{}, s Service) (scriptResponse, error) {
	resp := scriptResponse{}
	script, ok := _script.(string)
	if !ok {
		return resp, ErrBadRequest
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
			if r != nil {
				resp.History = append(resp.History, r)
			}
		}
	}
	return resp, nil
}

func runScriptCommand(command string, script interface{}, s Service) ([]interface{}, error) {
	switch command {
	case checkCmd:
		return runScriptCheck(script, s)
	case durCmd:
		return runScriptDur(script, s)
	case lockCmd:
		return runScriptLock(script, s)
	case unlockCmd:
		return runScriptUnlock(script, s)
	}
	return nil, errors.New("Unknown script command (" + command + ")")
}

func runScriptCheck(script interface{}, s Service) ([]interface{}, error) {
	var signature string
	err := readScriptJSON(script, "/sig", &signature)
	if err != nil {
		return nil, err
	}
	resp, err := s.Check(signature)
	return []interface{}{resp, err}, nil
}

func runScriptDur(script interface{}, s Service) ([]interface{}, error) {
	sd, ok := s.(ServiceDebug)
	if !ok {
		return nil, errors.New("Service does not implement ServiceDebug")
	}
	id, ok := script.(float64) // the JSON encoder sets the number to a float
	if !ok {
		return nil, errors.New("runScriptDur() on invalid duration")
	}
	sd.SetDuration(time.Duration(int64(id)))
	return nil, nil
}

func runScriptLock(script interface{}, s Service) ([]interface{}, error) {
	req := LockRequest{}
	opts := &LockOpts{}
	err := readScriptJSON(script, "/req", &req)
	if err != nil {
		return nil, err
	}
	err = readScriptJSON(script, "/opts", opts)
	if err != nil {
		return nil, err
	}
	resp, err := s.Lock(req, opts)
	return []interface{}{resp, err}, nil
}

func runScriptUnlock(script interface{}, s Service) ([]interface{}, error) {
	req := UnlockRequest{}
	err := readScriptJSON(script, "/req", &req)
	if err != nil {
		return nil, err
	}
	resp, err := s.Unlock(req, nil)
	return []interface{}{resp, err}, nil
}

func readScriptJSON(src interface{}, path string, dst interface{}) error {
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
	checkCmd  = "c"
	durCmd    = "dur"
	lockCmd   = "l"
	unlockCmd = "u"
)

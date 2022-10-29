package alfred

import (
	"fmt"

	aw "github.com/deanishe/awgo"
)

const ongoingProcess = "process.json"

type Process struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Step    string `json:"step"`
	Process string `json:"process"`
}

func LoadOngoingProcess(wf *aw.Workflow) (Process, error) {
	// fallback load function doing nothing
	nop := func() (interface{}, error) {
		return 0.0, nil
	}

	var prop Process
	if err := wf.Data.LoadOrStoreJSON(ongoingProcess, 0, nop, &prop); err != nil {
		return Process{}, fmt.Errorf("error loading the ongoing process: %w", err)
	}

	return prop, nil
}

func StoreOngoingProcess(wf *aw.Workflow, prop Process) error {
	if err := wf.Data.StoreJSON(ongoingProcess, prop); err != nil {
		return fmt.Errorf("error storing the ongoing process: %w", err)
	}

	return nil
}

package alfred

import (
	"encoding/json"
	"fmt"

	aw "github.com/deanishe/awgo"

	"github.com/cage1016/alfred-video2ringtone/fs"
)

const ongoingProcess = "process.json"

type Process struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Step    string `json:"step"`
	Process string `json:"process"`
}

func LoadOngoingProcess(wf *aw.Workflow) (Process, error) {
	df := fs.NewDefaultFs(GetOutput(wf))
	content, err := df.ReadFile(ongoingProcess)
	if err != nil {
		return Process{}, fmt.Errorf("error reading the ongoing ringTone: %w", err)
	}

	var data Process
	err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		return Process{}, fmt.Errorf("error unmarshalling the ongoing ringTone: %w", err)
	}

	return data, nil
}

func StoreOngoingProcess(wf *aw.Workflow, prop Process) error {
	df := fs.NewDefaultFs(GetOutput(wf))

	data, err := json.MarshalIndent(prop, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	err = df.WriteFile(ongoingProcess, string(data), true)
	if err != nil {
		return fmt.Errorf("error writing the ongoing ringTone: %w", err)
	}
	return nil
}

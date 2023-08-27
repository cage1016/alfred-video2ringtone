package alfred

import (
	"encoding/json"
	"fmt"

	aw "github.com/deanishe/awgo"

	"github.com/cage1016/alfred-video2ringtone/fs"
)

const ongoingRingTone = "ringtone.json"

type M4a struct {
	Title     string `json:"title"`
	Name      string `json:"name"`
	Info      string `json:"info"`
	CreatedAt int64  `json:"createdAt"`
}

type RingTone struct {
	Items map[string]M4a `json:"items"`
}

func LoadOngoingRingTone(wf *aw.Workflow) (RingTone, error) {
	df := fs.NewDefaultFs(GetOutput(wf))
	content, err := df.ReadFile(ongoingRingTone)
	if err != nil {
		return RingTone{}, fmt.Errorf("error reading the ongoing ringTone: %w", err)
	}

	var data RingTone
	err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		return RingTone{}, fmt.Errorf("error unmarshalling the ongoing ringTone: %w", err)
	}

	return data, nil
}

func StoreOngoingRingTone(wf *aw.Workflow, prop RingTone) error {
	df := fs.NewDefaultFs(GetOutput(wf))

	data, err := json.MarshalIndent(prop, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	err = df.WriteFile(ongoingRingTone, string(data), true)
	if err != nil {
		return fmt.Errorf("error writing the ongoing ringTone: %w", err)
	}
	return nil
}

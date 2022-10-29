package alfred

import (
	"fmt"

	aw "github.com/deanishe/awgo"
)

const ongoingRingTone = "ringtone.json"

type M4a struct {
	Name      string `json:"name"`
	Info      string `json:"info"`
	CreatedAt int64  `json:"createdAt"`
}

type RingTone struct {
	Item map[string]M4a `json:"item"`
}

func LoadOngoingRingTone(wf *aw.Workflow) (RingTone, error) {
	// fallback load function doing nothing
	nop := func() (interface{}, error) {
		return 0.0, nil
	}

	var prop RingTone
	if err := wf.Data.LoadOrStoreJSON(ongoingRingTone, 0, nop, &prop); err != nil {
		return RingTone{}, fmt.Errorf("error loading the ongoing ringTone: %w", err)
	}

	return prop, nil
}

func StoreOngoingRingTone(wf *aw.Workflow, prop RingTone) error {
	if err := wf.Data.StoreJSON(ongoingRingTone, prop); err != nil {
		return fmt.Errorf("error storing the ongoing ringTone: %w", err)
	}

	return nil
}

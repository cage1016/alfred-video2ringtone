package alfred

import (
	"bytes"
	"html/template"

	aw "github.com/deanishe/awgo"
)

const (
	Output       = "OUTPUT"
	SupportSites = "SUPPORT_SITES"
	Duration     = "DURATION"
	FadeIn       = "FADEIN"
	FadeOut      = "FADEOUT"
	Debug        = "DEBUG"
)

const helpTpl = `# HH:MM:SS
e.g. 00:02:00
Start 00:02:00 with {{.Duration}}s duration, {{.FadeIn}} fadeIn, {{.FadeOut}} fadeOut

# HH:MM:SS,duration
e.g. 00:02:00,20
Start 00:02:00 with 20s duration, {{.FadeIn}} fadeIn, {{.FadeOut}} fadeOut

# HH:MM:SS,duration,FadeIn&Out
e.g. 00:02:00,20,2
Start 00:02:00 with 20s duration, 2s fadeIn, 2s fadeOut

# HH:MM:SS,duration,FadeIn,FadeOut
e.g. 00:02:00,20,2,2
Start 00:02:00 with 20s duration, 2s fadeIn, 2s fadeOut

ps. Default duration is {{.Duration}}s, fadeIn is {{.FadeIn}}, fadeOut is {{.FadeOut}}
Tip. iOS maximum duration is 30s.`

func GetDuration(wf *aw.Workflow) string {
	return wf.Config.Get(Duration)
}

func GetFadein(wf *aw.Workflow) string {
	return wf.Config.Get(FadeIn)
}

func GetFadeout(wf *aw.Workflow) string {
	return wf.Config.Get(FadeOut)
}

func GetDebug(wf *aw.Workflow) bool {
	return wf.Config.GetBool(Debug)
}

var Create = func(name, t string) *template.Template {
	return template.Must(template.New(name).Parse(t))
}

func GetHelp(wf *aw.Workflow) string {
	t := Create("help", helpTpl)
	var tpl bytes.Buffer

	data := struct {
		Duration string
		FadeIn   string
		FadeOut  string
	}{
		Duration: GetDuration(wf),
		FadeIn:   GetFadein(wf),
		FadeOut:  GetFadeout(wf),
	}
	if err := t.Execute(&tpl, data); err != nil {
		return err.Error()
	}

	return tpl.String()
}

func GetOutput(wf *aw.Workflow) string {
	return wf.Config.Get(Output)
}

func GetSupportSites(wf *aw.Workflow) string {
	return wf.Config.Get(SupportSites)
}

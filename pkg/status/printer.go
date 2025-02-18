package status

import (
	"fmt"
	"html/template"
	"os"
	"sort"
	"time"

	"github.com/vmware-tanzu/sonobuoy/pkg/plugin/aggregation"
)

type PrintableStatus struct {
	GlobalStatus   string
	CurrentTime    string
	PluginStatuses []PrintablePluginStatus
}

type PrintablePluginStatus struct {
	Name     string
	Status   string
	Result   string
	Progress string
	Message  string
}

var runningStatusTemplate = `{{.CurrentTime}}> Global Status: {{.GlobalStatus}}
{{printf "%-32s | %-10s | %-10s | %-25s | %-50s" "JOB_NAME" "STATUS" "RESULTS" "PROGRESS" "MESSAGE"}}{{range $index, $pl := .PluginStatuses}}
{{printf "%-32s | %-10s | %-10s | %-25s | %-50s" $pl.Name $pl.Status $pl.Result $pl.Progress $pl.Message}}{{end}}
`

func PrintRunningStatus(s *aggregation.Status) error {
	ps := getPrintableRunningStatus(s)
	statusTemplate, err := template.New("statusTemplate").Parse(runningStatusTemplate)
	if err != nil {
		return err
	}

	err = statusTemplate.Execute(os.Stdout, ps)
	return err
}

func getPrintableRunningStatus(s *aggregation.Status) PrintableStatus {
	ps := PrintableStatus{
		GlobalStatus: s.Status,
		CurrentTime:  time.Now().Format(time.RFC1123),
	}

	for _, pl := range s.Plugins {
		var progress string
		var message string

		if pl.Progress != nil {
			progress = fmt.Sprintf("%d/%d (%d failures)", pl.Progress.Completed, pl.Progress.Total, len(pl.Progress.Failures))
		}

		if pl.Status == aggregation.RunningStatus {
			if pl.Progress != nil {
				message = pl.Progress.Message
			}
		} else if pl.ResultStatus == "" {
			message = "waiting for post-processor..."
		} else {
			passCount := pl.ResultStatusCounts["passed"]
			failedCount := pl.ResultStatusCounts["failed"]
			message = fmt.Sprintf("Total tests processed: %d (%d pass / %d failed)", passCount+failedCount, passCount, failedCount)
		}

		pls := PrintablePluginStatus{
			Name:     pl.Plugin,
			Status:   pl.Status,
			Result:   pl.ResultStatus,
			Progress: progress,
			Message:  message,
		}
		ps.PluginStatuses = append(ps.PluginStatuses, pls)
	}

	sort.Slice(ps.PluginStatuses, func(i, j int) bool {
		return ps.PluginStatuses[i].Name < ps.PluginStatuses[j].Name
	})

	return ps
}

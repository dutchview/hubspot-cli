package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

type PipelinesCmd struct {
	List   PipelinesListCmd   `cmd:"" help:"List pipelines."`
	Stages PipelinesStagesCmd `cmd:"" help:"List stages for a pipeline."`
}

type PipelinesListCmd struct {
	ObjectType string `arg:"" default:"tickets" help:"Object type (tickets, deals)."`
	JSON       bool   `short:"j" help:"Output as JSON."`
}

func (c *PipelinesListCmd) Run(client *api.Client) error {
	data, err := client.GetPipelines(c.ObjectType)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var resp struct {
		Results []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Stages []struct {
				ID    string `json:"id"`
				Label string `json:"label"`
			} `json:"stages"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	for _, p := range resp.Results {
		fmt.Printf("Pipeline: %s (ID: %s)\n", p.Label, p.ID)
		if len(p.Stages) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  STAGE ID\tLABEL")
			for _, s := range p.Stages {
				fmt.Fprintf(w, "  %s\t%s\n", s.ID, s.Label)
			}
			w.Flush()
		}
		fmt.Println()
	}
	return nil
}

type PipelinesStagesCmd struct {
	ObjectType string `arg:"" help:"Object type (tickets, deals)."`
	PipelineID string `arg:"" help:"Pipeline ID."`
	JSON       bool   `short:"j" help:"Output as JSON."`
}

func (c *PipelinesStagesCmd) Run(client *api.Client) error {
	data, err := client.GetPipelineStages(c.ObjectType, c.PipelineID)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var resp struct {
		Results []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STAGE ID\tLABEL")
	for _, s := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\n", s.ID, s.Label)
	}
	w.Flush()
	return nil
}

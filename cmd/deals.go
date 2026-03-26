package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

var dealProperties = []string{
	"dealname", "amount", "dealstage", "pipeline",
	"closedate", "createdate", "hs_lastmodifieddate",
	"hubspot_owner_id",
}

type DealsCmd struct {
	List   DealsListCmd   `cmd:"" help:"List deals."`
	Search DealsSearchCmd `cmd:"" help:"Search deals."`
	Get    DealsGetCmd    `cmd:"" help:"Get deal details."`
}

type DealsListCmd struct {
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *DealsListCmd) Run(client *api.Client) error {
	data, err := client.ListDeals(c.Max, dealProperties, c.After)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	resp, err := parseCRMList(data)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tAMOUNT\tSTAGE\tCLOSE DATE")
	for _, d := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			d.ID,
			truncate(prop(d.Properties, "dealname"), 40),
			prop(d.Properties, "amount"),
			prop(d.Properties, "dealstage"),
			formatTimestamp(prop(d.Properties, "closedate")),
		)
	}
	w.Flush()

	if resp.Paging != nil && resp.Paging.Next != nil {
		fmt.Fprintf(os.Stderr, "\nMore results available. Use --after %s\n", resp.Paging.Next.After)
	}
	return nil
}

type DealsSearchCmd struct {
	Query    string `arg:"" optional:"" help:"Search by deal name."`
	Stage    string `short:"s" help:"Filter by deal stage."`
	Pipeline string `short:"p" help:"Filter by pipeline."`
	Max      int    `short:"n" default:"20" help:"Maximum results."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *DealsSearchCmd) Run(client *api.Client) error {
	var filters []map[string]interface{}

	if c.Query != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "dealname",
			"operator":     "CONTAINS_TOKEN",
			"value":        c.Query,
		})
	}
	if c.Stage != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "dealstage",
			"operator":     "EQ",
			"value":        c.Stage,
		})
	}
	if c.Pipeline != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "pipeline",
			"operator":     "EQ",
			"value":        c.Pipeline,
		})
	}

	filterGroups := []map[string]interface{}{}
	if len(filters) > 0 {
		filterGroups = append(filterGroups, map[string]interface{}{
			"filters": filters,
		})
	}

	data, err := client.SearchDeals(filterGroups, dealProperties, c.Max)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	resp, err := parseCRMList(data)
	if err != nil {
		return err
	}

	if len(resp.Results) == 0 {
		fmt.Println("No deals found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tAMOUNT\tSTAGE\tCLOSE DATE")
	for _, d := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			d.ID,
			truncate(prop(d.Properties, "dealname"), 40),
			prop(d.Properties, "amount"),
			prop(d.Properties, "dealstage"),
			formatTimestamp(prop(d.Properties, "closedate")),
		)
	}
	w.Flush()
	return nil
}

type DealsGetCmd struct {
	DealID string `arg:"" help:"Deal ID."`
	JSON   bool   `short:"j" help:"Output as JSON."`
}

func (c *DealsGetCmd) Run(client *api.Client) error {
	data, err := client.GetDeal(c.DealID, dealProperties)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	obj, err := parseCRMObject(data)
	if err != nil {
		return err
	}

	fmt.Printf("Deal: %s (ID: %s)\n", prop(obj.Properties, "dealname"), obj.ID)
	fmt.Printf("Amount: %s\n", prop(obj.Properties, "amount"))
	fmt.Printf("Stage: %s\n", prop(obj.Properties, "dealstage"))
	fmt.Printf("Pipeline: %s\n", prop(obj.Properties, "pipeline"))
	fmt.Printf("Owner: %s\n", prop(obj.Properties, "hubspot_owner_id"))
	fmt.Printf("Close Date: %s\n", formatTimestamp(prop(obj.Properties, "closedate")))
	fmt.Printf("Created: %s\n", formatTimestamp(prop(obj.Properties, "createdate")))
	return nil
}

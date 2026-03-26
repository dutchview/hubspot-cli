package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

var ticketProperties = []string{
	"subject", "content", "hs_pipeline", "hs_pipeline_stage",
	"hs_ticket_priority", "hs_ticket_category", "hubspot_owner_id",
	"createdate", "hs_lastmodifieddate", "closed_date",
	"source_type", "hs_resolution",
	"si_jira_issue_key", "si_jira_issue_link",
	"si_jira_issue_status", "si_jira_issue_summary",
}

type TicketsCmd struct {
	List   TicketsListCmd   `cmd:"" help:"List tickets."`
	Search TicketsSearchCmd `cmd:"" help:"Search tickets."`
	Get    TicketsGetCmd    `cmd:"" help:"Get ticket details."`
	Create TicketsCreateCmd `cmd:"" help:"Create a new ticket."`
	Update TicketsUpdateCmd `cmd:"" help:"Update a ticket."`
	Delete TicketsDeleteCmd `cmd:"" help:"Delete a ticket."`
}

type TicketsListCmd struct {
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *TicketsListCmd) Run(client *api.Client) error {
	data, err := client.ListTickets(c.Max, ticketProperties, c.After)
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
	fmt.Fprintln(w, "ID\tSUBJECT\tPRIORITY\tSTAGE\tCREATED")
	for _, t := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			truncate(prop(t.Properties, "subject"), 50),
			prop(t.Properties, "hs_ticket_priority"),
			prop(t.Properties, "hs_pipeline_stage"),
			formatTimestamp(prop(t.Properties, "createdate")),
		)
	}
	w.Flush()

	if resp.Paging != nil && resp.Paging.Next != nil {
		fmt.Fprintf(os.Stderr, "\nMore results available. Use --after %s\n", resp.Paging.Next.After)
	}
	return nil
}

type TicketsSearchCmd struct {
	Query    string `arg:"" optional:"" help:"Search query text."`
	Pipeline string `short:"p" help:"Pipeline ID to filter by."`
	Stage    string `short:"s" help:"Pipeline stage ID to filter by."`
	Priority string `help:"Priority (HIGH, MEDIUM, LOW)."`
	Owner    string `short:"o" help:"Owner ID to filter by."`
	Max      int    `short:"n" default:"20" help:"Maximum results."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *TicketsSearchCmd) Run(client *api.Client) error {
	var filters []map[string]interface{}

	if c.Pipeline != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "hs_pipeline",
			"operator":     "EQ",
			"value":        c.Pipeline,
		})
	}
	if c.Stage != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "hs_pipeline_stage",
			"operator":     "EQ",
			"value":        c.Stage,
		})
	}
	if c.Priority != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "hs_ticket_priority",
			"operator":     "EQ",
			"value":        c.Priority,
		})
	}
	if c.Owner != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "hubspot_owner_id",
			"operator":     "EQ",
			"value":        c.Owner,
		})
	}
	if c.Query != "" {
		filters = append(filters, map[string]interface{}{
			"propertyName": "subject",
			"operator":     "CONTAINS_TOKEN",
			"value":        c.Query,
		})
	}

	filterGroups := []map[string]interface{}{}
	if len(filters) > 0 {
		filterGroups = append(filterGroups, map[string]interface{}{
			"filters": filters,
		})
	}

	sorts := []map[string]string{
		{"propertyName": "createdate", "direction": "DESCENDING"},
	}

	data, err := client.SearchTickets(filterGroups, ticketProperties, c.Max, 0, sorts)
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
		fmt.Println("No tickets found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSUBJECT\tPRIORITY\tSTAGE\tCREATED")
	for _, t := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			t.ID,
			truncate(prop(t.Properties, "subject"), 50),
			prop(t.Properties, "hs_ticket_priority"),
			prop(t.Properties, "hs_pipeline_stage"),
			formatTimestamp(prop(t.Properties, "createdate")),
		)
	}
	w.Flush()
	return nil
}

type TicketsGetCmd struct {
	TicketID string `arg:"" help:"Ticket ID."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *TicketsGetCmd) Run(client *api.Client) error {
	associations := []string{"contacts", "companies", "conversations"}
	data, err := client.GetTicket(c.TicketID, ticketProperties, associations)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var ticket struct {
		ID         string                 `json:"id"`
		Properties map[string]interface{} `json:"properties"`
		Associations map[string]struct {
			Results []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"results"`
		} `json:"associations"`
	}
	if err := json.Unmarshal(data, &ticket); err != nil {
		return err
	}

	fmt.Printf("Ticket: %s\n", ticket.ID)
	fmt.Printf("Subject: %s\n", prop(ticket.Properties, "subject"))
	fmt.Printf("Priority: %s\n", prop(ticket.Properties, "hs_ticket_priority"))
	fmt.Printf("Pipeline: %s\n", prop(ticket.Properties, "hs_pipeline"))
	fmt.Printf("Stage: %s\n", prop(ticket.Properties, "hs_pipeline_stage"))
	fmt.Printf("Category: %s\n", prop(ticket.Properties, "hs_ticket_category"))
	fmt.Printf("Source: %s\n", prop(ticket.Properties, "source_type"))
	fmt.Printf("Owner: %s\n", prop(ticket.Properties, "hubspot_owner_id"))
	fmt.Printf("Created: %s\n", formatTimestamp(prop(ticket.Properties, "createdate")))
	fmt.Printf("Modified: %s\n", formatTimestamp(prop(ticket.Properties, "hs_lastmodifieddate")))

	if closed := prop(ticket.Properties, "closed_date"); closed != "" {
		fmt.Printf("Closed: %s\n", formatTimestamp(closed))
	}

	if content := prop(ticket.Properties, "content"); content != "" {
		fmt.Printf("\nDescription:\n%s\n", content)
	}

	if resolution := prop(ticket.Properties, "hs_resolution"); resolution != "" {
		fmt.Printf("\nResolution:\n%s\n", resolution)
	}

	// JIRA link
	if jiraKey := prop(ticket.Properties, "si_jira_issue_key"); jiraKey != "" {
		fmt.Printf("\nJIRA: %s (%s)\n", jiraKey, prop(ticket.Properties, "si_jira_issue_status"))
		if link := prop(ticket.Properties, "si_jira_issue_link"); link != "" {
			fmt.Printf("      %s\n", link)
		}
	}

	// Associated contacts
	if assoc, ok := ticket.Associations["contacts"]; ok && len(assoc.Results) > 0 {
		fmt.Printf("\nContacts:\n")
		for _, a := range assoc.Results {
			contactData, err := client.GetContact(a.ID, []string{"firstname", "lastname", "email", "phone", "company"})
			if err != nil {
				fmt.Printf("  - %s (error fetching details)\n", a.ID)
				continue
			}
			contact, _ := parseCRMObject(contactData)
			name := prop(contact.Properties, "firstname") + " " + prop(contact.Properties, "lastname")
			fmt.Printf("  - %s <%s>", name, prop(contact.Properties, "email"))
			if phone := prop(contact.Properties, "phone"); phone != "" {
				fmt.Printf(" %s", phone)
			}
			if company := prop(contact.Properties, "company"); company != "" {
				fmt.Printf(" (%s)", company)
			}
			fmt.Println()
		}
	}

	// Associated companies
	if assoc, ok := ticket.Associations["companies"]; ok && len(assoc.Results) > 0 {
		fmt.Printf("\nCompanies:\n")
		for _, a := range assoc.Results {
			compData, err := client.GetCompany(a.ID, []string{"name", "domain"})
			if err != nil {
				fmt.Printf("  - %s (error fetching details)\n", a.ID)
				continue
			}
			comp, _ := parseCRMObject(compData)
			fmt.Printf("  - %s (%s)\n", prop(comp.Properties, "name"), prop(comp.Properties, "domain"))
		}
	}

	// Associated conversations
	if assoc, ok := ticket.Associations["conversations"]; ok && len(assoc.Results) > 0 {
		fmt.Printf("\nConversation threads: ")
		for i, a := range assoc.Results {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(a.ID)
		}
		fmt.Println()
	}

	return nil
}

type TicketsCreateCmd struct {
	Subject     string `short:"s" required:"" help:"Ticket subject."`
	Description string `short:"d" help:"Ticket description."`
	Pipeline    string `short:"p" help:"Pipeline ID."`
	Stage       string `help:"Pipeline stage ID."`
	Priority    string `help:"Priority (HIGH, MEDIUM, LOW)."`
	Owner       string `short:"o" help:"Owner ID."`
	JSON        bool   `short:"j" help:"Output as JSON."`
}

func (c *TicketsCreateCmd) Run(client *api.Client) error {
	props := map[string]string{
		"subject": c.Subject,
	}
	if c.Description != "" {
		props["content"] = c.Description
	}
	if c.Pipeline != "" {
		props["hs_pipeline"] = c.Pipeline
	}
	if c.Stage != "" {
		props["hs_pipeline_stage"] = c.Stage
	}
	if c.Priority != "" {
		props["hs_ticket_priority"] = c.Priority
	}
	if c.Owner != "" {
		props["hubspot_owner_id"] = c.Owner
	}

	data, err := client.CreateTicket(props)
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

	fmt.Printf("Created ticket %s: %s\n", obj.ID, prop(obj.Properties, "subject"))
	return nil
}

type TicketsUpdateCmd struct {
	TicketID    string `arg:"" help:"Ticket ID."`
	Subject     string `short:"s" help:"New subject."`
	Description string `short:"d" help:"New description."`
	Priority    string `help:"New priority (HIGH, MEDIUM, LOW)."`
	Stage       string `help:"New pipeline stage ID."`
	Owner       string `short:"o" help:"New owner ID."`
	JSON        bool   `short:"j" help:"Output as JSON."`
}

func (c *TicketsUpdateCmd) Run(client *api.Client) error {
	props := map[string]string{}
	if c.Subject != "" {
		props["subject"] = c.Subject
	}
	if c.Description != "" {
		props["content"] = c.Description
	}
	if c.Priority != "" {
		props["hs_ticket_priority"] = c.Priority
	}
	if c.Stage != "" {
		props["hs_pipeline_stage"] = c.Stage
	}
	if c.Owner != "" {
		props["hubspot_owner_id"] = c.Owner
	}

	if len(props) == 0 {
		return fmt.Errorf("no fields to update; use --subject, --description, --priority, --stage, or --owner")
	}

	data, err := client.UpdateTicket(c.TicketID, props)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	fmt.Printf("Updated ticket %s\n", c.TicketID)
	return nil
}

type TicketsDeleteCmd struct {
	TicketID string `arg:"" help:"Ticket ID."`
	Force    bool   `short:"f" help:"Skip confirmation."`
}

func (c *TicketsDeleteCmd) Run(client *api.Client) error {
	if !c.Force {
		if !confirmAction(fmt.Sprintf("Delete ticket %s?", c.TicketID)) {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := client.DeleteTicket(c.TicketID); err != nil {
		return err
	}

	fmt.Printf("Deleted ticket %s\n", c.TicketID)
	return nil
}

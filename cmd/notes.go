package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

var noteProperties = []string{
	"hs_note_body", "hs_timestamp", "hubspot_owner_id",
	"hs_attachment_ids", "hs_createdate", "hs_lastmodifieddate",
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

type NotesCmd struct {
	List   NotesListCmd   `cmd:"" help:"List notes."`
	Search NotesSearchCmd `cmd:"" help:"Search notes by owner."`
	Get    NotesGetCmd    `cmd:"" help:"Get note details."`
	Create NotesCreateCmd `cmd:"" help:"Create a note."`
	Update NotesUpdateCmd `cmd:"" help:"Update a note."`
	Delete NotesDeleteCmd `cmd:"" help:"Delete a note."`
}

type NotesListCmd struct {
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	Owner string `help:"Filter by owner ID."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *NotesListCmd) Run(client *api.Client) error {
	var data []byte
	var err error

	if c.Owner != "" {
		filterGroups := []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"propertyName": "hubspot_owner_id",
						"operator":     "EQ",
						"value":        c.Owner,
					},
				},
			},
		}
		sorts := []map[string]string{
			{"propertyName": "hs_timestamp", "direction": "DESCENDING"},
		}
		data, err = client.SearchNotes(filterGroups, noteProperties, c.Max, 0, sorts)
	} else {
		data, err = client.ListNotes(c.Max, noteProperties, c.After)
	}
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	return printNotesList(data)
}

type NotesSearchCmd struct {
	Query string `arg:"" help:"Search by owner name or owner ID."`
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *NotesSearchCmd) Run(client *api.Client) error {
	// Try to resolve owner name to ID
	ownerID := c.Query
	if !isNumeric(c.Query) {
		id, err := resolveOwnerID(client, c.Query)
		if err != nil {
			return err
		}
		ownerID = id
	}

	filterGroups := []map[string]interface{}{
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "hubspot_owner_id",
					"operator":     "EQ",
					"value":        ownerID,
				},
			},
		},
	}
	sorts := []map[string]string{
		{"propertyName": "hs_timestamp", "direction": "DESCENDING"},
	}

	data, err := client.SearchNotes(filterGroups, noteProperties, c.Max, 0, sorts)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	return printNotesList(data)
}

type NotesGetCmd struct {
	NoteID string `arg:"" help:"Note ID."`
	JSON   bool   `short:"j" help:"Output as JSON."`
}

func (c *NotesGetCmd) Run(client *api.Client) error {
	data, err := client.GetNote(c.NoteID, noteProperties)
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

	body := stripHTML(prop(obj.Properties, "hs_note_body"))
	fmt.Printf("Note ID: %s\n", obj.ID)
	fmt.Printf("Date: %s\n", formatTimestamp(prop(obj.Properties, "hs_timestamp")))
	fmt.Printf("Owner: %s\n", prop(obj.Properties, "hubspot_owner_id"))
	fmt.Printf("\n%s\n", body)
	return nil
}

type NotesCreateCmd struct {
	Body    string   `arg:"" help:"Note body text."`
	Company string   `help:"Associate with company ID."`
	Contact string   `help:"Associate with contact ID."`
	Deal    string   `help:"Associate with deal ID."`
	Ticket  string   `help:"Associate with ticket ID."`
	File    []string `help:"File path to attach (can be repeated)." type:"path"`
	JSON    bool     `short:"j" help:"Output as JSON."`
}

func (c *NotesCreateCmd) Run(client *api.Client) error {
	var attachmentIDs []string
	for _, f := range c.File {
		fmt.Fprintf(os.Stderr, "Uploading %s...\n", f)
		fileID, err := client.UploadFile(f)
		if err != nil {
			return fmt.Errorf("upload %s: %w", f, err)
		}
		fmt.Fprintf(os.Stderr, "Uploaded %s (ID: %s)\n", f, fileID)
		attachmentIDs = append(attachmentIDs, fileID)
	}

	data, err := client.CreateNote(c.Body, attachmentIDs)
	if err != nil {
		return err
	}

	obj, err := parseCRMObject(data)
	if err != nil {
		return err
	}

	// Associate with objects
	associations := []struct {
		objectType string
		objectID   string
	}{
		{"companies", c.Company},
		{"contacts", c.Contact},
		{"deals", c.Deal},
		{"tickets", c.Ticket},
	}

	for _, a := range associations {
		if a.objectID != "" {
			if err := client.CreateAssociation("notes", obj.ID, a.objectType, a.objectID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to associate with %s %s: %v\n", a.objectType, a.objectID, err)
			}
		}
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	fmt.Printf("Created note %s\n", obj.ID)
	return nil
}

type NotesUpdateCmd struct {
	NoteID string `arg:"" help:"Note ID."`
	Body   string `short:"b" required:"" help:"New note body text."`
	JSON   bool   `short:"j" help:"Output as JSON."`
}

func (c *NotesUpdateCmd) Run(client *api.Client) error {
	props := map[string]string{"hs_note_body": c.Body}
	data, err := client.UpdateNote(c.NoteID, props)
	if err != nil {
		return err
	}
	if c.JSON {
		printRawJSON(data)
		return nil
	}
	fmt.Printf("Updated note %s\n", c.NoteID)
	return nil
}

type NotesDeleteCmd struct {
	NoteID string `arg:"" help:"Note ID."`
	Force  bool   `short:"f" help:"Skip confirmation."`
}

func (c *NotesDeleteCmd) Run(client *api.Client) error {
	if !c.Force && !confirmAction(fmt.Sprintf("Delete note %s?", c.NoteID)) {
		return nil
	}
	if err := client.DeleteNote(c.NoteID); err != nil {
		return err
	}
	fmt.Printf("Deleted note %s\n", c.NoteID)
	return nil
}

func printNotesList(data []byte) error {
	resp, err := parseCRMList(data)
	if err != nil {
		return err
	}

	if len(resp.Results) == 0 {
		fmt.Println("No notes found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tNOTE")
	for _, n := range resp.Results {
		body := stripHTML(prop(n.Properties, "hs_note_body"))
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			n.ID,
			formatTimestamp(prop(n.Properties, "hs_timestamp")),
			truncate(body, 100),
		)
	}
	w.Flush()

	if resp.Paging != nil && resp.Paging.Next != nil {
		fmt.Fprintf(os.Stderr, "\nMore results available. Use --after %s\n", resp.Paging.Next.After)
	}
	return nil
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func resolveOwnerID(client *api.Client, name string) (string, error) {
	data, err := client.ListOwners(100, "")
	if err != nil {
		return "", fmt.Errorf("failed to list owners: %w", err)
	}

	var resp struct {
		Results []struct {
			ID        string `json:"id"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}

	nameLower := strings.ToLower(name)
	for _, o := range resp.Results {
		fullName := strings.ToLower(o.FirstName + " " + o.LastName)
		if strings.Contains(fullName, nameLower) {
			fmt.Fprintf(os.Stderr, "Resolved owner: %s %s (ID: %s)\n", o.FirstName, o.LastName, o.ID)
			return o.ID, nil
		}
	}

	return "", fmt.Errorf("no owner found matching %q", name)
}

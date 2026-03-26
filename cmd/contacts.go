package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

var contactProperties = []string{
	"firstname", "lastname", "email", "phone",
	"company", "jobtitle", "lifecyclestage",
	"createdate", "lastmodifieddate",
}

type ContactsCmd struct {
	List   ContactsListCmd   `cmd:"" help:"List contacts."`
	Search ContactsSearchCmd `cmd:"" help:"Search contacts."`
	Get    ContactsGetCmd    `cmd:"" help:"Get contact details."`
}

type ContactsListCmd struct {
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *ContactsListCmd) Run(client *api.Client) error {
	data, err := client.ListContacts(c.Max, contactProperties, c.After)
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
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCOMPANY\tPHONE")
	for _, c := range resp.Results {
		name := prop(c.Properties, "firstname") + " " + prop(c.Properties, "lastname")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.ID,
			name,
			prop(c.Properties, "email"),
			prop(c.Properties, "company"),
			prop(c.Properties, "phone"),
		)
	}
	w.Flush()

	if resp.Paging != nil && resp.Paging.Next != nil {
		fmt.Fprintf(os.Stderr, "\nMore results available. Use --after %s\n", resp.Paging.Next.After)
	}
	return nil
}

type ContactsSearchCmd struct {
	Query string `arg:"" help:"Search by name or email."`
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *ContactsSearchCmd) Run(client *api.Client) error {
	filterGroups := []map[string]interface{}{
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "email",
					"operator":     "CONTAINS_TOKEN",
					"value":        c.Query,
				},
			},
		},
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "firstname",
					"operator":     "CONTAINS_TOKEN",
					"value":        c.Query,
				},
			},
		},
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "lastname",
					"operator":     "CONTAINS_TOKEN",
					"value":        c.Query,
				},
			},
		},
	}

	data, err := client.SearchContacts(filterGroups, contactProperties, c.Max)
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
		fmt.Println("No contacts found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tCOMPANY\tPHONE")
	for _, c := range resp.Results {
		name := prop(c.Properties, "firstname") + " " + prop(c.Properties, "lastname")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			c.ID,
			name,
			prop(c.Properties, "email"),
			prop(c.Properties, "company"),
			prop(c.Properties, "phone"),
		)
	}
	w.Flush()
	return nil
}

type ContactsGetCmd struct {
	ContactID string `arg:"" help:"Contact ID."`
	JSON      bool   `short:"j" help:"Output as JSON."`
}

func (c *ContactsGetCmd) Run(client *api.Client) error {
	data, err := client.GetContact(c.ContactID, contactProperties)
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

	name := prop(obj.Properties, "firstname") + " " + prop(obj.Properties, "lastname")
	fmt.Printf("Contact: %s (ID: %s)\n", name, obj.ID)
	fmt.Printf("Email: %s\n", prop(obj.Properties, "email"))
	fmt.Printf("Phone: %s\n", prop(obj.Properties, "phone"))
	fmt.Printf("Company: %s\n", prop(obj.Properties, "company"))
	fmt.Printf("Job Title: %s\n", prop(obj.Properties, "jobtitle"))
	fmt.Printf("Lifecycle: %s\n", prop(obj.Properties, "lifecyclestage"))
	fmt.Printf("Created: %s\n", formatTimestamp(prop(obj.Properties, "createdate")))
	return nil
}

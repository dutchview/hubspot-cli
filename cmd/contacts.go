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
	Create ContactsCreateCmd `cmd:"" help:"Create a contact."`
	Update ContactsUpdateCmd `cmd:"" help:"Update a contact."`
	Delete ContactsDeleteCmd `cmd:"" help:"Delete a contact."`
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

type ContactsCreateCmd struct {
	Email     string `short:"e" required:"" help:"Email address."`
	FirstName string `help:"First name."`
	LastName  string `help:"Last name."`
	Phone     string `help:"Phone number."`
	Company   string `help:"Company name."`
	JobTitle  string `help:"Job title."`
	JSON      bool   `short:"j" help:"Output as JSON."`
}

func (c *ContactsCreateCmd) Run(client *api.Client) error {
	props := map[string]string{"email": c.Email}
	if c.FirstName != "" {
		props["firstname"] = c.FirstName
	}
	if c.LastName != "" {
		props["lastname"] = c.LastName
	}
	if c.Phone != "" {
		props["phone"] = c.Phone
	}
	if c.Company != "" {
		props["company"] = c.Company
	}
	if c.JobTitle != "" {
		props["jobtitle"] = c.JobTitle
	}

	data, err := client.CreateObject("contacts", props)
	if err != nil {
		return err
	}
	if c.JSON {
		printRawJSON(data)
		return nil
	}
	obj, _ := parseCRMObject(data)
	fmt.Printf("Created contact %s: %s %s (%s)\n", obj.ID, c.FirstName, c.LastName, c.Email)
	return nil
}

type ContactsUpdateCmd struct {
	ContactID string `arg:"" help:"Contact ID."`
	Email     string `short:"e" help:"Email address."`
	FirstName string `help:"First name."`
	LastName  string `help:"Last name."`
	Phone     string `help:"Phone number."`
	Company   string `help:"Company name."`
	JobTitle  string `help:"Job title."`
	JSON      bool   `short:"j" help:"Output as JSON."`
}

func (c *ContactsUpdateCmd) Run(client *api.Client) error {
	props := map[string]string{}
	if c.Email != "" {
		props["email"] = c.Email
	}
	if c.FirstName != "" {
		props["firstname"] = c.FirstName
	}
	if c.LastName != "" {
		props["lastname"] = c.LastName
	}
	if c.Phone != "" {
		props["phone"] = c.Phone
	}
	if c.Company != "" {
		props["company"] = c.Company
	}
	if c.JobTitle != "" {
		props["jobtitle"] = c.JobTitle
	}
	if len(props) == 0 {
		return fmt.Errorf("no fields to update")
	}

	data, err := client.UpdateObject("contacts", c.ContactID, props)
	if err != nil {
		return err
	}
	if c.JSON {
		printRawJSON(data)
		return nil
	}
	fmt.Printf("Updated contact %s\n", c.ContactID)
	return nil
}

type ContactsDeleteCmd struct {
	ContactID string `arg:"" help:"Contact ID."`
	Force     bool   `short:"f" help:"Skip confirmation."`
}

func (c *ContactsDeleteCmd) Run(client *api.Client) error {
	if !c.Force && !confirmAction(fmt.Sprintf("Delete contact %s?", c.ContactID)) {
		return nil
	}
	if err := client.DeleteObject("contacts", c.ContactID); err != nil {
		return err
	}
	fmt.Printf("Deleted contact %s\n", c.ContactID)
	return nil
}

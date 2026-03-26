package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

var companyProperties = []string{
	"name", "domain", "industry", "city", "country",
	"phone", "numberofemployees", "annualrevenue",
	"createdate", "hs_lastmodifieddate",
}

type CompaniesCmd struct {
	List   CompaniesListCmd   `cmd:"" help:"List companies."`
	Search CompaniesSearchCmd `cmd:"" help:"Search companies."`
	Get    CompaniesGetCmd    `cmd:"" help:"Get company details."`
	Create CompaniesCreateCmd `cmd:"" help:"Create a company."`
	Update CompaniesUpdateCmd `cmd:"" help:"Update a company."`
	Delete CompaniesDeleteCmd `cmd:"" help:"Delete a company."`
}

type CompaniesListCmd struct {
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *CompaniesListCmd) Run(client *api.Client) error {
	data, err := client.ListCompanies(c.Max, companyProperties, c.After)
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
	fmt.Fprintln(w, "ID\tNAME\tDOMAIN\tINDUSTRY\tCITY")
	for _, co := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			co.ID,
			truncate(prop(co.Properties, "name"), 40),
			prop(co.Properties, "domain"),
			prop(co.Properties, "industry"),
			prop(co.Properties, "city"),
		)
	}
	w.Flush()

	if resp.Paging != nil && resp.Paging.Next != nil {
		fmt.Fprintf(os.Stderr, "\nMore results available. Use --after %s\n", resp.Paging.Next.After)
	}
	return nil
}

type CompaniesSearchCmd struct {
	Query string `arg:"" help:"Search by name or domain."`
	Max   int    `short:"n" default:"20" help:"Maximum results."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *CompaniesSearchCmd) Run(client *api.Client) error {
	filterGroups := []map[string]interface{}{
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "name",
					"operator":     "CONTAINS_TOKEN",
					"value":        c.Query,
				},
			},
		},
		{
			"filters": []map[string]interface{}{
				{
					"propertyName": "domain",
					"operator":     "CONTAINS_TOKEN",
					"value":        c.Query,
				},
			},
		},
	}

	data, err := client.SearchCompanies(filterGroups, companyProperties, c.Max)
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
		fmt.Println("No companies found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDOMAIN\tINDUSTRY\tCITY")
	for _, co := range resp.Results {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			co.ID,
			truncate(prop(co.Properties, "name"), 40),
			prop(co.Properties, "domain"),
			prop(co.Properties, "industry"),
			prop(co.Properties, "city"),
		)
	}
	w.Flush()
	return nil
}

type CompaniesGetCmd struct {
	CompanyID string `arg:"" help:"Company ID."`
	JSON      bool   `short:"j" help:"Output as JSON."`
}

func (c *CompaniesGetCmd) Run(client *api.Client) error {
	data, err := client.GetCompany(c.CompanyID, companyProperties)
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

	fmt.Printf("Company: %s (ID: %s)\n", prop(obj.Properties, "name"), obj.ID)
	fmt.Printf("Domain: %s\n", prop(obj.Properties, "domain"))
	fmt.Printf("Industry: %s\n", prop(obj.Properties, "industry"))
	fmt.Printf("City: %s\n", prop(obj.Properties, "city"))
	fmt.Printf("Country: %s\n", prop(obj.Properties, "country"))
	fmt.Printf("Phone: %s\n", prop(obj.Properties, "phone"))
	fmt.Printf("Employees: %s\n", prop(obj.Properties, "numberofemployees"))
	fmt.Printf("Revenue: %s\n", prop(obj.Properties, "annualrevenue"))
	fmt.Printf("Created: %s\n", formatTimestamp(prop(obj.Properties, "createdate")))
	return nil
}

type CompaniesCreateCmd struct {
	Name     string `required:"" help:"Company name."`
	Domain   string `help:"Website domain."`
	Industry string `help:"Industry."`
	City     string `help:"City."`
	Country  string `help:"Country."`
	Phone    string `help:"Phone number."`
	JSON     bool   `short:"j" help:"Output as JSON."`
}

func (c *CompaniesCreateCmd) Run(client *api.Client) error {
	props := map[string]string{"name": c.Name}
	if c.Domain != "" {
		props["domain"] = c.Domain
	}
	if c.Industry != "" {
		props["industry"] = c.Industry
	}
	if c.City != "" {
		props["city"] = c.City
	}
	if c.Country != "" {
		props["country"] = c.Country
	}
	if c.Phone != "" {
		props["phone"] = c.Phone
	}

	data, err := client.CreateObject("companies", props)
	if err != nil {
		return err
	}
	if c.JSON {
		printRawJSON(data)
		return nil
	}
	obj, _ := parseCRMObject(data)
	fmt.Printf("Created company %s: %s\n", obj.ID, c.Name)
	return nil
}

type CompaniesUpdateCmd struct {
	CompanyID string `arg:"" help:"Company ID."`
	Name      string `help:"Company name."`
	Domain    string `help:"Website domain."`
	Industry  string `help:"Industry."`
	City      string `help:"City."`
	Country   string `help:"Country."`
	Phone     string `help:"Phone number."`
	JSON      bool   `short:"j" help:"Output as JSON."`
}

func (c *CompaniesUpdateCmd) Run(client *api.Client) error {
	props := map[string]string{}
	if c.Name != "" {
		props["name"] = c.Name
	}
	if c.Domain != "" {
		props["domain"] = c.Domain
	}
	if c.Industry != "" {
		props["industry"] = c.Industry
	}
	if c.City != "" {
		props["city"] = c.City
	}
	if c.Country != "" {
		props["country"] = c.Country
	}
	if c.Phone != "" {
		props["phone"] = c.Phone
	}
	if len(props) == 0 {
		return fmt.Errorf("no fields to update")
	}

	data, err := client.UpdateObject("companies", c.CompanyID, props)
	if err != nil {
		return err
	}
	if c.JSON {
		printRawJSON(data)
		return nil
	}
	fmt.Printf("Updated company %s\n", c.CompanyID)
	return nil
}

type CompaniesDeleteCmd struct {
	CompanyID string `arg:"" help:"Company ID."`
	Force     bool   `short:"f" help:"Skip confirmation."`
}

func (c *CompaniesDeleteCmd) Run(client *api.Client) error {
	if !c.Force && !confirmAction(fmt.Sprintf("Delete company %s?", c.CompanyID)) {
		return nil
	}
	if err := client.DeleteObject("companies", c.CompanyID); err != nil {
		return err
	}
	fmt.Printf("Deleted company %s\n", c.CompanyID)
	return nil
}

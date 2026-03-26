package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dutchview/hubspot-cli/internal/api"
)

type OwnersCmd struct {
	List OwnersListCmd `cmd:"" help:"List owners."`
	Get  OwnersGetCmd  `cmd:"" help:"Get owner details."`
}

type OwnersListCmd struct {
	Max   int    `short:"n" default:"100" help:"Maximum results."`
	After string `help:"Pagination cursor."`
	JSON  bool   `short:"j" help:"Output as JSON."`
}

func (c *OwnersListCmd) Run(client *api.Client) error {
	data, err := client.ListOwners(c.Max, c.After)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var resp struct {
		Results []struct {
			ID        string `json:"id"`
			Email     string `json:"email"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
			UserID    int    `json:"userId"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tEMAIL\tUSER ID")
	for _, o := range resp.Results {
		fmt.Fprintf(w, "%s\t%s %s\t%s\t%d\n",
			o.ID,
			o.FirstName, o.LastName,
			o.Email,
			o.UserID,
		)
	}
	w.Flush()
	return nil
}

type OwnersGetCmd struct {
	OwnerID string `arg:"" help:"Owner ID."`
	JSON    bool   `short:"j" help:"Output as JSON."`
}

func (c *OwnersGetCmd) Run(client *api.Client) error {
	data, err := client.GetOwner(c.OwnerID)
	if err != nil {
		return err
	}

	if c.JSON {
		printRawJSON(data)
		return nil
	}

	var owner struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		UserID    int    `json:"userId"`
	}
	if err := json.Unmarshal(data, &owner); err != nil {
		return err
	}

	fmt.Printf("Owner: %s %s (ID: %s)\n", owner.FirstName, owner.LastName, owner.ID)
	fmt.Printf("Email: %s\n", owner.Email)
	fmt.Printf("User ID: %d\n", owner.UserID)
	return nil
}

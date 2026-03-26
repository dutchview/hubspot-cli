package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/dutchview/hubspot-cli/cmd"
	"github.com/dutchview/hubspot-cli/internal/api"
	"github.com/dutchview/hubspot-cli/internal/config"
)

var version = "dev"

var CLI struct {
	Config string `short:"c" help:"Path to config file (.env format)" type:"path"`

	Tickets       cmd.TicketsCmd       `cmd:"" help:"Manage helpdesk tickets."`
	Conversations cmd.ConversationsCmd `cmd:"" help:"Manage conversation threads."`
	Contacts      cmd.ContactsCmd      `cmd:"" help:"Manage contacts."`
	Companies     cmd.CompaniesCmd     `cmd:"" help:"Manage companies."`
	Deals         cmd.DealsCmd         `cmd:"" help:"Manage deals."`
	Notes         cmd.NotesCmd         `cmd:"" help:"Manage notes/engagements."`
	Owners        cmd.OwnersCmd        `cmd:"" help:"Manage owners/users."`
	Pipelines     cmd.PipelinesCmd     `cmd:"" help:"List pipelines and stages."`
	Configure     cmd.ConfigureCmd     `cmd:"" help:"Show configuration help."`
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			fmt.Printf("hubspot-cli %s\n", version)
			return
		}
	}

	ctx := kong.Parse(&CLI,
		kong.Name("hubspot"),
		kong.Description("A command-line interface for HubSpot CRM ("+version+")"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	switch ctx.Command() {
	case "configure":
		err := ctx.Run()
		ctx.FatalIfErrorf(err)
		return
	}

	cfg, err := config.Load(CLI.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := api.NewClient(cfg.AccessToken)

	err = ctx.Run(client)
	ctx.FatalIfErrorf(err)
}

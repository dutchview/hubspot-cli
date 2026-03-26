package cmd

import "fmt"

type ConfigureCmd struct{}

func (c *ConfigureCmd) Run() error {
	fmt.Println(`HubSpot CLI Configuration
=========================

1. Create a Private App in HubSpot:
   https://app.hubspot.com/private-apps/

   Required scopes:
   - tickets (read/write)
   - crm.objects.contacts.read
   - crm.objects.companies.read
   - crm.objects.deals.read
   - conversations.read
   - conversations.write

2. Create the config file:

   mkdir -p ~/.config/hubspot
   echo "HUBSPOT_ACCESS_TOKEN=your_token_here" > ~/.config/hubspot/.env

3. Test the connection:

   hubspot owners list`)
	return nil
}

package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

type McBroken struct {
	Cities []City  `json:"cities"`
	Broken float64 `json:"broken"`
}

type City struct {
	City   string  `json:"city"`
	Broken float64 `json:"broken"`
}

var McBrokenIntegrationID = "f98d0f70-c537-4fda-ad69-50cb0f1a3013"

var McBrokenCommand = &dgc.Command{
	Name:          "mcbroken",
	Domain:        "astral.integrations.mcbroken",
	Aliases:       []string{"mcbroken"},
	Description:   "Is the ice cream machine broken?",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: McBrokenIntegrationID,
	Handler: func(ctx *dgc.Ctx) {
		httpClient := http.Client{}
		resp, err := httpClient.Get("https://mcbroken.com/stats.json")
		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}
		defer resp.Body.Close()
		var mcBroken McBroken

		err = json.NewDecoder(resp.Body).Decode(&mcBroken)
		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		var cities string
		for _, city := range mcBroken.Cities {
			cities += city.City + ": " + fmt.Sprintf("%g%%", city.Broken) + "\n"
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "McDonalds Machines Broken",
			Description: fmt.Sprintf("%f%% of the machines are broken.", mcBroken.Broken),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Cities",
					Value: cities,
				},
			},
			Color: 0x00FF00,
		}))
	},
}

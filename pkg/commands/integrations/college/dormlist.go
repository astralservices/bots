package integrations

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

type Dorm struct {
	House    string `json:"house"`
	Room     string `json:"room"`
	Resident string `json:"resident"`
}

var DormlistCommand = &dgc.Command{
	Name:        "dormlist",
	Domain:      "astral.integrations.dormlist",
	Aliases:     []string{"dormlist", "dorms"},
	Description: "Get a list of dorms.",
	Category:    "College",
	Usage:       "dormlist",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Arguments: []*discordgo.ApplicationCommandOption{
		{
			Name:        "house",
			Description: "The house to get the dorms for.",
			Type:        discordgo.ApplicationCommandOptionString,
		},
	},
	IntegrationID: CollegeIntegrationID,
	Handler: func(ctx *dgc.Ctx) {
		db := db.New()

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		house := ctx.Arguments.Get(0).Raw()
		data, err := db.GetIntegrationDataForWorkspace(*self.Workspace, CollegeIntegrationID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the integration data.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))
			return
		}

		var dorms []Dorm

		for _, v := range data {
			jsonStr, err := json.Marshal(v.Data)
			if err != nil {
				utils.ErrorHandler(err)
			}
			var d types.CollegeIntegrationData
			err = json.Unmarshal(jsonStr, &d)
			if err != nil {
				utils.ErrorHandler(err)
			}
			dorms = append(dorms, Dorm{
				House:    d.House,
				Room:     d.Room,
				Resident: v.User,
			})
		}

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the dorms.",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Error",
						Value:  err.Error(),
						Inline: false,
					},
				},
				Color: 0xff0000,
			}))
			return
		}

		// room numbers could have letters in them so we sort first by letter, then number
		sort.Slice(dorms, func(i, j int) bool {
			if dorms[i].Room[0] == dorms[j].Room[0] {
				return dorms[i].Room < dorms[j].Room
			}
			return dorms[i].Room[0] < dorms[j].Room[0]
		})

		if house == "" {
			var fields []*discordgo.MessageEmbedField

			var houses []string

			for _, v := range dorms {
				if !utils.StringInSlice(v.House, houses) {
					houses = append(houses, v.House)
				}
			}

			for _, v := range houses {
				// create a field for each house, then set the value to a newline separated list of dorms and residents ([user] - [room])
				val := ""

				for _, v2 := range dorms {
					if v2.House == v {
						val += fmt.Sprintf("<@%s> - %s\n", v2.Resident, v2.Room)
					}
				}

				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   v,
					Value:  val,
					Inline: false,
				})
			}

			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:  "Dorms",
				Color:  0xffff00,
				Fields: fields,
			}))
			return
		} else {
			var fields []*discordgo.MessageEmbedField

			val := ""

			for _, v2 := range dorms {
				if v2.House == house {
					val += fmt.Sprintf("<@%s> - %s\n", v2.Resident, v2.Room)
				}
			}

			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   house,
				Value:  val,
				Inline: false,
			})

			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:  "Dorms",
				Color:  0xffff00,
				Fields: fields,
			}))
			return
		}
	},
}

package integrations

import (
	"fmt"
	"os"
	"strings"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"github.com/shkh/lastfm-go/lastfm"
)

var LastfmIntegrationID = "f98d0f70-c537-4fda-ad69-50cb0f1a3013"

var ScrobblesCommand = &dgc.Command{
	Name:          "scrobbles",
	Domain:        "astral.integrations.scrobbles",
	Aliases:       []string{"scrobbles"},
	Description:   "Get your most recent scrobbles, or someone else's.",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: LastfmIntegrationID,
	Arguments: []*discordgo.ApplicationCommandOption{
		{
			Name:        "user",
			Description: "The user to get scrobbles for.",
			Type:        discordgo.ApplicationCommandOptionUser,
			Required:    false,
		},
	},
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		var user string
		var getProvider bool = true

		if ctx.Arguments.Amount() < 1 {
			user = ctx.Message.Author.ID
		} else if ctx.Arguments.Get(0).AsUserMentionID() != "" {
			user = ctx.Arguments.Get(0).AsUserMentionID()
		} else {
			u := ctx.Arguments.Get(0).Raw()

			// check if u contains only numbers
			isNotDigit := func(c rune) bool { return c < '0' || c > '9' }
			b := strings.IndexFunc(u, isNotDigit) == -1

			if b {
				getProvider = true
			} else {
				getProvider = false
			}

			user = u
		}

		var provider types.Provider

		if getProvider {
			p, err := database.GetProviderFromDiscord(user, "lastfm")

			if err != nil {
				ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
					Title:       "Error",
					Description: "An error occurred while fetching the Last.fm provider.",
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

			provider = p
		} else {
			provider = types.Provider{
				ProviderID: user,
			}
		}

		api := lastfm.New(os.Getenv("LASTFM_API_KEY"), os.Getenv("LASTFM_API_SECRET"))

		res, err := api.User.GetRecentTracks(map[string]interface{}{
			"user": provider.ProviderID,
		})

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the scrobbles.",
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

		if res.Total == 0 {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "No scrobbles found.",
				Color:       0xff0000,
			}))
			return
		}

		embed := discordgo.MessageEmbed{
			Title:       fmt.Sprintf("%s's Scrobbles", res.User),
			Description: "",
			Color:       0x00ff00,
		}
		for i, track := range res.Tracks {
			if i > 5 {
				break
			}
			embed.Description += fmt.Sprintf("`%d` - %s\nBy **%s**\n", i+1, track.Name, track.Artist.Name)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Total scrobbles",
			Value:  fmt.Sprintf("%d", res.Total),
			Inline: false,
		}, &discordgo.MessageEmbedField{
			Name:   "Last.fm Profile",
			Value:  fmt.Sprintf("[%s](https://last.fm/user/%s)", res.User, res.User),
			Inline: false,
		})

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, embed))
	},
}

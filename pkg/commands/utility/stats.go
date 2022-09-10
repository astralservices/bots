package utility

import (
	"fmt"
	"sort"
	"strings"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var StatsCommand = &dgc.Command{
	Name:        "stats",
	Aliases:     []string{"stats"},
	Domain:      "astral.utility.stats",
	Category:    "Utility",
	Usage:       "stats",
	Description: "View command statistics",
	Example:     "stats",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		self := ctx.CustomObjects.MustGet("self").(types.Bot)

		stats, err := database.GetStatistics(*self.ID)

		if err != nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Error",
				Description: "An error occurred while fetching the statistics.",
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

		if len(stats) == 0 {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "No Statistics",
				Description: "No statistics were found.",
				Color:       0xff0000,
			}))
			return
		}

		embed := discordgo.MessageEmbed{
			Title:       "Bot Statistics",
			Description: "Since the bot was added to the server",
			Color:       0x00ff00,
		}

		var commandStats []string

		botStats := types.BotAnalytics{
			Commands: make(map[string]int),
			Messages: 0,
			Members:  0,
		}

		// combine all stats into one
		for _, stat := range stats {
			if stat.Commands != nil {
				for domain, times := range stat.Commands {
					if _, ok := botStats.Commands[domain]; !ok {
						botStats.Commands[domain] = 0
					}

					botStats.Commands[domain] += times
				}
			}
			botStats.Messages += stat.Messages
			botStats.Members += stat.Members
		}

		var keys []string
		for k := range botStats.Commands {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return botStats.Commands[keys[i]] > botStats.Commands[keys[j]]
		})
		for i, k := range keys {
			if i > 5 {
				break
			}
			commandStats = append(commandStats, fmt.Sprintf("`%s`: %d", FindCommand(*ctx, k).Name, botStats.Commands[k]))
		}

		var commandsExecuted int

		for _, times := range botStats.Commands {
			commandsExecuted += times
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Commands Executed",
			Value:  fmt.Sprintf("%d", commandsExecuted),
			Inline: true,
		})

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Most Used Commands",
			Value:  strings.Join(commandStats, "\n"),
			Inline: false,
		})

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, embed))
	},
}

func FindCommand(ctx dgc.Ctx, domain string) dgc.Command {
	for _, command := range ctx.Router.Commands {
		if command.Domain == domain {
			return *command
		}

		for _, alias := range command.Aliases {
			if alias == domain {
				return *command
			}
		}
	}

	domainArr := strings.Split(domain, ".")
	return dgc.Command{
		Name: domainArr[len(domainArr)-1],
	}
}

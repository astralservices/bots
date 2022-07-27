package commands

import (
	"strings"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Help = &dgc.Command{
	Name:        "help",
	Aliases:     []string{"help", "h", "?", "cmds", "commands", "cmd", "command"},
	Category:    "Bot",
	Description: "Show all commands or help for a command",
	Usage:       "help [command]",
	Domain:      "astral.bot.help",
	Example:     "help",
	Handler: func(ctx *dgc.Ctx) {
		embed := discordgo.MessageEmbed{}
		if ctx.Arguments.Amount() == 0 {
			embed.Title = "Command List"
			// get all the commands categories
			categories := make(map[string][]dgc.Command)
			for _, command := range ctx.Router.Commands {
				if command.Category != "" {
					categories[command.Category] = append(categories[command.Category], *command)
				}
			}

			// loop through all the categories
			for category, commands := range categories {
				// create a new field for the category
				field := discordgo.MessageEmbedField{
					Name:   category,
					Value:  "",
					Inline: false,
				}
				// loop through all the commands in the category
				for _, command := range commands {
					// add the command to the field
					field.Value += "`" + command.Name + "` "
				}
				// add the field to the embed
				embed.Fields = append(embed.Fields, &field)
			}

			// send the embed
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, embed))
			return
		}

		// get the command
		command := ctx.Router.GetCmd(ctx.Arguments.Get(0).Raw())

		// if the command doesn't exist
		if command == nil {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Command Not Found",
				Color:       0xff0000,
				Description: "The command `" + ctx.Arguments.Get(0).Raw() + "` was not found.",
			}))
			return
		}

		// create the embed
		embed.Title = "Command: " + command.Name
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Aliases",
				Value: "`" + strings.Join(command.Aliases, "` `") + "`",
			},
			{
				Name:   "Category",
				Value:  command.Category,
				Inline: true,
			},
			{
				Name:   "Domain",
				Value:  "`" + command.Domain + "`",
				Inline: true,
			},
			{
				Name:   "Description",
				Value:  command.Description,
				Inline: false,
			},
			{
				Name:   "Usage",
				Value:  "`" + command.Usage + "`",
				Inline: false,
			},
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, embed))
	},
}

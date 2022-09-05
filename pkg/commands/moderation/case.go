package moderation

import (
	"fmt"
	"os"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var CaseCommand = &dgc.Command{
	Name:          "case",
	Domain:        "astral.moderation.case",
	Aliases:       []string{"case", "report"},
	Usage:         "case <case uuid>",
	Example:       "case ff3f3f3f-3f3f-3f3f-3f3f-3f3f3f3f3f3f",
	Category:      "Moderation",
	Description:   "View a case.",
	Slash:         true,
	SlashGuilds:   []string{os.Getenv("DEV_GUILD")},
	IntegrationID: "",
	Handler: func(ctx *dgc.Ctx) {
		database := db.New()

		if ctx.Arguments.Amount() < 1 {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, fmt.Errorf("Invalid arguments. Please provide a case uuid.")))
			return
		}

		caseId := ctx.Arguments.Get(0).Raw()

		c, err := database.GetReport(caseId)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		embed := utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: fmt.Sprintf("Case `%s`", *c.ID),
			Color: 0x00ff00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "User",
					Value:  fmt.Sprintf("<@%s>", c.User),
					Inline: true,
				},
				{
					Name:   "Moderator",
					Value:  fmt.Sprintf("<@%s>", c.Moderator),
					Inline: true,
				},
				{
					Name:   "Action",
					Value:  cases.Title(language.English, cases.Compact).String(c.Action),
					Inline: true,
				},
				{
					Name:  "Reason",
					Value: c.Reason,
				},
				{
					Name:  "Date",
					Value: fmt.Sprintf("<t:%d>", c.CreatedAt.Unix()),
				},
			},
		})

		ctx.Session.ChannelMessageSendComplex(ctx.Event.ChannelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Delete",
							Style:    discordgo.DangerButton,
							CustomID: fmt.Sprintf("delete_case_%s", *c.ID),
							Emoji: discordgo.ComponentEmoji{
								Name: "🗑️",
							},
						},
					},
				},
			},
			Reference: ctx.Event.Reference(),
		})

		ctx.Session.AddHandlerOnce(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Type == discordgo.InteractionMessageComponent {
				if i.MessageComponentData().CustomID == fmt.Sprintf("delete_case_%s", *c.ID) {
					err := database.DeleteReport(*c.ID)

					if err != nil {
						ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
						return
					}

					embed.Title = fmt.Sprintf("Case `%s` (deleted)", *c.ID)

					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseUpdateMessage,
						Data: &discordgo.InteractionResponseData{
							Embeds: []*discordgo.MessageEmbed{
								embed,
							},
							Components: []discordgo.MessageComponent{},
						},
					})
				}
			}
		})
	},
}

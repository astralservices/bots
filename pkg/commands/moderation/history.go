package moderation

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/Ressetkk/dgwidgets"
	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var HistoryCommand = &dgc.Command{
	Name:        "history",
	Domain:      "astral.moderation.history",
	Aliases:     []string{"history", "modlog", "cases"},
	Category:    "Moderation",
	Description: "Show the moderation history of a user, or the whole guild",
	Usage:       "history [user]",
	Example:     "history @AmusedGrape",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		ctx.Session.MessageReactionAdd(ctx.Event.ChannelID, ctx.Event.Message.ID, "⌛")

		p := dgwidgets.NewPaginator(ctx.Session, ctx.Message.ChannelID)

		database := db.New()

		filter := types.ReportFilter{}

		if ctx.Arguments.Amount() > 0 {
			userId := ctx.Arguments.Get(0).AsUserMentionID()

			if userId == "" {
				userId = ctx.Arguments.Get(0).Raw()
			}

			filter.User = userId
		}

		reports, err := database.GetReportsFiltered(ctx.Message.GuildID, filter)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		// sort reports by date descending
		sort.Slice(reports, func(i, j int) bool {
			return reports[i].CreatedAt.After(*reports[j].CreatedAt)
		})

		var embeds []*discordgo.MessageEmbed

		guild, err := ctx.Session.Guild(ctx.Event.GuildID)

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}

		// 5 reports per page
		for i := 0; i < len(reports); i += 5 {
			end := i + 5

			if end > len(reports) {
				end = len(reports)
			}

			user, err := ctx.Session.User(reports[i].User)

			if err != nil {
				return
			}

			var fields []*discordgo.MessageEmbedField

			for _, report := range reports[i:end] {
				mod, err := ctx.Session.User(report.Moderator)

				if err != nil {
					return
				}

				val := fmt.Sprintf("Reason: %s\nCreated At: <t:%d>\nID: `%s`", report.Reason, report.CreatedAt.Unix(), *report.ID)

				if filter.User == "" {
					val = fmt.Sprintf("User: <@%s>\nReason: %s\nCreated At: <t:%d>\nID: `%s`", report.User, report.Reason, report.CreatedAt.Unix(), *report.ID)
				}

				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s by %s#%s", cases.Title(language.English, cases.Compact).String(report.Action), mod.Username, mod.Discriminator),
					Value:  val,
					Inline: false,
				})
			}

			title := fmt.Sprintf("Moderation history for %s#%s", user.Username, user.Discriminator)

			if filter.User == "" {
				title = fmt.Sprintf("Moderation history for %s", guild.Name)
			}

			embeds = append(embeds, utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:  title,
				Color:  0x00ff00,
				Fields: fields,
			}))
		}

		p.Add(embeds...)

		p.SetPageFooters()

		p.ColourWhenDone = 0xffff

		p.Widget.Timeout = time.Minute * 2

		ctx.Session.MessageReactionRemove(ctx.Event.ChannelID, ctx.Event.Message.ID, "⌛", ctx.Session.State.User.ID)

		err = p.Spawn()

		if err != nil {
			ctx.ReplyEmbed(utils.ErrorEmbed(*ctx, err))
			return
		}
	},
}

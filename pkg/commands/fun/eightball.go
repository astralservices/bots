package commands

import (
	"math/rand"
	"os"
	"time"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Eightball = &dgc.Command{
	Name:        "8ball",
	Domain:      "astral.bot.eightball",
	Aliases:     []string{"8ball", "8b", "eightball"},
	Description: "Ask the magic 8ball a question.",
	Category:    "Fun",
	Usage:       "8ball [question]",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		answers := []string{
			// Positive outcomes
			"It is certain",
			"It is decidedly so",
			"Without a doubt",
			"Yes definitely",
			"You may rely on it",
			"As I see it, yes",
			"Most likely",
			"Outlook good",
			"Yes",
			"Signs point to yes",

			// Neutral outcomes
			"Reply hazy try again",
			"Ask again later",
			"Better not tell you now",
			"Cannot predict now",
			"Concentrate and ask again",

			// Negative outcomes
			"Don't count on it",
			"My reply is no",
			"My sources say no",
			"Outlook not so good",
			"Very doubtful",
		}

		rand.Seed(time.Now().Unix())

		amount := ctx.Arguments.Amount()

		if amount < 2 {
			ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
				Title:       "Not Enough Arguments",
				Description: "You must provide a question to ask the magic 8ball.",
				Color:       0xff0000,
			}))
			return
		}

		question := ctx.Arguments.Raw()

		answer := answers[rand.Intn(len(answers))]

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title: "Magic Eightball :8ball:",
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Question",
					Value: question,
				},
				{
					Name:  "Answer",
					Value: answer,
				},
			},
			Color: 0x00ff00,
		}))
	},
}

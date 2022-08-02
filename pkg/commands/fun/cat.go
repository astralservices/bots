package commands

import (
	"os"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	catapi "github.com/mlemesle/thecatapi-go/api"
)

var Cat = &dgc.Command{
	Name:        "cat",
	Domain:      "astral.fun.cat",
	Aliases:     []string{"cat", "c"},
	Description: "Sends a random cat image from TheCatApi.",
	Category:    "Fun",
	Usage:       "cat",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		cat, err := catapi.NewTheCatAPI("f7a9a450-2853-4c77-8b61-4a5431f110ac")

		if err != nil {
			utils.ErrorHandler(err)
		}

		images, err := cat.GetRandomPublicImage()

		if err != nil {
			utils.ErrorHandler(err)
		}

		randomImage := images[0]

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Cat",
			Description: "Have a cat!",
			Image: &discordgo.MessageEmbedImage{
				URL: randomImage.URL,
			},
			Color: 0x00ff00,
		}))
	},
}

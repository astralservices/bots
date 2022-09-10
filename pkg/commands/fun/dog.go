package commands

import (
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"github.com/thexxiv/dogapi-go/dogapi"
)

var Dog = &dgc.Command{
	Name:        "dog",
	Domain:      "astral.fun.dog",
	Aliases:     []string{"dog"},
	Description: "Sends a random dog image from TheDogApi.",
	Category:    "Fun",
	Usage:       "dog",
	Slash:       true,
	Handler: func(ctx *dgc.Ctx) {
		dog, err := dogapi.RandomImage()

		if err != nil {
			utils.ErrorHandler(err)
		}

		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Dog",
			Description: "Have a dog!",
			Image: &discordgo.MessageEmbedImage{
				URL: dog,
			},
			Color: 0x00ff00,
		}))
	},
}

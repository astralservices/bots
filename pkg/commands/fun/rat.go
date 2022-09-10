package commands

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
)

var Rat = &dgc.Command{
	Name:        "rat",
	Domain:      "astral.bot.rat",
	Aliases:     []string{"rat", "randomrat"},
	Description: "Get a random rat.",
	Category:    "Fun",
	Usage:       "rat",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		r, err := http.Get("https://meme-api.astralapp.io/gimme/rats")
		if err != nil {
			utils.ErrorHandler(err)
		}

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.ErrorHandler(err)
		}

		var rat MemeType
		err = json.Unmarshal(body, &rat)
		if err != nil {
			utils.ErrorHandler(err)
		}
		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Rat",
			Description: "Have a rat!",
			Image: &discordgo.MessageEmbedImage{
				URL: rat.URL,
			},
		}))
	},
}

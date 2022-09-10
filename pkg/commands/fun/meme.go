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

type MemeType struct {
	PostLink  string   `json:"postLink"`
	Subreddit string   `json:"subreddit"`
	Title     string   `json:"title"`
	URL       string   `json:"url"`
	Nsfw      bool     `json:"nsfw"`
	Spoiler   bool     `json:"spoiler"`
	Author    string   `json:"author"`
	UPS       int64    `json:"ups"`
	Preview   []string `json:"preview"`
}

var Meme = &dgc.Command{
	Name:        "meme",
	Domain:      "astral.bot.meme",
	Aliases:     []string{"meme", "randommeme"},
	Description: "Get a random meme.",
	Category:    "Fun",
	Usage:       "meme",
	Slash:       true,
	SlashGuilds: []string{os.Getenv("DEV_GUILD")},
	Handler: func(ctx *dgc.Ctx) {
		// make a request to https://meme-api.astralapp.io/gimme, then use the returned `url` to send an embed to the channel with `url` as the image.
		// use the `utils.GenerateEmbed` function to generate an embed with the title "Meme" and the description "Have a meme!"

		r, err := http.Get("https://meme-api.astralapp.io/gimme")
		if err != nil {
			utils.ErrorHandler(err)
		}

		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.ErrorHandler(err)
		}

		var meme MemeType
		err = json.Unmarshal(body, &meme)
		if err != nil {
			utils.ErrorHandler(err)
		}
		ctx.ReplyEmbed(utils.GenerateEmbed(*ctx, discordgo.MessageEmbed{
			Title:       "Meme",
			Description: "Have a meme!",
			Image: &discordgo.MessageEmbedImage{
				URL: meme.URL,
			},
			Color: 0x00ff00,
		}))
	},
}

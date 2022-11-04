package dalle

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	db "github.com/astralservices/bots/pkg/database/supabase"
	"github.com/astralservices/bots/pkg/types"
	"github.com/astralservices/bots/pkg/utils"
	"github.com/astralservices/dgc"
	d "github.com/astralservices/go-dalle"
	"github.com/bwmarrin/discordgo"
)

var GenerateCommand = &dgc.Command{
	Name:        "generate",
	Aliases:     []string{"generate", "gen"},
	Description: "Generate text from a prompt. The more descriptive the prompt, the better the results.",
	Usage:       "dalle generate <prompt>",
	Example:     "dalle generate A horse on an elevator",
	Category:    "Dalle",
	Domain:      "astral.integrations.dalle.generate",
	Handler: func(c *dgc.Ctx) {
		prompt := c.Arguments.Raw()

		if prompt == "" {
			c.ReplyEmbed(utils.ErrorEmbed(*c, errors.New("You must provide a prompt")))
			return
		}

		database := db.New()

		self := c.CustomObjects.MustGet("self").(types.Bot)

		integration, err := database.GetIntegrationForWorkspace(DalleIntegrationID, *self.Workspace)

		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}

		var apiKey string

		if k, ok := integration.Settings.(map[string]interface{})["apiKey"].(string); !ok {
			c.ReplyEmbed(utils.ErrorEmbed(*c, errors.New("No API key found for integration")))
			return
		} else {
			apiKey = k
		}

		client := d.NewClient(apiKey)

		c.Session.MessageReactionAdd(c.Event.ChannelID, c.Event.Message.ID, "⏳")

		generated, err := client.Generate(prompt, nil, nil, &c.Event.Author.ID, nil)

		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}

		url := generated[0].URL

		resp, err := http.Get(url)

		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}

		defer resp.Body.Close()

		file, err := os.Create(fmt.Sprintf("dalle.%s.png", c.Event.Author.ID))

		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}

		defer file.Close()

		_, err = io.Copy(file, resp.Body)

		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}

		fileUrl := fmt.Sprintf("attachment://%s", fmt.Sprintf("dalle.%s.png", c.Event.Author.ID))

		f, err := os.Open(fmt.Sprintf("dalle.%s.png", c.Event.Author.ID))
		if err != nil {
			c.ReplyEmbed(utils.ErrorEmbed(*c, err))
			return
		}
		defer f.Close()

		c.Session.ChannelMessageSendComplex(c.Message.ChannelID, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				utils.GenerateEmbed(*c, discordgo.MessageEmbed{
					Title:       "Generated Image",
					Description: fmt.Sprintf("Prompt: %s", prompt),
					Image: &discordgo.MessageEmbedImage{
						URL: fileUrl,
					},
				}),
			},
			Files: []*discordgo.File{
				{
					Name:        fmt.Sprintf("dalle.%s.png", c.Event.Author.ID),
					ContentType: "image/png",
					Reader:      f,
				},
			},
			Reference: c.Event.Reference(),
		})

		defer os.Remove(fmt.Sprintf("dalle.%s.png", c.Event.Author.ID))
	},
}

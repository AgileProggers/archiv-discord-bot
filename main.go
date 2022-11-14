package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AgileProggers/archiv-discord-bot/api"
	"github.com/bwmarrin/discordgo"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

var (
	s *discordgo.Session

	// registered commands
	commands = []*discordgo.ApplicationCommand{{
		Name:        "neuestes",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Neuestes Vod anzeigen",
	}, {
		Name:        "suche",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Archiv nach Titel oder Transcript durchsuchen",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "text",
				Description: "Titel oder Transcript",
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
	}, {
		Name:        "uuid",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Vod anhand der UUID anzeigen",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "uuid",
				Description: "UUID des Vods",
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
	}, {
		Name:        "stats",
		Type:        discordgo.ChatApplicationCommand,
		Description: "Archiv Statistiken",
	}}

	// handlers for registered commands
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"neuestes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var response api.SearchResponse
			if err := api.Search(&response, "", 1); err != nil {
				pterm.Error.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Fehler beim Api Request",
					},
				})
				return
			}
			var embeds []*discordgo.MessageEmbed
			for _, vod := range response.Result {
				embeds = append(embeds, &discordgo.MessageEmbed{
					Title: vod.Title,
					URL:   fmt.Sprintf("https://%s/vods/watch/%s", api.FrontendUrl, vod.UUID),
					Image: &discordgo.MessageEmbedImage{
						URL: fmt.Sprintf("https://%s/media/vods/%s-lg.jpg", api.BackendUrl, vod.Filename),
					},
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: fmt.Sprintf("https://%s/media/vods/%s-lg.jpg", api.BackendUrl, vod.Filename),
					},
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Datum",
							Value:  vod.Date.Format("02.01.2006, 15:04:05"),
							Inline: true,
						},
						{
							Name:   "Views",
							Value:  strconv.Itoa(vod.Viewcount),
							Inline: true,
						},
						{
							Name:   "Clips",
							Value:  strconv.Itoa(len(vod.Clips)),
							Inline: true,
						},
					},
					Timestamp: vod.Date.Format(time.RFC3339),
				})
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: embeds,
				},
			})
		},
		"suche": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 || data.Options[0].StringValue() == "" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Suche darf nicht leer sein",
					},
				})
				return
			}
			var response api.SearchResponse
			if err := api.Search(&response, data.Options[0].StringValue(), 10000); err != nil {
				pterm.Error.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Fehler beim Api Request",
					},
				})
				return
			}
			var content string
			if len(response.Result) == 1 {
				content = fmt.Sprintf("%d Vod gefunden\n", len(response.Result))
			} else {
				content = fmt.Sprintf("%d Vods gefunden\n", len(response.Result))
			}
			emotes := []string{
				":one:",
				":two:",
				":three:",
				":four:",
				":five:",
				":six:",
				":seven:",
				":eight:",
				":nine:",
				":keycap_ten:",
			}
			results := response.Result
			if len(results) > 10 {
				results = results[:10]
			}
			for i, vod := range results {
				content += fmt.Sprintf("\n%s [**%s**](<https://%s/vods/watch/%s>)", emotes[i], vod.Title, api.FrontendUrl, vod.UUID)
				content += fmt.Sprintf("\n_:calendar: %s | :eye: %d_ | :bar_chart: Search Score: %.2f \n", vod.Date.Format("02.01.2006, 15:04:05"), vod.Viewcount, vod.TitleRank+vod.TranscriptRank)
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
		},
		"uuid": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			data := i.ApplicationCommandData()
			if len(data.Options) == 0 || data.Options[0].StringValue() == "" {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "UUID darf nicht leer sein",
					},
				})
				return
			}
			var response api.UUIDResponse
			if err := api.UUID(&response, data.Options[0].StringValue()); err != nil {
				pterm.Error.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Fehler beim Api Request",
					},
				})
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{{
						Title: response.Result.Title,
						URL:   fmt.Sprintf("https://%s/vods/watch/%s", api.FrontendUrl, response.Result.UUID),
						Image: &discordgo.MessageEmbedImage{
							URL: fmt.Sprintf("https://%s/media/vods/%s-lg.jpg", api.BackendUrl, response.Result.Filename),
						},
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Datum",
								Value:  response.Result.Date.Format("02.01.2006, 15:04:05"),
								Inline: true,
							},
							{
								Name:   "Views",
								Value:  strconv.Itoa(response.Result.Viewcount),
								Inline: true,
							},
							{
								Name:   "Clips",
								Value:  strconv.Itoa(len(response.Result.Clips)),
								Inline: true,
							},
						},
					}},
				},
			})
		},
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var response api.StatsResponse
			if err := api.Stats(&response); err != nil {
				pterm.Error.Println(err)
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Fehler beim Api Request",
					},
				})
				return
			}

			content := fmt.Sprintf(`
**__Statistiken__**

:chart_with_upwards_trend: **Allgemein**
%d Vods
%d Clips
%.2f Stunden gestreamt

:speaking_head: **Wörter**
%d gesprochene Wörter
%d einzigartige Wörter
%d durchschnittliche Wörter pro Stream

:floppy_disk: **Größe**
%.2fTiB Archivgröße
%.2fMiB Datenbankgröße

:clapper: **Top Clipper**
:one: **%s**, %d Views, %d Clips
:two: **%s**, %d Views, %d Clips
:three: **%s**, %d Views, %d Clips`,
				response.Result.CountVodsTotal, response.Result.CountClipsTotal, response.Result.CountHStreamed,
				response.Result.CountTranscriptWords, response.Result.CountUniqueWords, response.Result.CountAvgWords,
				float64(response.Result.CountSizeBytes)/1024/1024/1024/1024, float64(response.Result.DatabaseSize)/1024/1024,
				response.Result.ClipsPerCreator[0].Name, response.Result.ClipsPerCreator[0].ViewCount, response.Result.ClipsPerCreator[0].ClipCount,
				response.Result.ClipsPerCreator[1].Name, response.Result.ClipsPerCreator[1].ViewCount, response.Result.ClipsPerCreator[1].ClipCount,
				response.Result.ClipsPerCreator[2].Name, response.Result.ClipsPerCreator[2].ViewCount, response.Result.ClipsPerCreator[2].ClipCount,
			)

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})

		},
	}
)

func init() {
	// parse cli args
	var token string
	pflag.StringVarP(&token, "token", "t", "", "The Discord bot token you got from the developer portal")
	pflag.StringVarP(&api.FrontendUrl, "frontend", "f", "archiv.wubbl0rz.tv", "The frontend URL of the archive")
	pflag.StringVarP(&api.BackendUrl, "backend", "b", "api.wubbl0rz.tv", "The backend URL of the archive")
	pflag.Parse()

	if token == "" {
		if os.Getenv("DISCORD_TOKEN") != "" {
			token = os.Getenv("DISCORD_TOKEN")
		} else {
			pterm.Error.Println("Required token missing. Usage:")
			pflag.PrintDefaults()
			os.Exit(1)
		}
	}

	if os.Getenv("ARCHIV_FRONTEND") != "" {
		api.FrontendUrl = os.Getenv("ARCHIV_FRONTEND")
	}

	if os.Getenv("ARCHIV_BACKEND") != "" {
		api.BackendUrl = os.Getenv("ARCHIV_BACKEND")
	}

	// create discord bot
	var err error
	s, err = discordgo.New("Bot " + token)
	if err != nil {
		pterm.Fatal.Println(err.Error())
	}
}

func main() {
	// debug login info
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		pterm.Info.Printfln("Logged in as: %s#%s", s.State.User.Username, s.State.User.Discriminator)
	})

	// print join server
	s.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		pterm.Success.Printfln("Joined server: \"%s\"", g.Name)
	})

	// print leave server
	s.AddHandler(func(s *discordgo.Session, g *discordgo.GuildDelete) {
		pterm.Warning.Printfln("Left server: \"%s\"", g.BeforeDelete.Name)
	})

	// add handlers
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// run bot
	if err := s.Open(); err != nil {
		pterm.Fatal.Println(err.Error())
	}

	// register commands
	createdCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	if err != nil {
		pterm.Fatal.Println(err.Error())
	}
	for _, cmd := range createdCommands {
		pterm.Success.Printfln("Registered command: \"%s\"", cmd.Name)
	}

	pterm.Info.Printfln("Join link: https://discord.com/oauth2/authorize?client_id=%s&scope=applications.commands%%20bot", s.State.User.ID)
	pterm.Info.Println("Bot running. Press ctrl+c to exit.")

	// wait for kill
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	defer s.Close()
	pterm.Info.Println("Bot stopped")
}

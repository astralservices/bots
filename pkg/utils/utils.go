package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"sort"
	"time"

	"github.com/astralservices/dgc"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/sentry-go"
	realtimego "github.com/overseedio/realtime-go"
)

type RealtimeOptions struct {
	Endpoint string
	Key      string
	Table    string

	OnInsert func(m realtimego.Message)
	OnDelete func(m realtimego.Message)
	OnUpdate func(m realtimego.Message)
}

type RealtimePayload[T any] struct {
	Columns         []map[string]string `json:"columns"`
	CommitTimestamp time.Time           `json:"commit_timestamp"`
	Errors          *any                `json:"errors,omitempty"`
	OldRecord       T                   `json:"old_record"`
	Record          T                   `json:"record"`
	Schema          string              `json:"schema"`
	Table           string              `json:"table"`
	Type            string              `json:"type"`
}

func (opts RealtimeOptions) SetupRealtime() {
	c, err := realtimego.NewClient(opts.Endpoint, opts.Key, realtimego.WithUserToken(opts.Key), realtimego.WithHeartbeatInterval(2))

	if err != nil {
		panic(err)
	}

	// connect to server
	err = c.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// create and subscribe to channel
	db := "realtime"
	schema := "public"
	table := opts.Table
	ch, err := c.Channel(realtimego.WithTable(&db, &schema, &table))
	if err != nil {
		log.Fatal(err)
	}

	// setup hooks
	ch.OnInsert = opts.OnInsert
	ch.OnDelete = opts.OnDelete
	ch.OnUpdate = opts.OnUpdate

	// subscribe to channel
	err = ch.Subscribe()
	if err != nil {
		log.Fatal(err)
	}
}

type OrderedMap struct {
	Order []string
	Map   map[string]string
}

func (om *OrderedMap) UnmarshalJSON(b []byte) error {
	json.Unmarshal(b, &om.Map)

	index := make(map[string]int)
	for key := range om.Map {
		om.Order = append(om.Order, key)
		esc, _ := json.Marshal(key) //Escape the key
		index[key] = bytes.Index(b, esc)
	}

	sort.Slice(om.Order, func(i, j int) bool { return index[om.Order[i]] < index[om.Order[j]] })
	return nil
}

func (om OrderedMap) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	l := len(om.Order)
	for i, key := range om.Order {
		km, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(km)
		buf.WriteRune(':')
		vm, err := json.Marshal(om.Map[key])
		if err != nil {
			return nil, err
		}
		buf.Write(vm)
		if i != l-1 {
			buf.WriteRune(',')
		}
		fmt.Println(buf.String())
	}
	buf.WriteRune('}')
	fmt.Println(buf.String())
	return buf.Bytes(), nil
}

func ErrorHandler(err error) error {
	sentry.CaptureException(err)

	log.Println(err)

	return err
}

func ConvertStringToActivityType(in string) discordgo.ActivityType {
	switch in {
	case "PLAYING":
		return discordgo.ActivityTypeGame
	case "STREAMING":
		return discordgo.ActivityTypeStreaming
	case "LISTENING":
		return discordgo.ActivityTypeListening
	case "WATCHING":
		return discordgo.ActivityTypeWatching
	case "COMPETING":
		return discordgo.ActivityTypeCompeting
	}
	return discordgo.ActivityTypeGame
}

type MemoryUsage struct {
	Allocated      string
	AllocatedTotal string
	Sys            string
}

func GetMemoryUsage() MemoryUsage {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mu := MemoryUsage{
		Allocated:      fmt.Sprintf("%.2f MB", float64(m.Alloc)/1024/1024),
		AllocatedTotal: fmt.Sprintf("%.2f MB", float64(m.TotalAlloc)/1024/1024),
		Sys:            fmt.Sprintf("%.2f MB", float64(m.Sys)/1024/1024),
	}

	return mu
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func WithCustomRatelimit(cmd *dgc.Command, ratelimit int64) *dgc.Command {
	// ratelimit is N per minute, so if the ratelimit is 2, they can run the command twice in a minute
	// if the ratelimit is 0, they can run the command as many times as they want
	if ratelimit == 0 {
		return cmd
	}

	// create a time.Duration from the ratelimit. 60 seconds / ratelimit
	// if the ratelimit is 2, the duration is 30 seconds
	// if the ratelimit is 5, the duration is 12 seconds
	duration := time.Duration((60 / ratelimit)) * time.Second

	cmd.RateLimiter = dgc.NewRateLimiter(duration, 1, func(c *dgc.Ctx) {
		c.ReplyEmbed(GenerateEmbed(*c, discordgo.MessageEmbed{
			Title:       "Slow down!",
			Description: "This command is ratelimited!",
			Color:       0xff0000,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Ratelimit",
					Value:  fmt.Sprintf("%d per minute (every %d seconds)", ratelimit, duration/time.Second),
					Inline: true,
				},
			},
		}))
	})
	return cmd
}

func IntPointer(i int) *int {
	return &i
}

func Int64Pointer(i int64) *int64 {
	return &i
}

func BoolPointer(b bool) *bool {
	return &b
}

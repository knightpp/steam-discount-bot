package main

import (
	"os"
	"time"

	"steam-discount/storage"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	tb "gopkg.in/tucnak/telebot.v2"
)

var strg storage.Storager

func getEnvVarOrExit(name string) string {
	v, ok := os.LookupEnv(name)
	if !ok {
		log.Fatalf("no %s env var", name)
	}
	return v
}

func main() {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	_ = godotenv.Load()

	redisAddr := getEnvVarOrExit("REDIS_URL")
	token := getEnvVarOrExit("BOT_TOKEN")
	port := getEnvVarOrExit("PORT")
	publicUrl := getEnvVarOrExit("PUBLIC_URL")

	strg = storage.NewRedis(redisAddr)
	b, err := tb.NewBot(tb.Settings{
		Token: token,
		// Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Poller: &tb.Webhook{
			Listen:   ":" + port,
			Endpoint: &tb.WebhookEndpoint{PublicURL: publicUrl},
		},
	})

	if err != nil {
		log.Fatal(err)
		return
	}
	adder := func(f func(*tb.Bot, *tb.Message)) func(*tb.Message) {
		return func(m *tb.Message) {
			f(b, m)
		}
	}

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, "I am working :)")
	})

	b.Handle("/sub", adder(sub))
	b.Handle("/subs", adder(subs))
	b.Handle("/test", adder(test))

	go func() {
		time.Sleep(30 * time.Second)
		log.Trace("background_refresher slept for 30 second, starting main loop")
		for {
			err := backgroundRefresher2(b)
			if err != nil {
				log.WithError(err).Error("Background refresher returned error")
			}
			log.Trace("background_refresher was executed, sleeping for 4 hours")
			time.Sleep(4 * time.Hour)
		}
	}()
	log.Info("Started")

	b.Start()
}

// Calls background refresher
func test(b *tb.Bot, m *tb.Message) {
	log.Info("test function was called")
	if m.Sender.Username != "knightpp" {
		log.Warnf("test function was called by @%s, denied request", m.Sender.Username)
		return
	}
	err := backgroundRefresher2(b)
	if err != nil {
		log.WithError(err).Error("Background refresher returned error")
	}
}

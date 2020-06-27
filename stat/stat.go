package stat

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Oppen/dcbot/bot"
	"github.com/Oppen/dcbot/module"
)

type Stat struct {
	module.DefaultCommandHandler

	InitTime time.Time
}

func init() {
	module.RegisterModule(&Stat{})
}

var _ module.Module = &Stat{}
var _ module.CommandHandler = &Stat{}

func (s *Stat) Init(*bot.Bot) error {
	s.InitTime = time.Now()
	module.RegisterCommandHandler("stat", s)
	return nil
}

func (s *Stat) HandleCommand(b *bot.Bot, u *tgbotapi.Update) {
	log.Printf("/stat")

	t := time.Now()
	uptime := t.Sub(s.InitTime)
	msgText := fmt.Sprintf("Arriba: %s\n", uptime.String())

	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	mem := memStats.HeapAlloc

	// FIXME: there's probably a less repetitive way that isn't too ugly
	switch {
	case mem > 10 * 1024*1024*1024:
		msgText += fmt.Sprintf("Memoria: %.2fGB\n", float64(mem) / (1024*1024*1024))
	case mem > 10 * 1024*1024:
		msgText += fmt.Sprintf("Memoria: %.2fMB\n", float64(mem) / (1024*1024))
	case mem > 10 * 1024:
		msgText += fmt.Sprintf("Memoria: %.2fkB\n", float64(mem) / 1024)
	default:
		msgText += fmt.Sprintf("Memoria: %dB\n", mem)
	}

	// TODO: rusage for system and user CPU time

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgText)
	if _, err := b.Send(msg); err != nil {
		log.Println(err)
	}
}

func (s *Stat) Help() string {
	return "Indica si el bot está vivo y devuelve estadísticas sobre el sistema y él mismo."
}

package main

/* TODO: pick a logger. Reasonable options are:
 * go.uber.org/zap
 * https://github.com/sirupsen/logrus
 * https://github.com/rs/zerolog
 * https://github.com/apex/log
 * log/syslog
 *
 * PDF parser (for asm):
 * pdfcpu
 * unidoc/pdf
 */
import (
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"sync"
	"syscall"

	"github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/Oppen/dcbot/bot"
	"github.com/Oppen/dcbot/module"
	"github.com/Oppen/dcbot/util/zjson"

	// Handlers, for initialization only.
	_ "github.com/Oppen/dcbot/stat"
	//_ "github.com/Oppen/dcbot/asm"
	//_ "github.com/Oppen/dcbot/plan"
)

const (
	DefaultNumWorkers = 8
	DefaultNumBatches = 1

	WorkQueueLen = 100
	BatchQueueLen = 5
)

var (
	DefaultTTL, _ = time.ParseDuration("24h")
)

type BatchWorkerState struct {
	BatchQueue <-chan tgbotapi.Update
}

type WorkerState struct {
	WorkQueue <-chan tgbotapi.Update
	BatchQueue chan<- tgbotapi.Update
}

func BatchWorker(bot *bot.Bot, worker *BatchWorkerState) {
	// TODO: acknowledge messages as appropriate.
	// Keep in mind requests not fully completed should probably *not* be
	// acknowledged. An option is to persist relevant information for
	// unfinished requests, and restart them or expire them as with other
	// messages.
	// TODO: possibly fire up goroutines before attempting IO, so this one
	// can serve the next job. It would probably be difficult to track, tho.
	// Could be done passing enough metadata for logging and answering and
	// adding to a waitgroup to be waited for at exit.
	// It may be better to do in the specific handlers, or to register IO
	// handlers to send IO tasks, to avoid repetition and passing WG to the
	// regular handlers.
	for {
		work, ok := <-worker.BatchQueue
		_ = work
		// Channel was closed, that's our cue to exit.
		if !ok {
			break
		}
	}
	log.Println("Los metalúrgicos hacen paro")
}

func Worker(b *bot.Bot, worker *WorkerState) {
	for {
		update, ok := <-worker.WorkQueue
		// Channel was closed, that's our queue to exit.
		if !ok {
			break
		}
		log.Println("Update")
		if update.Message == nil {
			continue
		}
		log.Println("Mensaje")
		if date := time.Unix(int64(update.Message.Date), 0); bot.Expired(date, b.Config.TTL) {
			continue
		}
		log.Println("No expirado")
		if !update.Message.IsCommand() {
			continue
		}
		log.Println("Comando")
		handler := module.GetCommandHandler(update.Message.Command())
		if handler.TakesLong() {
			select {
			case worker.BatchQueue<- update:
			default:
				// Send queue full message
			}
			continue
		}
		log.Println("Instantáneo")
		handler.HandleCommand(b, &update)
	}
	log.Println("Los obreros hacen paro")
}

func main() {
	// FIXME
	log := log.New(os.Stderr, "", 0)

	// Saca todo relativo al directorio indicado acá.
	rootDir := os.Getenv("DCBOT_ROOT")

	var b bot.Bot
	if err := zjson.Load(rootDir + "bot_config.zz", &b); err != nil {
		log.Fatal("Config read failed:", err)
	}

	log.Printf("%+v\n", b)

	tgbot, err := tgbotapi.NewBotAPI(b.Config.Token)
	if err != nil {
		log.Fatal("Authorization failed:", err)
	}
	b.BotAPI = tgbot
	b.Debug = true

	defer func() {
		// In case of crash we want a copy of the current config, so an
		// operator can restore manually it if it's in a consistent state
		// without losing too much data.
		// This is also known as "CπCO".
		if r := recover(); r != nil {
			err := zjson.Store(rootDir + "bot_config.panicked.zz", &b.Config)
			if err != nil {
				log.Print("Config stage failed: %w", err)
			}
			// Now we can panic again
			panic(r)
		}
	}()

	log.Printf("Authorized on account %s", b.Self.UserName)

	// Let's register our handlers... Not like before :^)
	module.InitModules(&b)

	if b.Config.TTL == nil {
		b.Config.TTL = &bot.Duration{Duration: DefaultTTL}
	}

	nBatches := b.Config.NumBatches
	if nBatches == 0 {
		nBatches = DefaultNumBatches
	}
	nWorkers := b.Config.NumWorkers
	if nWorkers == 0 {
		nWorkers = DefaultNumWorkers
	}

	BatchWg := sync.WaitGroup{}
	BatchWg.Add(nBatches)

	// FIXME: single element un-acknowledged poll to set the filter.
	// This should go away once the Go API merges the filter feature.
	_, _ = b.MakeRequest("getUpdates", url.Values{
		"offset": []string{"0"},
		"limit": []string{"1"},
		"timeout": []string{"0"},
		"allowed_updates": []string{"[message]"},
	})

	u := tgbotapi.NewUpdate(0)
	// May cause long shutdown, but reduces traffic and CPU usage
	// TODO: use the config
	u.Timeout = 120

	b.Buffer = WorkQueueLen
	updateQueue, err := b.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("GetUpdatesChan failed:", err)
	}

	batchQueue := make(chan tgbotapi.Update, BatchQueueLen)
	batchesState := make([]BatchWorkerState, nBatches)
	for i := 0; i < nBatches; i++ {
		go func(i int) {
			defer BatchWg.Done()

			batchesState[i].BatchQueue = batchQueue
			BatchWorker(&b, &batchesState[i])
		}(i)
	}

	WorkerWg := sync.WaitGroup{}
	WorkerWg.Add(nWorkers)

	workersState := make([]WorkerState, nWorkers)
	for i := 0; i < nWorkers; i++ {
		go func(i int) {
			defer WorkerWg.Done()

			workersState[i].WorkQueue = updateQueue
			workersState[i].BatchQueue = batchQueue
			Worker(&b, &workersState[i])
		}(i)
	}

	// Wait for exit signals. Note the `/restart` and `/quit` commands send signals.
	sCh := make(chan os.Signal, 1)
	signal.Notify(sCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT)
	s := <-sCh
	log.Println("Caught", s)

	/*
	 * Quiting from a command handler would be done as follows:
	 * pid := os.Getpid()
	 * p, _ := os.FindProcess(pid)
	 * p.Signal(os.Interrupt)
	 */

	// Closing each queue and waiting for the affected routines, first for
	// regular workers and then for batch workers, guarantee all updates get
	// processed before we quit.
	// `StopReceivingUpdates` closes the API update channel, so no need to
	// do it manually.
	log.Println("Let's stop")
	b.StopReceivingUpdates()
	log.Println("Wait the workers")
	WorkerWg.Wait()
	close(batchQueue)
	log.Println("Wait the heavy workers")
	BatchWg.Wait()

	// FIXME:
	// Just storing the number of the next one breaks if a week since the
	// last update has passed, because the docs state that in that case the
	// sequence number is not strictly increasing from the last one, but
	// random, while the acknowledgement happens when you ask for an update
	// with a sequence number greater than the last one received, which the
	// Go API can't do automatically, causing the last issued update to be
	// duplicated.
	// Thus, we just ask for the next update to mark the previous ones
	// acknowledged and ignore the answer.
	// Because the last ID is only updated in a copy of the config, we need
	// to get that one, too, and it's easier this way.

	if err := zjson.Store(rootDir + "bot_config.zz", &b.Config); err != nil {
		log.Fatal("Config write failed:", err)
	}
}

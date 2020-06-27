package bot

import (
	"sync"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MediaCacheEntry struct {
	Path     string `json:"path"`
	FileID   string `json:"file_id"`
	Checksum string `json:"checksum"`
}

type MediaCache struct {
	Entries  map[string]*MediaCacheEntry `json:"entries"`
	mx       sync.RWMutex
}

type Config struct {
	// Don't persist, it'd be meaningless
	Root  string `json:"-"`
	Token string `json:"api_token"`
	NumWorkers int `json:"workers"`
	NumBatches int `json:"batches"`
	PollTimeout int `json:"poll_timeout"`
	TTL *Duration `json:"update_ttl"`
}

type Bot struct {
	*tgbotapi.BotAPI `json:"-"`
	Config Config `json:"global_config"`
	MediaCache MediaCache `json:"media_cache"`
}

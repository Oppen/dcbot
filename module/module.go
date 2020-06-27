package module

import (
	"fmt"
	"log"
	"reflect"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Oppen/dcbot/bot"
)

// TODO: get some introspection about the loaded modules and commands,
// especially for reporting via `/stat`.

// Go doesn't have superb support for runtime modules.
// Once loaded, you can't unload it, and it only initializes once.
// This means you need to restart the application to update components, which
// is the most important usage for it in this context.
// Size is also not its strong point.
// Thus, we only care about modules as a way to structure code.
// Any new functionality should register a `Module` with `RegisterModule` to be
// able to initialize all parts that depend on the main configuration and/or
// the connection with Telegram to be up.
// This should be done in the `init` function of each package.
// It's advised to define an empty struct type to use as tag, unless there's
// a strong need for it.
// Most of the work done in `Init` should point towards properly installing the
// `CommandHandler`s.
type Module interface {
	Init(*bot.Bot) error
}

// TODO: store the name for centralized error logging
var moduleInitializers = []Module{}

func InitModules(b *bot.Bot) {
	for _, m := range moduleInitializers {
		moduleName := reflect.TypeOf(m).String()
		log.Println("Initializing", moduleName)
		if err := m.Init(b); err != nil {
			log.Println(moduleName, "initialization failed :(")
		}
	}
}

func RegisterModule(module Module) {
	moduleInitializers = append(moduleInitializers, module)
}

// Generally, what we refer as "functionality" is really one or more commands.
// Those should be registered by the module's `Init` function by calling
// `RegisterCommandHandler` with the command name and a `CommandHandler`.
// A module may register as many handlers as it sees fit.
// `CommandHandler`s implement a `Help` method that returns a description of
// what it does and usage instructions and a `HandleCommand` one that responds
// to an update.
// It also needs to tell the user whether it requires privileges to run, byt
// implementing `RequiresPrivileges` and if it needs to be batched and move on
// with `TakesLong`.
// TODO: should all commands just return a message? And/or an error?
// Even long running ones should generally response both the acknowledgement of
// the request and its result at the end.
type CommandHandler interface {
	HandleCommand(*bot.Bot, *tgbotapi.Update)
	Help() string
	TakesLong() bool
	RequiresPriveleges() bool
}

var handlers = make(map[string]CommandHandler)

// This should be called by either init functions of packages or
// by tests to provide a mock or similar functionality.
func RegisterCommandHandler(cmd string, handler CommandHandler) {
	handlers[cmd] = handler
}

// For unknown commands.
type InvalidCommandHandler struct{}

var _ CommandHandler = InvalidCommandHandler{}

func (InvalidCommandHandler) HandleCommand(b *bot.Bot, u *tgbotapi.Update) {
	msgText := fmt.Sprintf("%s: comando desconocido", u.Message.Command())
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgText)
	if _, err := b.Send(msg); err != nil {
		log.Println(err)
	}
}
func (InvalidCommandHandler) Help() string {
	return "comando inválido"
}
func (InvalidCommandHandler) TakesLong() bool {
	return false
}
func (InvalidCommandHandler) RequiresPriveleges() bool {
	return false
}

// Convenience type to embed in handlers that have a "default" behavior, i.e.,
// don't require special privileges nor do they take longer than otherwise
// acceptable, so they don't have to deal with this boilerplate.
// Note it doesn't implement the whole interface, nor should it, as all
// commands should implement their own handlers and help messages.
type DefaultCommandHandler struct{}

var _ CommandHandler = DefaultCommandHandler{}

func (DefaultCommandHandler) HandleCommand(b *bot.Bot, u *tgbotapi.Update) {
	msgText := fmt.Sprintf("%s: lo estoy procrastinando :(", u.Message.Command())
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgText)
	if _, err := b.Send(msg); err != nil {
		log.Println(err)
	}
}
func (DefaultCommandHandler) Help() string {
	return "no implementado"
}
func (DefaultCommandHandler) TakesLong() bool {
	return false
}
func (DefaultCommandHandler) RequiresPriveleges() bool {
	return false
}

// This will be called by a worker when a command arrives.
func GetCommandHandler(cmd string) (CommandHandler) {
	handler, ok := handlers[cmd]
	if !ok {
		return InvalidCommandHandler{}
	}
	return handler
}

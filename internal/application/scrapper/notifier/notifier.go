package notifier

import (
	"fmt"

	"log/slog"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain/models"
	"github.com/es-debug/backend-academy-2024-go-template/pkg"
	"github.com/go-co-op/gocron/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Notifier handles notification delivery to users about updates in their tracked resources.
type Notifier struct {
	repository LinksService
	external   ExternalClient
	sender     Sender
	sch        gocron.Scheduler
	guard      chan struct{}
}

// TODO: update external client interface in order to fetch only the newest activity (accept timestamp).
type ExternalClient interface {
	RetrieveStackOverflowUpdates(link string) ([]models.StackOverflowUpdate, error)
	RetrieveGitHubUpdates(link string) ([]models.GitHubUpdate, error)
}

type LinksService interface {
	GetChatLinks(chatID int64, includeAll bool) ([]models.Link, error)
	UpdateLinkActivity(linkID int64, status bool) error
	GetChatsIDs() ([]int64, error)
}

type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}

// New instantiates a new Notifier entity.
func New(repository LinksService, external ExternalClient, sender Sender, sch gocron.Scheduler) *Notifier {
	return &Notifier{
		repository: repository,
		external:   external,
		sender:     sender,
		sch:        sch,
		guard:      make(chan struct{}, runtime.GOMAXPROCS(0)),
	}
}

func (n *Notifier) Run() error {
	_, err := n.sch.NewJob(
		gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(10, 0, 0),
			),
		),
		gocron.NewTask(
			func() {
				n.PushUpdates()
			},
		),
	)

	if err != nil {
		return err
	}

	n.sch.Start()

	return nil
}

func (n *Notifier) PushUpdates() {
	chatIDs, err := n.repository.GetChatsIDs()
	if err != nil {
		slog.Error(
			"Notifier: failed to get chat IDs",
			slog.String("msg", err.Error()),
		)
	}

	wg := sync.WaitGroup{}

	for _, chatID := range chatIDs {
		wg.Add(1)
		n.guard <- struct{}{}

		go func(id int64) {
			defer func() {
				<-n.guard
				wg.Done()
			}()

			if err := n.processChatID(id); err != nil {
				slog.Error("Processing failed",
					slog.Int64("chat_id", id),
					slog.String("error", err.Error()))
			}
		}(chatID)
	}

	wg.Wait()
}

// ProcessChatID handles notifications for a single chat:
// 1. Retrieves active tracked links
// 2. Generates update messages
// 3. Sends notifications
// 4. Marks links as processed
// Returns error if fatal processing failure occurs.
func (n *Notifier) processChatID(chatID int64) error {
	links, err := n.repository.GetChatLinks(chatID, false)
	if err != nil {
		return fmt.Errorf("failed to get chat links: %w", err)
	}

	for _, link := range links {
		if link.Url == nil {
			slog.Error("notifier: link's url field is missing")
			continue
		}

		msg, err := n.serializeMessage(*link.Url)
		if err != nil {
			slog.Error(
				"notifier: failed to serialize the message",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if _, err := n.sender.Send(tgbotapi.NewMessage(chatID, msg)); err != nil {
			slog.Error(
				"Notifier: failed to send message to telegram",
				slog.String("msg", err.Error()),
			)

			continue
		}

		if link.Id == nil {
			slog.Error("notifier: link's id field is missing")
			continue
		}

		if err := n.repository.UpdateLinkActivity(*link.Id, false); err != nil {
			slog.Error(
				"notifier: failed to update link activity",
				slog.String("msg", err.Error()),
			)

			continue
		}
	}

	return nil
}

// serializeMessage serializes telegram message for certain service.
func (n *Notifier) serializeMessage(url string) (string, error) {
	var (
		msg string
		err error
	)

	service, err := pkg.ServiceFromURL(url)
	if err != nil {
		return "", err
	}

	switch service {
	case "github":
		msg, err = n.serializeMessageGitHub(url)
	case "stackoverflow":
		msg, err = n.serializeMessageStackOverflow(url)
	default:
		return "", fmt.Errorf("unsupported service: %s", service)
	}

	return msg, err
}

// serializeMessageGitHub formats retrieved GitHub update in an appropriate way.
func (n *Notifier) serializeMessageGitHub(url string) (string, error) {
	updates, err := n.external.RetrieveGitHubUpdates(url)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve github updates: %w", err)
	}

	if len(updates) == 0 {
		return "", ErrEmptyUpdates
	}

	var b strings.Builder

	for _, update := range updates {
		fmt.Fprintf(&b, "📌 New %s\n", update.Title)
		fmt.Fprintf(&b, "👤 Author: %s\n", update.User.Login)
		fmt.Fprintf(&b, "🕒 Date: %s\n", update.CreatedAt)
		fmt.Fprintf(&b, "🔗 View: %s\n\n", update.Body)
	}

	return b.String(), nil
}

// serializeMessageStackOverflow formats retrieved StackOverflow update in an appropriate way.
func (n *Notifier) serializeMessageStackOverflow(url string) (string, error) {
	updates, err := n.external.RetrieveStackOverflowUpdates(url)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve stackoverflow updates: %w", err)
	}

	if len(updates) == 0 {
		return "", ErrEmptyUpdates
	}

	var b strings.Builder

	for _, update := range updates {
		fmt.Fprintf(&b, "📌 New %s\n", update.Type)
		fmt.Fprintf(&b, "👤 Author: %s\n", update.Owner.Username)
		fmt.Fprintf(&b, "🕒 Date: %s\n", time.Unix(update.CreatedAt, 0).Format(time.RFC1123))
		fmt.Fprintf(&b, "🔗 View: %s\n\n", update.Body)
	}

	return b.String(), nil
}

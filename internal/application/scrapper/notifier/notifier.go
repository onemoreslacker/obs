package notifier

type Notifier struct {
}

type Updater interface {
	RetrieveUpdates()
}

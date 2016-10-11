package types

type NotificationType uint8

const (
	ChangedNotification = iota
	AddedNotification
	RemovedNotification
)

type Notification struct {
	Type NotificationType
	Item Item
}

type Binding interface {
	GetValue(string) Item
	Notifications() <-chan Notification
}

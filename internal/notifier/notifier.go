package notifier

import (
	"fmt"

	"github.com/gen2brain/beeep"
)

// Notifier handles desktop notifications
type Notifier struct {
	appName string
}

// New creates a new notifier
func New(appName string) *Notifier {
	return &Notifier{
		appName: appName,
	}
}

// NotifyDrafts sends a notification about the number of drafts
func (n *Notifier) NotifyDrafts(count int) error {
	title := n.appName
	message := fmt.Sprintf("You have %d draft(s) in your Gmail", count)

	if count == 0 {
		message = "No drafts in your Gmail"
	} else if count == 1 {
		message = "You have 1 draft in your Gmail"
	}

	return beeep.Notify(title, message, "")
}

// NotifyDraftsWithDetails sends a notification with draft details
func (n *Notifier) NotifyDraftsWithDetails(count int, emptyCount int) error {
	title := n.appName
	message := fmt.Sprintf("You have %d draft(s) in your Gmail", count)

	if emptyCount > 0 {
		message += fmt.Sprintf(" (%d empty)", emptyCount)
	}

	return beeep.Notify(title, message, "")
}

// NotifyCleanup sends a notification about deleted empty drafts
func (n *Notifier) NotifyCleanup(deletedCount int) error {
	if deletedCount == 0 {
		return nil
	}

	title := n.appName
	message := fmt.Sprintf("Deleted %d old empty draft(s)", deletedCount)

	return beeep.Notify(title, message, "")
}

// NotifyError sends an error notification
func (n *Notifier) NotifyError(err error) error {
	title := fmt.Sprintf("%s - Error", n.appName)
	message := fmt.Sprintf("Error: %v", err)

	return beeep.Notify(title, message, "")
}

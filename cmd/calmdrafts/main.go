package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"calmdrafts/internal/config"
	"calmdrafts/internal/gmail"
	"calmdrafts/internal/notifier"
)

const appName = "CalmDrafts"

func main() {
	configPath := flag.String("config", "config.json", "Path to configuration file")
	checkNow := flag.Bool("check", false, "Run a single check and exit")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create notifier
	notif := notifier.New(appName)

	// Create Gmail client
	ctx := context.Background()
	client, err := gmail.NewClient(ctx, cfg.CredentialsPath, cfg.TokenPath)
	if err != nil {
		log.Fatalf("Error creating Gmail client: %v", err)
		notif.NotifyError(err)
		os.Exit(1)
	}

	fmt.Printf("%s started. Checking drafts every %v\n", appName, cfg.CheckInterval)

	if *checkNow {
		// Run a single check and exit
		if err := checkAndCleanDrafts(ctx, client, notif, cfg); err != nil {
			log.Printf("Error during check: %v", err)
			os.Exit(1)
		}
		return
	}

	// Set up periodic checking
	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run initial check
	if err := checkAndCleanDrafts(ctx, client, notif, cfg); err != nil {
		log.Printf("Error during initial check: %v", err)
	}

	// Main loop
	for {
		select {
		case <-ticker.C:
			if err := checkAndCleanDrafts(ctx, client, notif, cfg); err != nil {
				log.Printf("Error during check: %v", err)
			}
		case sig := <-sigChan:
			fmt.Printf("\nReceived signal %v, shutting down gracefully...\n", sig)
			return
		}
	}
}

// checkAndCleanDrafts performs a full check: lists drafts, notifies user, and cleans up old empty drafts
func checkAndCleanDrafts(ctx context.Context, client *gmail.Client, notif *notifier.Notifier, cfg *config.Config) error {
	fmt.Printf("[%s] Checking drafts...\n", time.Now().Format("2006-01-02 15:04:05"))

	// List all drafts
	drafts, err := client.ListDrafts(ctx)
	if err != nil {
		notif.NotifyError(err)
		return fmt.Errorf("error listing drafts: %v", err)
	}

	// Count empty drafts
	emptyCount := 0
	for _, draft := range drafts {
		if draft.IsEmpty {
			emptyCount++
		}
	}

	fmt.Printf("Found %d draft(s) (%d empty)\n", len(drafts), emptyCount)

	// Notify user about drafts
	if err := notif.NotifyDraftsWithDetails(len(drafts), emptyCount); err != nil {
		log.Printf("Error sending notification: %v", err)
	}

	// Clean up old empty drafts
	deletedCount := 0
	cutoffTime := time.Now().Add(-cfg.CleanupAge)

	for _, draft := range drafts {
		if draft.IsEmpty && draft.InternalDate.Before(cutoffTime) {
			age := time.Since(draft.InternalDate)
			fmt.Printf("Deleting empty draft (ID: %s, age: %v)\n", draft.ID, age.Round(time.Hour))

			if err := client.DeleteDraft(ctx, draft.ID); err != nil {
				log.Printf("Error deleting draft %s: %v", draft.ID, err)
				continue
			}
			deletedCount++
		}
	}

	if deletedCount > 0 {
		fmt.Printf("Deleted %d old empty draft(s)\n", deletedCount)
		if err := notif.NotifyCleanup(deletedCount); err != nil {
			log.Printf("Error sending cleanup notification: %v", err)
		}
	}

	return nil
}

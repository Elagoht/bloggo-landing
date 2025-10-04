package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"bloggo-landing/internal/cache"
)

type Event struct {
	Event     string                 `json:"event"`
	Entity    string                 `json:"entity"`
	ID        *int                   `json:"id"`
	Slug      *string                `json:"slug"`
	OldSlug   *string                `json:"oldSlug"`
	Action    string                 `json:"action"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type Handler struct {
	cacheManager *cache.Manager
	secret       string
}

func New(cacheManager *cache.Manager, secret string) *Handler {
	return &Handler{
		cacheManager: cacheManager,
		secret:       secret,
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if h.secret != "" {
		signature := r.Header.Get("X-Webhook-Signature")
		if !h.verifySignature(body, signature) {
			log.Printf("Invalid webhook signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var event Event
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to unmarshal webhook event: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.processEvent(&event); err != nil {
		log.Printf("Failed to process webhook event: %v", err)
		http.Error(w, "Failed to process event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) processEvent(event *Event) error {
	log.Printf("Processing webhook event: event=%s, entity=%s, action=%s", event.Event, event.Entity, event.Action)

	switch event.Event {
	case "cms.sync":
		log.Println("CMS sync triggered - rebuilding all pages")
		return h.cacheManager.RebuildAll()

	case "post.created":
		if event.Slug == nil {
			return fmt.Errorf("missing slug in event")
		}
		return h.cacheManager.UpdatePost(*event.Slug)

	case "post.updated":
		if event.Slug == nil {
			return fmt.Errorf("missing slug in event")
		}
		// Delete old slug's prerendered file if slug changed
		if event.OldSlug != nil && *event.OldSlug != *event.Slug {
			log.Printf("Slug changed from %s to %s - deleting old prerendered file", *event.OldSlug, *event.Slug)
			if err := h.cacheManager.DeletePost(*event.OldSlug); err != nil {
				log.Printf("Failed to delete old slug %s: %v", *event.OldSlug, err)
			}
		}
		return h.cacheManager.UpdatePost(*event.Slug)

	case "post.deleted":
		if event.Slug == nil {
			return fmt.Errorf("missing slug in event")
		}
		return h.cacheManager.DeletePost(*event.Slug)

	case "category.created", "category.updated", "category.deleted",
		"tag.created", "tag.updated", "tag.deleted":
		log.Println("Category/tag changed - rebuilding all pages")
		return h.cacheManager.RebuildAll()

	default:
		log.Printf("Unknown event type: %s - ignoring", event.Event)
		return nil
	}
}

func (h *Handler) verifySignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Bloggo sends the raw secret as the signature
	return signature == h.secret
}

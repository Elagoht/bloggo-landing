package main

import (
	"log"

	"bloggo-landing/config"
	"bloggo-landing/internal/cache"
	"bloggo-landing/internal/fetcher"
	"bloggo-landing/internal/prerender"
	"bloggo-landing/internal/server"
	"bloggo-landing/internal/webhook"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	log.Println("Initializing bloggo landing page server...")

	bloggoFetcher := fetcher.New(cfg.BloggoAPIURL, cfg.BloggoAPIToken)

	renderer, err := prerender.New(
		bloggoFetcher,
		cfg.TemplatesDir,
		cfg.PrerenderedDir,
	)
	if err != nil {
		log.Fatalf("Failed to initialize renderer: %v", err)
	}

	log.Println("Prerendering all pages...")
	if err := renderer.RenderAll(); err != nil {
		log.Printf("Warning: Failed to prerender pages: %v", err)
		log.Println("Server will start but some pages may not be available")
	} else {
		log.Println("Prerendering completed successfully")
	}

	cacheManager := cache.New(renderer)

	webhookHandler := webhook.New(cacheManager, cfg.WebhookSecret)

	srv := server.New(
		cfg.ServerPort,
		cfg.PrerenderedDir,
		cfg.StaticDir,
		webhookHandler,
	)

	log.Printf("Server starting on http://localhost:%s", cfg.ServerPort)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

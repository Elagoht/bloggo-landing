package cache

import (
	"fmt"
	"log"

	"bloggo-landing/internal/prerender"
)

type Manager struct {
	renderer *prerender.Renderer
}

func New(renderer *prerender.Renderer) *Manager {
	return &Manager{
		renderer: renderer,
	}
}

func (m *Manager) InvalidatePost(slug string) error {
	log.Printf("Invalidating post: %s", slug)

	if err := m.renderer.DeletePost(slug); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

func (m *Manager) UpdatePost(slug string) error {
	log.Printf("Updating document: %s", slug)

	if err := m.renderer.RenderPostBySlug(slug); err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	if err := m.renderer.RenderDocuments(); err != nil {
		return fmt.Errorf("failed to update documents index: %w", err)
	}

	return nil
}

func (m *Manager) CreatePost(slug string) error {
	log.Printf("Creating document: %s", slug)

	return m.UpdatePost(slug)
}

func (m *Manager) DeletePost(slug string) error {
	log.Printf("Deleting document: %s", slug)

	if err := m.renderer.DeletePost(slug); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	if err := m.renderer.RenderDocuments(); err != nil {
		return fmt.Errorf("failed to update documents index: %w", err)
	}

	return nil
}

func (m *Manager) RebuildAll() error {
	log.Printf("Rebuilding all pages")

	return m.renderer.RenderAll()
}

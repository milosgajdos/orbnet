package stars

import (
	"context"
	"strings"
	"sync"

	"github.com/google/go-github/v61/github"
	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
	"github.com/milosgajdos/orbnet/pkg/graph/style"
)

const (
	// DefaultWeight is default edge weight.
	DefaultWeight = 1.0
)

// Stars builds GitHub stars graph.
type Stars struct {
	g     graph.Adder
	nodes map[string]*memory.Node
	mu    *sync.RWMutex
}

// NewBuilder creates a new GH stars graph builder and returns it.
func NewBuilder(g graph.Adder) (*Stars, error) {
	return &Stars{
		g:     g,
		nodes: make(map[string]*memory.Node),
		mu:    &sync.RWMutex{},
	}, nil
}

func (s *Stars) addNode(uid, label string, attrs map[string]interface{}, style style.Style) (*memory.Node, error) {
	attrs["style"] = style.Type
	attrs["shape"] = style.Shape
	attrs["color"] = style.Color

	opts := []memory.Option{
		memory.WithUID(uid),
		memory.WithLabel(label),
		memory.WithAttrs(attrs),
		memory.WithStyle(style),
	}

	node := s.g.NewNode()

	n, err := memory.NewNode(node.ID(), opts...)
	if err != nil {
		return nil, err
	}
	s.g.AddNode(n)

	return n, nil
}

func (s *Stars) linkNodes(from, to *memory.Node, label string, attrs map[string]interface{}, style style.Style) (*memory.Edge, error) {
	attrs["style"] = style.Type
	attrs["shape"] = style.Shape
	attrs["color"] = style.Color

	opts := []memory.Option{
		memory.WithLabel(label),
		memory.WithAttrs(attrs),
		memory.WithStyle(style),
	}

	e, err := memory.NewEdge(from, to, opts...)
	if err != nil {
		return nil, err
	}
	s.g.SetWeightedEdge(e)

	return e, nil
}

func (s *Stars) update(repos []*github.StarredRepository) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, repo := range repos {
		uid := *repo.Repository.Owner.NodeID
		ownerNode, ok := s.nodes[uid]
		if !ok {
			style := OwnerEntity.DefaultStyle()
			label := OwnerEntity.String()
			attrs, err := OwnerAttrs(repo.Repository.Owner)
			if err != nil {
				return err
			}
			ownerNode, err = s.addNode(uid, label, attrs, style)
			if err != nil {
				return err
			}
			s.nodes[uid] = ownerNode
		}

		uid = *repo.Repository.NodeID
		repoNode, ok := s.nodes[uid]
		if !ok {
			style := RepoEntity.DefaultStyle()
			label := RepoEntity.String()
			attrs, err := RepoAttrs(repo.Repository, repo.StarredAt)
			if err != nil {
				return err
			}
			repoNode, err = s.addNode(uid, label, attrs, style)
			if err != nil {
				return err
			}
			s.nodes[uid] = repoNode
		}

		if e := s.g.Edge(repoNode.ID(), ownerNode.ID()); e == nil {
			style := LinkEntity.DefaultStyle()
			rel := OwnedByEdgeLabel
			attrs := LinkAttrs(rel, DefaultWeight)
			if _, err := s.linkNodes(repoNode, ownerNode, rel, attrs, style); err != nil {
				return err
			}
		}

		for _, topic := range repo.Repository.Topics {
			uid = strings.ToLower(topic) + "-" + TopicEntity.String()
			topicNode, ok := s.nodes[uid]
			if !ok {
				style := TopicEntity.DefaultStyle()
				label := TopicEntity.String()
				attrs := TopicAttrs(strings.ToLower(topic))
				topicNode, err = s.addNode(uid, label, attrs, style)
				if err != nil {
					return err
				}
				s.nodes[uid] = topicNode
			}

			if e := s.g.Edge(repoNode.ID(), topicNode.ID()); e == nil {
				style := LinkEntity.DefaultStyle()
				rel := HasTopicEdgeLabel
				attrs := LinkAttrs(rel, DefaultWeight)
				if _, err := s.linkNodes(repoNode, topicNode, rel, attrs, style); err != nil {
					return err
				}
			}
		}

		if repo.Repository.Language != nil {
			uid = strings.ToLower(*repo.Repository.Language) + "-" + LangEntity.String()
			langNode, ok := s.nodes[uid]
			if !ok {
				style := LangEntity.DefaultStyle()
				label := LangEntity.String()
				attrs := LangAttrs(strings.ToLower(*repo.Repository.Language))
				langNode, err = s.addNode(uid, label, attrs, style)
				if err != nil {
					return err
				}
				s.nodes[uid] = langNode
			}

			if e := s.g.Edge(repoNode.ID(), langNode.ID()); e == nil {
				style := LinkEntity.DefaultStyle()
				rel := IsLangEdgeLabel
				attrs := LinkAttrs(rel, DefaultWeight)
				if _, err := s.linkNodes(repoNode, langNode, rel, attrs, style); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Build builds a graph by adding nodes and edges from the data received on repos channel.
func (s *Stars) Build(ctx context.Context, reposChan <-chan interface{}) error {
	for {
		select {
		case data := <-reposChan:
			// reposChan has been closed
			if data == nil {
				return nil
			}
			if repos, ok := data.([]*github.StarredRepository); ok {
				if err := s.update(repos); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

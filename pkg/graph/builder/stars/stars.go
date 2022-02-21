package stars

import (
	"context"
	"strings"
	"sync"

	"github.com/google/go-github/v34/github"
	"github.com/milosgajdos/orbnet/pkg/graph"
	"github.com/milosgajdos/orbnet/pkg/graph/memory"
	"github.com/milosgajdos/orbnet/pkg/graph/style"
)

const (
	// DefaultWeight is default edge weight.
	DefaultWeight = 1.0
)

// Stars is a stars builder.
type Stars struct {
	g     graph.Adder
	nodes map[string]*memory.Node
	mu    *sync.RWMutex
}

// NewBuilder creates a new builder and returns it.
func NewBuilder(g graph.Adder) (*Stars, error) {
	return &Stars{
		g:     g,
		nodes: make(map[string]*memory.Node),
		mu:    &sync.RWMutex{},
	}, nil
}

func (s *Stars) addNode(uid, name, fullName, label string, style style.Style) (*memory.Node, error) {
	attrs := map[string]interface{}{
		"name":      name,
		"full_name": fullName,
		"style":     style.Type,
		"shape":     style.Shape,
		"color":     style.Color,
	}

	opts := []memory.Option{
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

func (s *Stars) linkNodes(from, to *memory.Node, relation, label string, weight float64, style style.Style) (*memory.Edge, error) {
	attrs := map[string]interface{}{
		"relation": relation,
		"style":    style.Type,
		"shape":    style.Shape,
		"color":    style.Color,
		"weight":   weight,
	}

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
		owner := *repo.Repository.Owner.Login
		if repo.Repository.Organization != nil {
			owner = *repo.Repository.Organization.Login
		}

		uid := owner + "-" + Owner.String()
		ownerNode, ok := s.nodes[uid]
		if !ok {
			style := Owner.DefaultStyle()
			label := Owner.String()
			ownerNode, err = s.addNode(uid, owner, owner, label, style)
			if err != nil {
				return err
			}
			s.nodes[uid] = ownerNode
		}

		uid = *repo.Repository.NodeID
		repoNode, ok := s.nodes[uid]
		if !ok {
			style := Repo.DefaultStyle()
			label := Repo.String()
			repoNode, err = s.addNode(uid, *repo.Repository.Name, *repo.Repository.FullName, label, style)
			if err != nil {
				return err
			}
			s.nodes[uid] = repoNode
		}

		if e := s.g.Edge(repoNode.ID(), ownerNode.ID()); e == nil {
			style := Link.DefaultStyle()
			relation := OwnsEdgeLabel
			if _, err := s.linkNodes(repoNode, ownerNode, relation, relation, DefaultWeight, style); err != nil {
				return err
			}
		}

		for _, topic := range repo.Repository.Topics {
			uid = strings.ToLower(topic) + "-" + Topic.String()
			topicNode, ok := s.nodes[uid]
			if !ok {
				style := Topic.DefaultStyle()
				label := Topic.String()
				name := strings.ToLower(topic)
				topicNode, err = s.addNode(uid, name, name, label, style)
				if err != nil {
					return err
				}
				s.nodes[uid] = topicNode
			}

			if e := s.g.Edge(repoNode.ID(), topicNode.ID()); e == nil {
				style := Link.DefaultStyle()
				relation := TopicEdgeLabel
				if _, err := s.linkNodes(repoNode, topicNode, relation, relation, DefaultWeight, style); err != nil {
					return err
				}
			}
		}

		if repo.Repository.Language != nil {
			uid = strings.ToLower(*repo.Repository.Language) + "-" + Lang.String()
			langNode, ok := s.nodes[uid]
			if !ok {
				style := Lang.DefaultStyle()
				label := Lang.String()
				name := strings.ToLower(*repo.Repository.Language)
				langNode, err = s.addNode(uid, name, name, label, style)
				if err != nil {
					return err
				}
				s.nodes[uid] = langNode
			}

			if e := s.g.Edge(repoNode.ID(), langNode.ID()); e == nil {
				style := Link.DefaultStyle()
				relation := LangEdgeLabel
				if _, err := s.linkNodes(repoNode, langNode, relation, relation, DefaultWeight, style); err != nil {
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

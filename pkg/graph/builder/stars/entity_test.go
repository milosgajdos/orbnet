package stars

import (
	"reflect"
	"testing"

	"github.com/milosgajdos/orbnet/pkg/graph/style"
)

func TestEntityString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Entity   Entity
		Expected string
	}{
		{OwnerEntity, ownerString},
		{RepoEntity, repoString},
		{TopicEntity, topicString},
		{LangEntity, langString},
		{LinkEntity, linkString},
		{-100, unknownString},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("EntityString", func(t *testing.T) {
			t.Parallel()
			if entStr := tc.Entity.String(); entStr != tc.Expected {
				t.Errorf("expected: %s, got: %s", tc.Expected, entStr)
			}
		})
	}
}

func TestEntityStyle(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Entity   Entity
		Expected style.Style
	}{
		{OwnerEntity, style.Style{Type: DefaultStyleType, Shape: OwnerShape, Color: OwnerColor}},
		{RepoEntity, style.Style{Type: DefaultStyleType, Shape: RepoShape, Color: RepoColor}},
		{TopicEntity, style.Style{Type: DefaultStyleType, Shape: TopicShape, Color: TopicColor}},
		{LangEntity, style.Style{Type: DefaultStyleType, Shape: LangShape, Color: LangColor}},
		{LinkEntity, style.Style{Type: DefaultStyleType, Shape: LinkShape, Color: LinkColor}},
		{-100, style.Style{Type: DefaultStyleType, Shape: UnknownShape, Color: UnknownColor}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("EntityStyle", func(t *testing.T) {
			t.Parallel()
			if s := tc.Entity.DefaultStyle(); !reflect.DeepEqual(s, tc.Expected) {
				t.Errorf("expected: %v, got: %v", tc.Expected, s)
			}
		})
	}
}

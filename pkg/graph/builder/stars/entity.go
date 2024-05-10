package stars

import "github.com/milosgajdos/orbnet/pkg/graph/style"

// Entity is a GitHub stars graph entity
type Entity int

const (
	OwnerEntity Entity = iota
	RepoEntity
	TopicEntity
	LangEntity
	LinkEntity
)

const (
	ownerString   = "Owner"
	repoString    = "Repo"
	topicString   = "Topic"
	langString    = "Lang"
	linkString    = "Link"
	unknownString = "Unknown"
)

// String implements fmt.Stringer
func (e Entity) String() string {
	switch e {
	case OwnerEntity:
		return ownerString
	case RepoEntity:
		return repoString
	case TopicEntity:
		return topicString
	case LangEntity:
		return langString
	case LinkEntity:
		return linkString
	default:
		return unknownString
	}
}

// DefaultStyle returns default style.Style.
func (e Entity) DefaultStyle() style.Style {
	switch e {
	case OwnerEntity:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: OwnerShape,
			Color: OwnerColor,
		}
	case RepoEntity:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: RepoShape,
			Color: RepoColor,
		}
	case TopicEntity:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: TopicShape,
			Color: TopicColor,
		}
	case LangEntity:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: LangShape,
			Color: LangColor,
		}
	case LinkEntity:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: LinkShape,
			Color: LinkColor,
		}
	default:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: UnknownShape,
			Color: UnknownColor,
		}
	}
}

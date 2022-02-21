package stars

import "github.com/milosgajdos/orbnet/pkg/graph/style"

// Entity is a GitHub stars graph entity
type Entity int

const (
	Owner Entity = iota
	Repo
	Topic
	Lang
	Link
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
	case Owner:
		return ownerString
	case Repo:
		return repoString
	case Topic:
		return topicString
	case Lang:
		return langString
	case Link:
		return linkString
	default:
		return unknownString
	}
}

// DefaultStyle returns default style.Style.
func (e Entity) DefaultStyle() style.Style {
	switch e {
	case Owner:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: OwnerShape,
			Color: OwnerColor,
		}
	case Repo:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: RepoShape,
			Color: RepoColor,
		}
	case Topic:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: TopicShape,
			Color: TopicColor,
		}
	case Lang:
		return style.Style{
			Type:  DefaultStyleType,
			Shape: LangShape,
			Color: LangColor,
		}
	case Link:
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

package customtags

import (
	"github.com/awslabs/goformation/v4/cloudformation/tags"
)

func New() []tags.Tag {

	return []tags.Tag{
		{
			Key:   "Owned",
			Value: "Hugo",
		},
		{
			Key:   "Cloudformation",
			Value: "Hugo",
		},
	}
}

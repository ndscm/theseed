package aips

import (
	"net/url"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// ResourceName is the parsed form of an AIP-122 resource name such as
// "people/123/brainSteps/456". A resource name is a sequence of
// collection-id/resource-id segment pairs; Ids holds those pairs keyed by
// collection id (e.g. {"people": "123", "brainSteps": "456"}), so the parent
// and the leaf resource id can be looked up by their collection. Url is the
// parsed URL the name was read from, retaining any query string or other URL
// components.
type ResourceName struct {
	Url *url.URL
	Ids map[string]string
}

func ParseResourceName(resourceName string) (*ResourceName, error) {
	resourceName = strings.TrimSpace(resourceName)
	resourceUrl, err := url.Parse(resourceName)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	path := resourceUrl.Opaque
	if path == "" {
		path = resourceUrl.EscapedPath()
	}

	ids := map[string]string{}
	path = strings.Trim(path, "/")
	if path != "" {
		parts := strings.Split(path, "/")
		if len(parts)%2 != 0 {
			return nil, seederr.WrapErrorf("invalid resource name %q: expected collection/id pairs", resourceName)
		}
		for i := 0; i < len(parts); i += 2 {
			collection, err := url.PathUnescape(parts[i])
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			id, err := url.PathUnescape(parts[i+1])
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			if collection == "" || id == "" {
				return nil, seederr.WrapErrorf("invalid resource name %q: empty collection or id segment", resourceName)
			}
			ids[collection] = id
		}
	}
	result := &ResourceName{
		Url: resourceUrl,
		Ids: ids,
	}
	return result, nil
}

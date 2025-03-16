package entities

// Link represents tracked resource link.
type Link struct {
	Filters *[]string
	Id      *int64
	Tags    *[]string
	Url     *string
}

// NewLink instantiates a new link entity.
func NewLink(id int64, url string, tags, filters []string) Link {
	return Link{
		Id:      &id,
		Url:     &url,
		Tags:    &tags,
		Filters: &filters,
	}
}

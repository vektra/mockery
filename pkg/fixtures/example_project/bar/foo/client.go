package foo

type Client interface {
	Search(query string) ([]string, error)
}

package repository

type NewsletterRepository interface {
	ListNewsletters() ([]string, error)
	CreateNewsletter(name string) error
}

type inMemoryNewsletterRepo struct {
	newsletters []string
}

func NewInMemoryNewsletterRepo() NewsletterRepository {
	return &inMemoryNewsletterRepo{newsletters: []string{"Sample Newsletter"}}
}

func (r *inMemoryNewsletterRepo) ListNewsletters() ([]string, error) {
	return r.newsletters, nil
}

func (r *inMemoryNewsletterRepo) CreateNewsletter(name string) error {
	r.newsletters = append(r.newsletters, name)
	return nil
}

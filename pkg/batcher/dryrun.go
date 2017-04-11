package batcher

type dryrun struct {
	contents chan interface{}
}

// NewDryrun creates a dryrun batcher.
// Used for debugging.
func NewDryrun(capacity int) Batcher {
	return &dryrun{contents: make(chan interface{}, capacity)}
}

func (d *dryrun) Close() {
	close(d.contents)
}

func (d *dryrun) Put(item interface{}) error {
	d.contents <- item
	return nil
}

func (d *dryrun) Get() (interface{}, error) {
	item := <-d.contents
	return item, nil
}

func (d *dryrun) Succeed() {}

func (d *dryrun) Fail() (rewind bool) {
	return
}

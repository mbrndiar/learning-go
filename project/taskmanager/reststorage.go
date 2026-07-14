package taskmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/mbrndiar/learning-go/project/taskclient"
)

// RESTStorage adapts a taskclient.Client to the Storage interface. It converts
// between the domain Task and the client's wire Task and translates the
// client's not-found sentinel into ErrTaskNotFound so callers see a uniform
// error regardless of backend.
type RESTStorage struct {
	client *taskclient.Client
}

// NewRESTStorage returns a RESTStorage backed by the given client.
func NewRESTStorage(client *taskclient.Client) (*RESTStorage, error) {
	if client == nil {
		return nil, errors.New("taskmanager: rest client must not be nil")
	}
	return &RESTStorage{client: client}, nil
}

// List returns every remote task.
func (s *RESTStorage) List(ctx context.Context) ([]Task, error) {
	remote, err := s.client.List(ctx)
	if err != nil {
		return nil, translateClientError(err)
	}
	tasks := make([]Task, len(remote))
	for i, task := range remote {
		tasks[i] = fromClientTask(task)
	}
	return tasks, nil
}

// Get returns a single remote task by identifier.
func (s *RESTStorage) Get(ctx context.Context, id int) (Task, error) {
	task, err := s.client.Get(ctx, id)
	if err != nil {
		return Task{}, translateClientError(err)
	}
	return fromClientTask(task), nil
}

// Add creates a remote task.
func (s *RESTStorage) Add(ctx context.Context, title string) (Task, error) {
	task, err := s.client.Add(ctx, title)
	if err != nil {
		return Task{}, translateClientError(err)
	}
	return fromClientTask(task), nil
}

// Complete marks a remote task as done.
func (s *RESTStorage) Complete(ctx context.Context, id int) (Task, error) {
	task, err := s.client.Complete(ctx, id)
	if err != nil {
		return Task{}, translateClientError(err)
	}
	return fromClientTask(task), nil
}

// Remove deletes a remote task.
func (s *RESTStorage) Remove(ctx context.Context, id int) error {
	if err := s.client.Remove(ctx, id); err != nil {
		return translateClientError(err)
	}
	return nil
}

// fromClientTask converts a wire task into the domain model.
func fromClientTask(task taskclient.Task) Task {
	return Task{ID: task.ID, Title: task.Title, Done: task.Done}
}

// translateClientError maps client-specific errors onto the storage contract
// while preserving the underlying error for callers that inspect it further.
func translateClientError(err error) error {
	if errors.Is(err, taskclient.ErrNotFound) {
		return fmt.Errorf("%w: %w", ErrTaskNotFound, err)
	}
	return err
}

package state

import (
	"errors"
	"sync"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/zmb3/spotify/v2"
)

type QueueUpdatedMsg struct{}

func UpdateQueue() tea.Cmd {
	return func() tea.Msg {
		return QueueUpdatedMsg{}
	}
}

type QueueManager struct {
	tracks []spotify.SimpleTrack
	mutex  sync.RWMutex
}

func (q *QueueManager) Enqueue(track spotify.SimpleTrack) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.tracks = append(q.tracks, track)
}

func (q *QueueManager) Dequeue() (spotify.SimpleTrack, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.tracks) == 0 {
		return spotify.SimpleTrack{}, errors.New("queue is empty")
	}

	track := q.tracks[0]
	q.tracks = q.tracks[1:]
	return track, nil
}

func (q *QueueManager) Peek() (spotify.SimpleTrack, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if len(q.tracks) == 0 {
		return spotify.SimpleTrack{}, errors.New("queue is empty")
	}

	return q.tracks[0], nil
}

func (q *QueueManager) List() []spotify.SimpleTrack {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	result := make([]spotify.SimpleTrack, len(q.tracks))
	copy(result, q.tracks)
	return result
}

func (q *QueueManager) Size() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.tracks)
}

func (q *QueueManager) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.tracks) == 0
}

func (q *QueueManager) PopAt(index int) (spotify.SimpleTrack, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if index < 0 || index >= len(q.tracks) {
		return spotify.SimpleTrack{}, errors.New("index out of range")
	}

	track := q.tracks[index]

	q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)

	return track, nil
}

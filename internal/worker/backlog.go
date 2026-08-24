package worker

import (
	"sort"
	"ticketdesk/internal/model"
)

type Backlog struct{ tasks []model.WorkerTask }

func NewBacklog(tasks []model.WorkerTask) *Backlog {
	return &Backlog{tasks: append([]model.WorkerTask(nil), tasks...)}
}

func (b *Backlog) Add(tasks ...model.WorkerTask) { b.tasks = append(b.tasks, tasks...) }

func (b *Backlog) Next() (model.WorkerTask, bool) {
	for index, task := range b.tasks {
		if task.Ready() {
			b.tasks = append(b.tasks[:index], b.tasks[index+1:]...)
			return task, true
		}
	}
	return model.WorkerTask{}, false
}

func (b *Backlog) Pending() int {
	count := 0
	for _, task := range b.tasks {
		if task.State != model.TaskDone {
			count++
		}
	}
	return count
}

func (b *Backlog) Ordered() []model.WorkerTask {
	result := append([]model.WorkerTask(nil), b.tasks...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (b *Backlog) ClearDone() {
	kept := b.tasks[:0]
	for _, task := range b.tasks {
		if task.State != model.TaskDone {
			kept = append(kept, task)
		}
	}
	b.tasks = kept
}

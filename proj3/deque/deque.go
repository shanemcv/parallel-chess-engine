package deque

import (
	"sync/atomic"
)

type Task func()

type Deque struct {
	capacity int
	tasks    []atomic.Pointer[Task]
	bottom   atomic.Uint64
	top      atomic.Uint64
}

func NewDeque(capacity int) *Deque {
	tasks := make([]atomic.Pointer[Task], capacity)

	return &Deque{capacity: capacity, tasks: tasks}
}

func (dq *Deque) PushBottom(task Task) {
	bottom := dq.bottom.Load()
	task_ptr := &task
	dq.tasks[bottom].Store(task_ptr)
	dq.bottom.Store(bottom + 1)
}

func (dq *Deque) PopTop() (Task, bool) {
	for {
		top := dq.top.Load()
		bottom := dq.bottom.Load()
		if top >= bottom {
			return nil, false
		}
		task_ptr := dq.tasks[top].Load()
		task := *task_ptr
		if dq.top.CompareAndSwap(top, top+1) {
			return task, true
		}
	}
}

func (dq *Deque) PopBottom() (Task, bool) {
	if dq.bottom.Load() == 0 {
		return nil, false
	}
	bottom := dq.bottom.Add(^uint64(0))

	top := dq.top.Load()

	if top > bottom {
		dq.bottom.Store(top)
		return nil, false
	}

	task_ptr := dq.tasks[bottom].Load()
	task := *task_ptr

	if top == bottom {
		if !dq.top.CompareAndSwap(top, top+1) {
			dq.bottom.Store(top + 1)
			return nil, false
		}
		dq.bottom.Store(top + 1)
	}

	return task, true

}

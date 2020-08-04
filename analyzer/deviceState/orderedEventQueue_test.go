package deviceState

import (
	"math/rand"
	"testing"
)

const randseed = 31415926

func pushEvent(t *testing.T, eq *OrderedEventQueue, id int64, c chan int64) error {
	err := eq.Push(id, func() {
		c <- id
	})
	if t != nil && err != nil {
		t.Errorf("eq.Push failed: %v", err)
	}
	return err
}

func TestOrderedEventQueue(t *testing.T) {
	rand.Seed(randseed)
	eq := &OrderedEventQueue{lastid: -1}
	var i int64
	c := make(chan int64)
	for i = 0; i < 20; i += 1 {
		go pushEvent(t, eq, i, c)
		j := <-c
		if j != i {
			t.Errorf("wrong event executed")
		}
	}
	for i = 21; i < 40; i += 1 {
		pushEvent(t, eq, i, c)
	}
	select {
	case <-c:
		t.Errorf("out of order event should not be executed")
	default:
	}

	go pushEvent(t, eq, 20, c)

	for i = 20; i < 40; i += 1 {
		j := <-c
		if j != i {
			t.Errorf("wrong event executed")
		}
	}

	select {
	case <-c:
		t.Errorf("queue is supposed to be empty")
	default:
	}
	if pushEvent(nil, eq, 10, c) == nil {
		t.Errorf("should have rejected past eid")
	}
	select {
	case <-c:
		t.Errorf("queue is supposed to be empty")
	default:
	}

	nums := make([]int64, 60)
	for i = 0; i < 60; i++ {
		nums[i] = i + 41
	}
	rand.Shuffle(60, func(i, j int) { nums[i], nums[j] = nums[j], nums[i] })
	for i = 0; i < 60; i++ {
		pushEvent(t, eq, nums[i], c)
	}
	select {
	case <-c:
		t.Errorf("out of order event should not be executed")
	default:
	}
	go pushEvent(t, eq, 40, c)

	for i = 40; i <= 100; i += 1 {
		j := <-c
		if j != i {
			t.Errorf("wrong event executed")
		}
	}
}

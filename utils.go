package main

import (
	"container/heap"
	"encoding/csv"
	"log"
	"os"
)

// SET

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](values ...T) Set[T] {
	s := Set[T]{}
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

func (s Set[T]) Add(values ...T) {
	for _, v := range values {
		s[v] = struct{}{}
	}
}

func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

// PRIORITY QUEUE

type Item struct {
	Value    string
	Priority int
	Index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool { return pq[i].Priority > pq[j].Priority }

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(v any) {
	n := len(*pq)
	item := v.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Update(item *Item, value string, priority int) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

func (pq PriorityQueue) ContainsValue(v string) bool {
	for _, item := range pq {
		if item.Value == v {
			return true
		}
	}
	return false
}

// CSV

func SaveResultsToFile(results [][]string) {
	if file, err := os.OpenFile("results/products.csv", os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		log.Fatal("Error with opening file: ", err)
	} else {
		defer file.Close()

		w := csv.NewWriter(file)
		if err = w.WriteAll(results); err != nil {
			log.Fatal("Error with writing to csv file: ", err)
		}
	}
}

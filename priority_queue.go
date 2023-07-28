package main

import (
	"container/heap"
	"fmt"
)

type ConfServerNode struct {
	Val  int
	Next *ConfServerNode
}

// 定义一个优先级队列
type PriorityConfServerQueue []*ConfServerNode

func (priorityConfServerQueue PriorityConfServerQueue) Len() int {
	return len(priorityConfServerQueue)
}

// 小于(<)就是小顶堆, 大于(>)就是大顶堆
func (priorityConfServerQueue PriorityConfServerQueue) Less(i, j int) bool {
	return priorityConfServerQueue[i].Val < priorityConfServerQueue[j].Val
}
func (priorityConfServerQueue PriorityConfServerQueue) Swap(i, j int) {
	priorityConfServerQueue[i], priorityConfServerQueue[j] = priorityConfServerQueue[j], priorityConfServerQueue[i]
	return
}

func (priorityConfServerQueue *PriorityConfServerQueue) Push(x interface{}) {
	*priorityConfServerQueue = append(*priorityConfServerQueue, x.(*ConfServerNode))
}

func (priorityConfServerQueue *PriorityConfServerQueue) Pop() interface{} {
	old := *priorityConfServerQueue
	n := len(old)
	x := old[n-1]
	*priorityConfServerQueue = old[0 : n-1]
	return x
}

func main() {
	priorityConfServerQueue := &PriorityConfServerQueue{}
	heap.Init(priorityConfServerQueue)
	node1 := &ConfServerNode{Val: 4}
	node2 := &ConfServerNode{Val: 3}
	node3 := &ConfServerNode{Val: 2}
	node4 := &ConfServerNode{Val: 6}
	node5 := &ConfServerNode{Val: 1}
	heap.Push(priorityConfServerQueue, node1)
	heap.Push(priorityConfServerQueue, node2)
	heap.Push(priorityConfServerQueue, node3)
	heap.Push(priorityConfServerQueue, node4)
	heap.Push(priorityConfServerQueue, node5)

	for priorityConfServerQueue.Len() > 0 {
		node := heap.Pop(priorityConfServerQueue)
		fmt.Print(node.(*ConfServerNode).Val, " ")
	}
}

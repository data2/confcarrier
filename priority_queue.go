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
	return len(p)
}

// 小于(<)就是小顶堆, 大于(>)就是大顶堆
func (priorityConfServerQueue PriorityConfServerQueue) Less(i, j int) bool {
	return p[i].Val < p[j].Val
}
func (priorityConfServerQueue PriorityConfServerQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	return
}

func (priorityConfServerQueue *PriorityConfServerQueue) Push(x interface{}) {
	*priorityConfServerQueue = append(*p, x.(*ConfServerNode))
}

func (priorityConfServerQueue *PriorityConfServerQueue) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	*priorityConfServerQueue = old[0 : n-1]
	return x
}

func main() {
	priorityConfServerQueue := &PriorityConfServerQueue{}
	heap.Init(p)
	node1 := &ConfServerNode{Val: 4}
	node2 := &ConfServerNode{Val: 3}
	node3 := &ConfServerNode{Val: 2}
	node4 := &ConfServerNode{Val: 6}
	node5 := &ConfServerNode{Val: 1}
	heap.Push(p, node1)
	heap.Push(p, node2)
	heap.Push(p, node3)
	heap.Push(p, node4)
	heap.Push(p, node5)

	for p.Len() > 0 {
		node := heap.Pop(p)
		fmt.Print(node.(*ConfServerNode).Val, " ")
	}
}

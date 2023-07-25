package main

import (
	"container/heap"
	"fmt"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

// 定义一个优先级队列
type PriorityQueue []*ListNode

func (p PriorityQueue) Len() int {
	return len(p)
}

// 小于(<)就是小顶堆, 大于(>)就是大顶堆
func (p PriorityQueue) Less(i, j int) bool {
	return p[i].Val < p[j].Val
}
func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	return
}

func (p *PriorityQueue) Push(x interface{}) {
	*p = append(*p, x.(*ListNode))
}

func (p *PriorityQueue) Pop() interface{} {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[0 : n-1]
	return x
}

func main() {
	p := &PriorityQueue{}
	heap.Init(p)
	node1 := &ListNode{Val: 4}
	node2 := &ListNode{Val: 3}
	node3 := &ListNode{Val: 2}
	node4 := &ListNode{Val: 6}
	node5 := &ListNode{Val: 1}
	heap.Push(p, node1)
	heap.Push(p, node2)
	heap.Push(p, node3)
	heap.Push(p, node4)
	heap.Push(p, node5)

	for p.Len() > 0 {
		node := heap.Pop(p)
		fmt.Print(node.(*ListNode).Val, " ")
	}
}

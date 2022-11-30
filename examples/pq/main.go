package main

import (
	"fmt"

	pq "github.com/sunvim/utils/priorityqueue"
)

func main() {
	q := pq.NewPriorityQueue(5, false)

	q.Put(&Node{Number: 5, Name: "hello"})
	q.Put(&Node{Number: 3, Name: "world"})
	q.Put(&Node{Number: 4, Name: "sky"})
	q.Put(&Node{Number: 1, Name: "mobus"})
	q.Put(&Node{Number: 2, Name: "sunqc"})

	for i := 0; i < 5; i++ {
		v, _ := q.Get(i)
		fmt.Printf("item: %+v \n", v)
	}

}

type Node struct {
	Number uint64
	Name   string
}

func (n *Node) Compare(other pq.Item) int {
	o := other.(*Node)
	if o.Number > n.Number {
		return 1
	} else if o.Number == n.Number {
		return 0
	}
	return -1
}

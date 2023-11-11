package linklist

import (
	"errors"
)

type List struct {
	head *node
}

type node struct {
	value int
	next  *node
}

func NewList() *List {
	return &List{}
}

func createNode(value int) *node {
	return &node{value: value}
}

func (list *List) Add(value int) {
	newNode := createNode(value)
	if list.head == nil {
		list.head = newNode
		return
	}
	curNode := list.head
	for curNode.next != nil {
		curNode = curNode.next
	}
	curNode.next = newNode
}

func (list *List) Delete(index int) (int, error) {
	if index < 0 {
		return 0, errors.New("index can't be less than 0")
	}
	if index >= list.Length() {
		return 0, errors.New("index out of range")
	}

	var value int
	if index == 0 {
		value = list.head.value
		list.head = list.head.next
		return value, nil
	}

	curNode := list.head
	for i := 1; i < index; i++ {
		curNode = curNode.next
	}

	value = curNode.next.value
	curNode.next = curNode.next.next

	return value, nil
}

func (list *List) Contains(value int) bool {
	curNode := list.head
	for curNode != nil {
		if curNode.value == value {
			return true
		}
		curNode = curNode.next
	}
	return false
}

func (list *List) Length() int {
	length := 0
	curNode := list.head
	for ; curNode != nil; curNode = curNode.next {
		length++
	}
	return length
}

func (list *List) ToSlice() []int {
	var result []int
	curNode := list.head
	for ; curNode != nil; curNode = curNode.next {
		result = append(result, curNode.value)
	}
	return result
}

func (list *List) Insert(value int, index int) error {
	if index < 0 {
		return errors.New("index can't be less than 0")
	}
	if index > list.Length() {
		return errors.New("index out of range")
	}

	newNode := createNode(value)

	if index == 0 {
		newNode.next = list.head
		list.head = newNode
		return nil
	}

	curNode := list.head
	for i := 1; i < index; i++ {
		curNode = curNode.next
	}

	newNode.next = curNode.next
	curNode.next = newNode

	return nil
}

func (list *List) Set(value int, index int) error {
	if index < 0 {
		return errors.New("index can't be less than 0")
	}
	if index >= list.Length() {
		return errors.New("index out of range")
	}

	curNode := list.head
	for i := 0; i < index; i++ {
		curNode = curNode.next
	}
	curNode.value = value

	return nil
}

func (list *List) Get(index int) (int, error) {
	if index < 0 {
		return 0, errors.New("index can't be less than 0")
	}
	if index >= list.Length() {
		return 0, errors.New("index out of range")
	}

	curNode := list.head
	for i := 0; i < index; i++ {
		curNode = curNode.next
	}
	return curNode.value, nil
}

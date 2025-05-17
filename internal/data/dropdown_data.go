package data

import "strings"

type DropdownData[T any] struct {
	enabled      bool
	index        int32
	selectorText string
	elements     []T
	hasContent   bool
}

func NewDropdownData[T any]() *DropdownData[T] {
	return &DropdownData[T]{
		selectorText: " ",
	}
}

func NewDropdownDataWithValues[T any](pairs ...Tuple[string, T]) *DropdownData[T] {
	return NewDropdownData[T]().SetData(pairs...)
}

func (this *DropdownData[T]) AddData(pair Tuple[string, T]) *DropdownData[T] {
	this.elements = append(this.elements, pair.Value)
	text := strings.ToUpper(pair.Key)
	if !this.hasContent {
		this.selectorText = text
		this.hasContent = !this.hasContent
		return this
	}

	this.selectorText += ";" + text
	return this
}

func (this *DropdownData[T]) SetData(pairs ...Tuple[string, T]) *DropdownData[T] {
	this.ClearData()
	for _, pair := range pairs {
		this.AddData(pair)
	}
	return this
}

func (this *DropdownData[T]) ClearData() *DropdownData[T] {
	this.selectorText = " "
	this.elements = nil
	this.hasContent = false
	return this
}

func (this *DropdownData[T]) IsSelected() bool {
	return this.index >= 0 && this.index < int32(len(this.elements))
}

func (this *DropdownData[T]) GetSelectedValue() (value T, exists bool) {
	exists = this.IsSelected()
	if exists {
		return this.elements[this.index], exists
	}

	return value, exists
}

func (this *DropdownData[T]) GetText() string {
	return this.selectorText
}

func (this *DropdownData[T]) IsActive() bool {
	return this.enabled
}

func (this *DropdownData[T]) SwitchActive() {
	this.enabled = !this.enabled
}

func (this *DropdownData[T]) GetIndex() *int32 {
	return &this.index
}

func (this *DropdownData[T]) IsEmpty() bool {
	return !this.hasContent
}

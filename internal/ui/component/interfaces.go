package component

import "iter"

type Renderer interface {
	Update()
	Render()
}

type PanelSelector interface {
	SetSelectedPanel(panel Renderer)
	IteratePanels() iter.Seq2[string, Renderer]
	PanelCount() int
}

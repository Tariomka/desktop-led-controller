package component

type EditPanel struct{ PanelBase }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	ep.renderPanel()
}

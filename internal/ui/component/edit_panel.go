package component

type EditPanel struct{ Panel }

func (ep *EditPanel) Update() { ep.resize() }

func (ep *EditPanel) Render() {
	ep.renderPanel()
}

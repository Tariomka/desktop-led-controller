package component

type MenuPanel struct{ PanelBase }

func (menu *MenuPanel) Update() { menu.resize() }

func (menu *MenuPanel) Render() {
	menu.renderPanel()
}

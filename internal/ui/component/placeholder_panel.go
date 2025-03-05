package component

type PlaceholderPanel struct{ PanelBase }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() { pp.renderPanel() }

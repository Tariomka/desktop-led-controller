package component

type PlaceholderPanel struct{ Panel }

func (pp *PlaceholderPanel) Update() { pp.resize() }

func (pp *PlaceholderPanel) Render() { pp.renderPanel() }

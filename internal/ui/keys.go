package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Create     key.Binding
	Delete     key.Binding
	Rider      key.Binding
	VSCode     key.Binding
	Codex      key.Binding
	Explorer   key.Binding
	DockerUp   key.Binding
	DockerDown key.Binding
	GitPull    key.Binding
	GitFetch   key.Binding
	Refresh    key.Binding
	Confirm    key.Binding
	Cancel     key.Binding
	Tab        key.Binding
	Actions    key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
	),
	Rider: key.NewBinding(
		key.WithKeys("o"),
	),
	VSCode: key.NewBinding(
		key.WithKeys("v"),
	),
	Codex: key.NewBinding(
		key.WithKeys("x"),
	),
	Explorer: key.NewBinding(
		key.WithKeys("e"),
	),
	DockerUp: key.NewBinding(
		key.WithKeys("u"),
	),
	DockerDown: key.NewBinding(
		key.WithKeys("s"),
	),
	GitPull: key.NewBinding(
		key.WithKeys("g"),
	),
	GitFetch: key.NewBinding(
		key.WithKeys("f"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
	),
	Actions: key.NewBinding(
		key.WithKeys("enter"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
	),
}

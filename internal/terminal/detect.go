package terminal

// NewBackend creates a Backend based on the given name.
// Valid names: "tmux", "cmux". Default is "tmux".
func NewBackend(name string) Backend {
	switch name {
	case "cmux":
		return &CmuxBackend{}
	default:
		return &TmuxBackend{}
	}
}

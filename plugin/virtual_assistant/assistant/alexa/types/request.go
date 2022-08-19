package types

type Request struct {
	Directive DirectiveOrEvent `json:"directive"`
}

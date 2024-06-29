package contextkey

type ContextKey int

const (
	SlogCtx ContextKey = iota // 0
	ReqIdCtx
)

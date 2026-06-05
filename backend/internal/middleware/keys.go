package middleware

// Context keys under which AuthRequired stores the authenticated identity in the
// gin context. Handlers read them with c.GetInt64/c.Get to learn who is calling.
const (
	CtxKeyUserID   = "userID"   // int64 user ID from the JWT claims
	CtxKeyUserRole = "userRole" // string role from the JWT claims ("customer"/"admin")
)

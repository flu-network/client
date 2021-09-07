package cli

// AddRequest is a dummy request useful for debugging.
type AddRequest struct {
	A, B int
}

// AddResponse is a dummy response to an AddRequest, useful for debugging.
type AddResponse struct {
	Response int
	Success  bool
}

// Add is a dummy method that adds two ints. Useful for debugging.
func (m *Methods) Add(args *AddRequest, reply *AddResponse) error {
	reply.Response = args.A + args.B
	reply.Success = true
	return nil
}

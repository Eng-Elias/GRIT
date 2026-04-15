package temporal

// errStopIter is a sentinel error to stop the commit iterator early.
var errStopIter = &stopIterError{}

type stopIterError struct{}

func (e *stopIterError) Error() string { return "stop iteration" }

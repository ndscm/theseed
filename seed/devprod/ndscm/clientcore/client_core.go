package clientcore

type ClientCore struct {
}

func (cc *ClientCore) NdShell(args []string) error {
	return NdShell(args)
}

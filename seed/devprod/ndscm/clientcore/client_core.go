package clientcore

type ClientCore struct {
}

func (cc *ClientCore) NdSetup(args []string) error {
	return NdSetup(args)
}

func (cc *ClientCore) NdShell(args []string) error {
	return NdShell(args)
}

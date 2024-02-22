package core

func (this *Core) approveNodeCert() error {
	_, err := execCmd(string(approveNodeCert))
	return err
}

package fileclient

func (fc *fclient) WorkspaceEnv() (string, error) {
	wc, err := getNewWsContext()
	if err != nil {
		return "", err
	}

	return wc.GetEnv()
}

func (fc *fclient) DirEnv() (string, error) {
	ctxEnv, err := fc.GetDataContext().GetEnv()
	if err == nil {
		return ctxEnv, nil
	}

	wsEnv, err := fc.WorkspaceEnv()
	if err != nil {
		return "", err
	}

	return wsEnv, nil
}

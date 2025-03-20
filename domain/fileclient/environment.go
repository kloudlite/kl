package fileclient

func (fc *fclient) WorkspaceEnv() (string, error) {
	wc, err := getNewWsContext()
	if err != nil {
		return "", err
	}

	return wc.GetEnv()
}

func (fc *fclient) DirEnv() (string, error) {
	wsEnv, err := fc.WorkspaceEnv()
	if err != err {
		return "", err
	}

	return wsEnv, nil
}

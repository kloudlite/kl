package fileclient

import "fmt"

func (fc *fclient) WorkspaceEnv() (string, error) {
	wc, err := getNewWsContext()
	if err != nil {
		return "", err
	}

	return wc.GetEnv()
}

func (fc *fclient) DirEnv() (string, error) {
	wsEnv, err := fc.WorkspaceEnv()
	if err == nil {
		return wsEnv, nil
	}

	ctxEnv, err := fc.GetDataContext().GetEnv()
	if err == nil {
		return ctxEnv, nil
	}

	return "", fmt.Errorf("no env found")
}

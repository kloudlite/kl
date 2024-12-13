package fileclient

func (fc *fclient) CurrentEnv() (string, error) {

	wc, err := getNewWsContext()
	if err != nil {
		return "", err
	}

	return wc.GetEnv()
}

func (fc *fclient) SelectEnv(string) error {
	return nil
}

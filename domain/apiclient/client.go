package apiclient

var Api = func() ApiClient {
	ac, err := New()
	if err != nil {
		panic(err)
	}

	return ac
}()

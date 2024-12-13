package fileclient

var File = func() FileClient {
	fc, err := New()
	if err != nil {
		panic(err)
	}

	return fc
}()

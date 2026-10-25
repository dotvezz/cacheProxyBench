package requests

type NopReader struct{}

func (r NopReader) Read(p []byte) (n int, err error) {
	return len(p), nil
}

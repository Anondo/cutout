package cutout

func executeFallbacks(fbf []func() (*Response, error)) (*Response, error) {

	fResp := &Response{}
	var err error

	for _, fb := range fbf {
		fResp, err = fb()

		if err != nil {
			continue
		}

		break
	}

	return fResp, err
}

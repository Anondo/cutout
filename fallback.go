package cutout

func executeFallbacks(fbf []func() (*Response, error)) (*Response, error) {

	fResp := &Response{}
	var err error

	for _, fb := range fbf { //as cutout supports multi-level fallbacks
		fResp, err = fb()

		if err != nil {
			continue // if one fails, try the next one
		}

		break
	}

	return fResp, err
}

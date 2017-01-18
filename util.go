package main

func present(reqFields ...string) bool {
	for _, field := range reqFields {
		if field == "" {
			return false
		}
	}

	return true
}

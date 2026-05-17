package main

func sendErr(m *model, err error) {
	m.err = &errData{
		errMsg:   err,
		errState: m.state,
	}

	m.state = StateError
	m.cursor = 0
}

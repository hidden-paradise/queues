package internal

import "errors"

type Job struct {
	Name    string
	Payload string
	Failed  bool
}

func NewJob(name, payload string) (Job, error) {
	if name == "" {
		return Job{}, errors.New("name cannot be empty")
	}
	if payload == "" {
		return Job{}, errors.New("payload cannot be empty")
	}

	return Job{
		Name:    name,
		Payload: payload,
		Failed:  false,
	}, nil
}

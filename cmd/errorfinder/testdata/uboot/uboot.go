package uboat

import "errors"

var ErrSentinel = errors.New("days of no horizon, claustrophobia, condition red")

var Datum = 1

type StructuredError struct{}

func (StructuredError) Error() string { return "don't crash" }

type DataContainer struct{}

func (DataContainer) AllesWasDrinIst() {}

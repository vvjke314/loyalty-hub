package model

import "errors"

var ErrAlreadyExits = errors.New("login already taken")
var ErrBadOrderNumber = errors.New("bad order number")
var ErrOrderAlreadyExists = errors.New("such order already exists")
var ErrOrderLoadedByAnotherPerson = errors.New("such order loaded by another person")
var ErrInsufficientFunds = errors.New("insufficient funds")

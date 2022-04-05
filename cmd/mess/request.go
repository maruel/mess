package main

import "time"

const (
	maxHardTimeout = 7*24*time.Hour + 10*time.Second
	maxExpiration  = 7*24*time.Hour + 10*time.Second
	minHardTimeout = time.Second
	evictionCutOff = 550 * 24 * time.Hour
)

//go:build !windows

package main

func readSecretLine() (string, error) { return readLine() }

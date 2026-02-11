package main

type printer interface {
	Printf(format string, v ...any)
	PrintErrf(format string, i ...any)
}

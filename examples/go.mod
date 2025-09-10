module github.com/sunvim/utils/examples

go 1.24.7

replace (
	github.com/sunvim/utils/cachem => ../cachem
	github.com/sunvim/utils/logger => ../logger
)

require (
	github.com/rs/zerolog v1.34.0
	github.com/sunvim/utils/cachem v0.0.0-00010101000000-000000000000
	github.com/sunvim/utils/logger v0.0.0-00010101000000-000000000000
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

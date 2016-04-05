# Go-opus #
A work-in-progress native implementation of RFC-6716, the Opus Audio Codec, in
Go. An unaffiliated project, [gopus](https://github.com/layeh/gopus)
provides a `cgo` wrapper for `libopus`.

## Justification ##
The dominant reason I have for undertaking this project is to further
my experience with Go. Simultaneously it allows me to apply software
design principles to the specification in order to develop clean code
in a "Go" way.

The other reason for this project is that I'd like to implement the
specification natively in Go. There are some good reasons for doing
this as related in
[dave.cheney.net/2016/01/18/cgo-is-not-go](). However, if pressured I would say
`gopus` is a thin enough wrapper and does the job elegantly enough
which brings me back to the above -- this is a mostly a learning
experience.

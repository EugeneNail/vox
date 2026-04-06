module github.com/EugeneNail/vox/message

go 1.26.1

require (
	github.com/EugeneNail/vox/lib-common v0.0.0
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
)

require (
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
)

replace github.com/EugeneNail/vox/lib-common => ../lib-common

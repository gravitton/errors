# Errors

[![Latest Stable Version][ico-release]][link-release]
[![Build Status][ico-workflow]][link-workflow]
[![Coverage Status][ico-coverage]][link-coverage]
[![Quality Score][ico-code-quality]][link-code-quality]
[![Go Report Card][ico-go-report-card]][link-go-report-card]
[![Go Dev Reference][ico-go-dev-reference]][link-go-dev-reference]
[![Software License][ico-license]][link-licence]

**Multi error**: Concurrent safe representation of a list of errors as a single error.

**Data error**: Additional context for error using data fields and cause (previous) error.


## Installation

```bash
go get github.com/gravitton/errors
```


## Usage

```go
package main

import (
	"github.com/gravitton/errors"
	"sync"
)

func Process() error {
	errs := errors.NewMulti()

	if err := process1(); err != nil {
		errs.Append(err)
	}

	if err := process2(); err != nil {
		errs.Append(err)
	}

	return errs.ErrorOrNil()
}

func ProcessConcurrent() error {
	errs := errors.NewMulti()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			if err := process(i); err != nil {
				errs.Append(errors.Wrap(err).WithField("process", i))
			}
		}(i)
	}

	wg.Wait()

	return errs.ErrorOrNil()
}
```


## Credits

- [Tomáš Novotný](https://github.com/tomas-novotny)
- [All Contributors][link-contributors]


## License

The MIT License (MIT). Please see [License File][link-licence] for more information.


[ico-license]:              https://img.shields.io/github/license/gravitton/errors.svg?style=flat-square&colorB=blue
[ico-workflow]:             https://img.shields.io/github/actions/workflow/status/gravitton/errors/main.yml?branch=main&style=flat-square
[ico-release]:              https://img.shields.io/github/v/release/gravitton/errors?style=flat-square&colorB=blue
[ico-go-dev-reference]:     https://img.shields.io/badge/go.dev-reference-blue?style=flat-square
[ico-go-report-card]:       https://goreportcard.com/badge/github.com/gravitton/errors?style=flat-square
[ico-coverage]:             https://img.shields.io/scrutinizer/coverage/g/gravitton/errors/main.svg?style=flat-square
[ico-code-quality]:         https://img.shields.io/scrutinizer/g/gravitton/errors.svg?style=flat-square

[link-author]:              https://github.com/gravitton
[link-release]:             https://github.com/gravitton/errors/releases
[link-contributors]:        https://github.com/gravitton/errors/contributors
[link-licence]:             ./LICENSE.md
[link-changelog]:           ./CHANGELOG.md
[link-workflow]:            https://github.com/gravitton/errors/actions
[link-go-dev-reference]:    https://pkg.go.dev/github.com/gravitton/errors
[link-go-report-card]:      https://goreportcard.com/report/github.com/gravitton/errors
[link-coverage]:            https://scrutinizer-ci.com/g/gravitton/errors/code-structure
[link-code-quality]:        https://scrutinizer-ci.com/g/gravitton/errors

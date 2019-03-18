# zagane

## Install

```
$ go get -u github.com/gcpug/zagane
```

## How to use

```
$ zagane pkgname
```

## Checks

* unstopiter: it finds iterators which did not stop

All checks can be ignored by a comment `//lint:ignore zagane reason`.

## Analyze with golang.org/x/tools/go/analysis

You can get analyzers of zagane from [zagane.Analyzers](https://godoc.org/github.com/gcpug/zagane/zagane/#Analyzers).
And you can use them with [unitchecker](https://golang.org/x/tools/go/analysis/unitchecker).

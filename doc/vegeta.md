# Load testing

Use [Vegeta](https://github.com/tsenart/vegeta) for simple load testing. See their website for installation.

Quick install by go `go get -u github.com/tsenart/vegeta`

```
$ cd test/vegeta
$ vegeta attack -targets=target -duration=10s -rate 10000 > results.bin
$ vegeta report -inputs=results.bin
```
or
```
$ vegeta attack -targets=target -duration=10s -rate 10000 | vegeta report
```
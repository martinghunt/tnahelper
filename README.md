# tnahelper

This is not intended to be used as a stand-alone tool.
It is a helper binary that's used internally by
[TNA](https://github.com/martinghunt/tna).


## Notes for developing the code

Run the tests:
```
go test -cover -v ./...
```

Release process:
```
git tag -a vX.Y.Z -m "Version X.Y.Z"
git push origin vX.Y.Z
```

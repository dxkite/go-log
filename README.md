# Go Simple Log

## Usage

```go
// log caller
log.SetLogCaller(true)
// log use text writer
log.SetOutput(NewTextWriter(os.Stdout))
if err := log.Output(1, "default", Linfo, "information\n"); err != nil {
	t.Error(err)
}
log.Println(Ldebug, Group("user"), "user info", "some", 1, "@", "11")
log.Println(Group("user"), Lwarn, "user info", "some", 1, "@", "11")
log.Println(Lerror, "user info", "some", 1, "@", "11")
log.Warn("user info", "some", 1, "@", "11")
log.Warn(Group("user"), "user info", "some", 1, "@", "11")
```
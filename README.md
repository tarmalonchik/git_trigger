# git_trigger

you can run command with 4 arguments

1. arg is repo name example: `tarmalonchik/git_trigger`

2. arg is the path to place where you want to store pulled project, example: `/root`

3. arg is command needed to add to make command example: `build`

4. arg is the branch you would like to use, example: `master`

result command:
```
go run cmd/main.go  tarmalonchik/project_name /root build master
```


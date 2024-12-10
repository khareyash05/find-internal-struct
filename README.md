# Find Internal Struct

A Go Package to find all the definitions for all types of objects used in the file <br>

To run the code,

Update the file path here with the intended file
```
mainFilePath := "/home/runner/ThinUntimelyNanocad/utils/h.go"
```

Update the project root path here -
```
projectRoot := "/home/runner/ThinUntimelyNanocad"
```

and run `go run main.go`

You will see output in struct_logs.txt like this

```
Structs in internal package: main/utils/sample
File: /home/runner/ThinUntimelyNanocad/utils/sample/g.go
  Struct: Alpha
    Field: Name, Type: string
```

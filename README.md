<p align="center">
    <img src="./logo/logo.svg" width="200"/> 
</p>

## About
**palang** is an educational hobby project that falls under the toy-compiler category. It is simple, statically-typed, single pass, and compiles to x86-64 assembly. For simplicity, I link it with `libc`.

### Requirements
- Linux x86-64
- Golang
- FASM
- GCC

## Usage

Build the compiler:
```console
$ go build -o palang main.go
```

Compile and run a program:
```console
$ ./palang ./examples/example-10.palang > main.asm
$ fasm main.asm
$ gcc -o main main.o
$ ./main
```

## Examples

Examples can be found in the [`examples/`](./examples) directory.


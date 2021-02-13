# Handle

Package handle reduces the amount of boilerplate required to handle errors.

To use handle, the enclosing function must use named return values. The 
error returned can be wrapped:

    do(s string) (err error) {
        check, handle := handle.Errorf(&err, "do(%s)", s); defer handle()
    
        // ...
    
        return
    }

or returned unmodified:

    do(s string) (err error) {
        check, handle := handle.Error(&err); defer handle()
    
        // ...
    
        return
    }

With a deferred handle any call to `check` with a non-nil error will cause the
enclosing function to return.

    // Return if err is not nil.
    f, err := os.Open(name); check(err)

## License

[MIT](LICENSE)

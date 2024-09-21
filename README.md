# FGScript
A macro system for Flowgorithm internal file format.

```ruby
# Comments outside of the function are ignored
# (They will not appear in the generated file)
fn Main () _ None
  # Comments are allowed only at the beggining of the line
  var n Integer

  println "Write a number:"
  scan n

  println "Fibonacci:" Fibonacci(n)
end

fn Fibonacci (n Integer) x Integer
  # We can define multiple variables at once
  var x y i Integer

  # Remember to always initialize variables before use
  set x 0
  set y 1

  # Loop n-1 times
  for i from 1 to n step 1
    # Expressions have the same syntax as in Flowgorithm
    set x x+y
    set y x-y
  end
end
```

# Installation
```sh
go install github.com/mandriota/fgscript/cmd/fgscript@latest
```

# Fgscript
A macro system for Flowgorithm internal file format.

```ruby
fn Main () _ None
  # Comments are allowed only at the beggining of the line
  # You cannot use comments outside of function
	
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

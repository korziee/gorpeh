// TODO write this as a test instead
// https://gobyexample.com/testing-and-benchmarking
#!/bin/bash

for j in {1..10}
do
  for i in {1..100}
  do
    # the & at the end backgrounds the process
    nc localhost 1234 &
  done
sleep 0.2
done


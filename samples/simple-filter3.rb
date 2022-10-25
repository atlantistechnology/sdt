#!/usr/bin/env ruby
def mod5?(items)
  items.to_a.select {|item| item % 5 == 0}
end

puts(mod5?(1..100))

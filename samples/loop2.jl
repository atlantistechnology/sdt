# Julia program to illustrate 
# the use of For loop
  
print("List Iteration\n") 
l = ["freaks", "for", "geeks"] 
for i in l
    println(i) 
end
  
# Function just to separate changes
function plus_two(x)
    #perform some operations
    return x + 2
end

# Iterating over a tuple (immutable) 
print("\nTuple Iteration\n") 
t = ("geeks", 
     "for", 
     "geeks") 
for i in t
    println(i) 
end
  
# Function just to separate changes
function plus_three(x)
    #perform some operations
    return x + 3
end

# Iterating over a String 
print("\nString Iteration\n")     
s = "Freaks"
for i in s 
    println(i) 
end

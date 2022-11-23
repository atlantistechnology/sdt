# Julia program to illustrate 
# the use of For loop
  
print("List Iteration\n") 
l = ["geeks", "for", "geeks"] 
for i in l
    println(i) 
end
  
# Iterating over a tuple (immutable) 
print("\nTuple Iteration\n") 
t = ("geeks", "for", "geeks") 
for i in t
    println(i) 
end
  
# Iterating over a String 
print("\nString Iteration\n")     
s = "Geeks"
for i in s 
    println(i) 
end

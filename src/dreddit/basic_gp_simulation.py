import random

x = 1000000

info = [False]*x

counter = 0
start = random.randint(0, x-1)
print start
info[start] = True

while True:
	for i in range(x):
		j = random.randint(0, x-1)
		if info[i]:
			info[j] = True
		if info[j]:
			info[i] = True
	counter += 1
	#print sum(info)
	if sum(info) >= x:
		print counter
		break

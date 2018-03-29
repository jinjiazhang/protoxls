import sys
import time
import random
import hashlib

def genkey_29():
	num = 0x65
	while num == 0x65:
		num = random.randint(0x63, 0x7a)
	return chr(num)

def genkey_31(k29):
	n29 = ord(k29)
	minn = max(0x61, 0x61 + 0x69 - n29)
	maxn = min(0x7a, 0x7a + 0x69 - n29)
	if n29 == 0x66:
		minn = max(0x6c, minn)
	num = random.randint(minn, maxn)
	return chr(num)

def genkey_27(k29, k31):
	n29 = ord(k29)
	n31 = ord(k31)
	num = n29 + n31 - 0x69
	return chr(num)

def genkey_fmt(user, fmt):
	reverse = user[::-1]
	md5hex = hashlib.md5(reverse).hexdigest()

	key = fmt
	for index in range(0, 16):
		key = key.replace('*', md5hex[index], 1)
	k29 = genkey_29()
	k31 = genkey_31(k29)
	k27 = genkey_27(k29, k31)

	# also can use fix char
	key = key.replace('?', k27, 1)
	key = key.replace('?', k29, 1)
	key = key.replace('?', k31, 1)
	return key

def genkey_windows(user):
	fmt = "windows-2*2*2*0*0*c*e*0*6*b*6*6*a*?*?*?*"
	return genkey_fmt(user, fmt)

def genkey_linux(user):
	fmt = "linux-e*d*1*7*9*a*a*1*0*0*2*3*4*?*?*?*"
	return genkey_fmt(user, fmt)

if __name__ == '__main__':
    user = "protoxls"
    if len(sys.argv) >= 2:
    	user = sys.argv[1]

    random.seed(time.time())
    key_w = genkey_windows(user)
    key_l = genkey_linux(user)
    print("Name: %s" % user)
    print("Key: %s" % key_w)
    print("Key: %s" % key_l)

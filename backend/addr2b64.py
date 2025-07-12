import json
import sys
import base64

if len(sys.argv) < 2:
    print("usage: addr2b64.py <address_book_file>")
    sys.exit(2)

with open(sys.argv[1], "r") as f:
    js = json.load(f)

as_str = base64.standard_b64encode(json.dumps(js).encode('utf-8')).decode('ascii')

print(as_str)

import requests
import json
import argparse

CONF_API_PREFIX = "/notifier"

def get_std_headers(token):
    return {"X-Token": token}


def get_reminders(host_name, token, ca_bundle):
    url = f'{host_name}{CONF_API_PREFIX}/api/reminder'
    response = requests.get(url, verify=ca_bundle, headers=get_std_headers(token))
    response.raise_for_status()
    data = response.json()

    res = list(map(lambda x: x["reminder"], data["reminders"]))

    return res


def get_address_book(host_name, token, ca_bundle):
    url = f'{host_name}{CONF_API_PREFIX}/api/addressbook'
    response = requests.get(url, verify=ca_bundle, headers=get_std_headers(token))
    response.raise_for_status()
    return response.json()


def save_address_book(data, host_name, token, ca_bundle):
    for i in data:
        url = f'{host_name}{CONF_API_PREFIX}/api/addressbook/{i["id"]}'
        body = {
            "addr_type": i["addr_type"],
            "address": i["address"],
            "display_name": i["display_name"],
            "is_default": i["is_default"] 
        }
        
        response = requests.put(url, data=json.dumps(body).encode('utf-8'), verify=ca_bundle, headers=get_std_headers(token))
        response.raise_for_status()


def save_reminders(data, host_name, token, ca_bundle):
    for i in data:
        url = f'{host_name}{CONF_API_PREFIX}/api/reminder/{i["id"]}'
        body = {
            "description": i["description"],
            "kind": i["kind"],
            "param": i["param"],
            "recipients": i["recipients"],
            "spec": i["spec"],
            "warning_at": i["warning_at"]
        }
        
        response = requests.put(url, data=json.dumps(body).encode('utf-8'), verify=ca_bundle, headers=get_std_headers(token))
        response.raise_for_status()


def do_backup(host_name, ca_bundle, out_file, token):
    addr_book = get_address_book(host_name, token, ca_bundle)
    reminders = get_reminders(host_name, token, ca_bundle)
    backup = {"address_book": addr_book, "reminders":reminders}
    bkp = json.dumps(backup).encode('utf-8')
    with open(out_file, "wb") as f:
        f.write(bkp)


def do_restore(in_file, host_name, token, ca_bundle=None):
    with open(in_file, "rb") as f:
        data = json.load(f)
    
    save_address_book(data["address_book"], host_name, token, ca_bundle)
    save_reminders(data["reminders"], host_name, token, ca_bundle)


def main():
    COMMANDS = ["backup", "restore"]
    parser = argparse.ArgumentParser(description='Tool, um Backups des Datenbestandes von mobilenotifier zu erstellen und wiederherzustellen')
                                        
    parser.add_argument("command", choices=COMMANDS, help="Was soll getan werden: backup oder restore.")
    parser.add_argument("-n", "--host-name", required=True, help="Hostnamen des mobile notifier APIs")
    parser.add_argument("-o", "--output-file", default=None, help="Ausgabedatei für backup")
    parser.add_argument("-i", "--input-file", default= None, help="Eingabedatei für restoe")
    parser.add_argument("-c", "--ca-bundle", default=None, help="Datei, die das CA-Bundle enthält. Falls das benötigt wird")
    parser.add_argument("-t", "--token-file", required=True, default=None, help="File which contains a valid JWT for accessing the REST backend")

    args = parser.parse_args()

    try:
        with open(args.token_file, "rb") as f:
            token_raw = f.read()

        token = token_raw.decode("utf-8")
        token = token.strip()

        if args.command == "backup":
            if args.output_file == None:
                print("Ausgabedatei muss angegeben werden")
                return

            do_backup(args.host_name, args.ca_bundle, args.output_file, token)

        if args.command == "restore":
            if args.input_file == None:
                print("Einagbedatei muss angegeben werden")
                return
            
            do_restore(args.input_file, args.host_name, token, args.ca_bundle)
    except Exception as e:
        print(e)
    except KeyboardInterrupt:
        pass

main()
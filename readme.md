# Mobile notifier

This project implements a `Vue.js` webapp (and the corresponding backend in `Go`) which allows to manage reminders for events and to define notifications 
for these reminders. Currently notifications via IFTTT and e-mail are implemented. Here IFTTT is used to send text messages (SMS) or push
messages to mobile phones. It can be run either as a classic daemon on a server or in a kubernetes cluster. The tooling to create the neccessary images and 
the needed `.yml` kubernetes config files are also part of the project.

As the core functionality of this project is already part of the calender apps of all well known mobile phone OSs and/or is already offered as a SaaS solution
from several vendors (even though these often lack sending reminders via SMS) the main purpose for creating this application was to get some exercise using `Vue.js`
and kubernetes. Nonetheless it is functional and useful if you want to run it on your own systems. Please read the [important notice](#important-notice) below.

# Building the software

## Frontend

You need to install `node.js` before being able to build the frontend. After that the frontend can be built by running either `npm run builddev` or `npm run buildprod` 
from the `frontend` subdirectory. The API URL where the app expects its backend service can be configured through `.env.dev` and `.env.prod`. The resulting package 
is stored in the `dist` subdirectory of the `frontend` folder from where it can be copied or referenced.

## Backend

The backend depends on [`bbolt`](https://github.com/etcd-io/bbolt) as a file oriented database and [`swaggo http-swagger`](https://github.com/swaggo/http-swagger)
for auto generation of Swagger documentation. Install these dependencies according to their documentation before building the backend service. The container image 
can be built by running `sh conbuild.sh` in the `backend` subdirectory. If you want to run the backend service as a simple binary you need to execute 
`swag init -g controller/swagger_base.go` followed by `go build` in the `backend` subdirectory. The backend service can be configured using the following environment 
variables:

|Name|Value|Secret|
|-|-|-|
|LOCALDIR| If set then this variable has to specifiy the path to the directory where the compiled fronted can be found| No |
|DB_PATH| This variable has to define the file name of the `bbolt` database file| No |
|SWAGGER_URL| This variable has to specify the URL under which the swagger documentation can be reached | No |
|MN_CLIENT_TZ| Here you have to specify the time zone where the user lives. In my case this is `Europe/Berlin` according to the [IANA time zone database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)| No |  
|NOTIFIER_API_KEY| Secret value which is used to authorize calls to the `/notifier/api/send/{recipient}` API method | Yes |
|MN_MAIL_SERVER| This variable has to contain the FQDN of the SMTP server which is used to send mail notifications| No |
|MN_MAIL_SERVER_PORT| Here the port used by the SMTP server defined above can be set | No |
|MN_MAIL_SENDER_ADDR| This variable has to contain the mail address which is used as the sender address for mail notificarions| Yes |
|MN_MAIL_SENDER_PW| Here the password used by the sender address on the configured SMTP server has to be specified | Yes |
|MN_MAIL_SUBJECT| This variable determines the Subject of notification e-mails | No |
|MN_ADDR_BOOK| If set then this variable has to contain a base64 encoded JSON string which specifies recipients which are to be merged into the database. The format of the JSON data is specified below (see Address Book)| Yes |
|IFTTT_API_KEY| The IFTTT API key used when sending text (SMS) or push messages| Yes |

All variables marked as being secret in the table above have to be provided in a kubernetes secret named `notifier-secret` when the backend is run in a kubernetes cluster. All non secret variables
are defined in `notifier.yml` through a config map named `backend-config`. Strictly speaking the mail address of the sender is not a secret but it does not belong in the `notifier.yml` file
which is part of this repository.

## Address Book

The address book is a data structure which specifies each known recipient and is stored in the database file used by the backend. If the address book is represented as a JSON array then each
entry has to conform to the following schema

|Name | Type | Value |
|-|-|-|
|display_name| string | This specifies the name of the recipient as displayed in the webapp. This value can be changed|
|id| string | This is a UUID which defines the recipient internally. It this is changed a new recipient is created |
|addr_type| string | Currently the address types `IFTTT` and `Mail` are defined |
|address|string| The address in the context of the address type. I.e. currently either the mail address of the recipient or the name of the IFTTT recipe |
|is_default|bool| Is `true` if this recipient should be a default recipient for new reminder notifications |

Example:

```json
[
    {
        "display_name": "Pushmessage",
        "id": "DD187B04-BBD1-4E39-82AE-A827F499B7C3",
        "address": "MakePushMessage",
        "addr_type": "IFTTT",
        "is_default": true
    },
    {
        "display_name": "Mailrecipient",
        "id": "99DA295B-4E49-48BA-A2E4-97BFD1D3EE5F",
        "address": "niemand@nix.de",
        "addr_type": "Mail",
        "is_default": false
    },    
    .....
]
```

This software assumes that there is an IFTTT webhook for each recipient of type `IFTTT` which can be called via the following URL `https://maker.ifttt.com/trigger/{address}/with/key/{IFTTT_API_KEY}`.

The address book is persisted in the database and can be managed through mobilenotifier's web UI. Alternatively the contents of the address book can be changed through the environment variable
`MN_ADDR_BOOK`. If it is set to a base64 encoded JSON string (having the structure described above) then the values contained in the JSON data are merged into the database. Setting this variable
is optional and should only be done during development.

This repo also contains a small Python script `addr2b64.py` which allows to generate a compacted and base64 encoded version of a JSON address book. The output of this script can be
used to set the `MN_ADDR_BOOK` environment variable. When set through a kubernetes secret the script output has to be base64 encoded a second time.

# Configuring IFTTT

The software expects to be able to call IFTTT webhooks ("normal" webhooks without a JSON payload). The action to take if the webhook trigger was received is either sending an SMS via `Android SMS` or 
`Send a notification from the IFTTT app` for a push message. In these actions add `Value1` as the sole "ingredient". `Value1` will then be replaced by the actual message to be sent. Please note that the
webhook feature requires you to subscribe to an IFTTT pro account which is not free.

I utilize an older Android phone which is not in daily use anymore as the "host" for the IFTTT "Android SMS" action. This phone then sends the reminder SMS messages to the recipients in our family. My first
idea was to use the API of the messaging app Threema to send messages but unfortunately they closed their API for non business customers. I know that IFTTT has some actions which allow to send WhatsApp
messages but these seem to have somewhat arbitrary limits on the number of messages which can be sent and all looked a bit dodgy. Signal has an API which was reverse engineered from the Android app but that
also falls in the dodgy category in my book. Telegram has an API but I don't like Telegram. There are dedicated SMS Gateway providers but for my expected volume of messages I did not want to embarass myself
when talking to them.

# Deploying the software

## Important notice 

Even though the infrastructure for implementing authentication and authorization is there (see `backend/tools/auth_handler.go`) nearly all methods of the backend service can be called without 
authentication and authorization. The only exception is the API method to manually send a notification message which uses an API key. Or in other words **if you deploy this software in such a way 
that it can be called from the public internet you effectively allow other people to send text messages, e-mails and push messages to all recipients configured in your IFTTT recipes and/or specified
in the address book**. I assume this is not what you want especially if there is a cap on the number of text messages which can be sent without additional costs in your cell phone plan. Additionally
there may be bugs in this software which allow people to successfully attack your systems in other ways. I am not aware of such bugs but I might have made mistakes.

In other words: **Please, do not deploy this software in a fashion which exposes it to the public internet**. I run it in my home network on my own K3S kubernetes cluster and for this environment the
lack of authentication and authorization is acceptable.

## Running the software during development

You first have to build the frontend by executing `npm run builddev` in the `frontend` subdirectory and `swag init -g controller/swagger_base.go` followed by `go build` in the `backend` subdirectory. 
I use a shell script called `start.sh` for running all programs locally. Here an example for this script

```sh
export LOCALDIR=../frontend/dist/ 
export DB_PATH=../notifier.db
export SWAGGER_URL="http://localhost:5100/notifier/api/swagger/doc.json"
export MN_CLIENT_TZ=Europe/Berlin
export NOTIFIER_API_KEY=xxxxxxx
export MN_MAIL_SERVER=mail.whereever.net
export MN_MAIL_SERVER_PORT=587
export MN_MAIL_SENDER_ADDR=sender@whereever.net
export MN_MAIL_SENDER_PW=xxxxxxxxxxxxxxxxxxx
export MN_MAIL_SUBJECT="Erinnerung MobileNotifier"
./notifier
``` 

After that the webapp can be used via the URL `http://localhost:5100/notifier/app/index.html`. Please note that the script does not set the environment variable `IFTTT_API_KEY` in order to prevent sending
actual text messages when simply testing the software. If the environment variable is not set the backend service automatically switches to a dummy sender which only logs the fact that a message would have
been sent. If you set the environment variable to the correct value for your IFTTT account then your webhooks are called for real.

## Running the software on a server

In order to run this software on one of your machines in a classical fashion you can build a production version of the frontend after setting `VITE_API_URL` in `.env.prod` to the correct value. 
Then set `SWAGGER_URL` and  `LOCALDIR` in your environment to the correct values and you will be able to run the software without a separate web server. Of course you also can use a web server if
you prefer doing so. In this case make sure that `LOCALDIR` is not set.

## Kubernetes deployment

Important notice: Currently there is no helm chart for customizing the deployment.

The file `notifier.yml` specifies two deployments with its corresponding services as well as an ingress. The deployment `notifier-backend-deployment` deals with the backend REST service. The other
deployment `notifier-nginx` creates an nginx instance which serves the static `Vue.js` webapp. These two deployments are tied together with a `traefik` ingress which also performs TLS termination.

The container image for the backend can be created as described above using the script `conbuild.sh`. As the the image is created locally and not pulled from a registry the backend deployment has 
its `imagePullPolicy` set to `Never`. I copy the image over to the machines making up my K3S cluster and after that import it manually via `k3s ctr images rm` and `k3s ctr images import`

The file `pvs.yml` defines two persistent volume claims. One (`notifier-appdata`) for the `bbolt` database file of the backend service and the other (`notifier-statichtml`) for the static files of the
`Vue.js` webapp which is used by the nginx instance. Currently the PVCs use the NFS storage class which might not be available in your cluster. This is convenient for me as my NAS drive is configured to
provide the persistent volumes which makes it very easy to deploy the webapp: I simply mount the corresponding folder, copy the files from the `dist` folder on my development machine to the NFS drive and
do a `kubectl rollout restart` of the nginx deployment. The persistent volume claim for the `bbolt` file is defined to have `accessMode` `ReadWriteOnce`. Therefore it is not possible to run more than one
replica of the backend service.
 
You will need an additional file (not part of this repo) containing all the secrets including the TLS certificate and its private key. This file has to have the following structure

```yml
apiVersion: v1
kind: Secret
metadata:
  name: notifier-secret
data:
  IFTTT_API_KEY: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
  NOTIFIER_API_KEY: AAAAAAAAAAAAAAAAAAAAAAAAAAAAa
  MN_MAIL_SENDER_PW: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
  MN_MAIL_SENDER_ADDR: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAaa

---
apiVersion: v1
kind: Secret
metadata:
  name: notifier-tls
data:
  tls.crt: |
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
  tls.key: |
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
    AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA
type: kubernetes.io/tls
```

I have not configured Let's Encrypt. I use my own [minica](https://github.com/rmsk2/minica) to issue certificates. I run this software in its own kubernetes namespace.

# Performing backups

The script `backup.py` can be used to create a backup of the current state of a mobile notifier instance. Additionally it allows to restore this backup into an
"empty" mobile notifier instance. The script uses mobile notifier's REST API to extract and restore the data which makes up the state.

```
usage: backup.py [-h] -n HOST_NAME [-o OUTPUT_FILE] [-i INPUT_FILE] [-c CA_BUNDLE] {backup,restore}

Tool, um Backups des Datenbestandes von mobilenotifier zu erstellen und wiederherzustellen

positional arguments:
  {backup,restore}      Was soll getan werden: backup oder restore.

options:
  -h, --help            show this help message and exit
  -n HOST_NAME, --host-name HOST_NAME
                        Hostnamen des mobile notifier APIs
  -o OUTPUT_FILE, --output-file OUTPUT_FILE
                        Ausgabedatei für backup
  -i INPUT_FILE, --input-file INPUT_FILE
                        Eingabedatei für restoe
  -c CA_BUNDLE, --ca-bundle CA_BUNDLE
                        Datei, die das CA-Bundle enthält. Falls das benötigt wird
```

The option `-c/--ca-bundle` can be used to reference a file which contains a private root certificate in PEM-format which is to be used to verify the TLS server
certificate. If you do not use TLS or use a certificate of a publicly trusted CA then you can ignore this option. The option `-n/--host-name` has to specify
not only the host name of the machine which runs mobile notifier's backend but also the protocol, i.e. `http` or `https`. 

Let's assume the backend runs on the machine `kubernetes-cluster.example.com` which uses a TLS certificate issued by private root, where the root certifciate 
is stored in the file `my-private-root.pem`. Then the following commands can be used to create 

`python3 backup -o mobilenotifier.bak -n https://kubernetes-cluster.example.com -c my-private-root.pem` 

and restore a backup

`python3 restore -i mobilenotifier.bak -n https://kubernetes-cluster.example.com -c my-private-root.pem`

You can use the variable `CONF_API_PREFIX` to specify any additional path components which are needed to access the API in addition to the host name. The default
value is `/notifier`.

# Using the webapp

The webapp is currently in german and I did not attempt to add any sort of internationalization, sorry. Here are screenshots of the five different panels of the webapp.

![](/monat.png?raw=true "Monatliche Ereignisse")

![](/neu.png?raw=true "Neues Ereignis")

![](/alle.png?raw=true "Alle Ereignisse")

![](/about.png?raw=true "Über mobilenotifier")

![](/empfaenger.png?raw=true "Verwaltung der Empfänger")

And here a screenshot of the Swagger API info page.

![](/swagger.png?raw=true "Swagger API Info")
#!/bin/bash
# Creates self-signed certificates to use with surf
# Please provide the domain name you are creating the certificates for
#
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PRIVATE_KEY="base.key"
ROOT_CRT="base.pem"
SERVER_KEY="server.key"
SERVER_CSR="server.csr"
SERVER_CRT="server.crt"

DAYS=10000
NUMBITS=2048

DEFAULT_SERVER="*.localhost:80"
MANUAL="
Creates self-signed certificates to use with surf server and client.
You need to pass the domain name that server will be hosted as a wildcard.
For example: If server is hosted on www.mySurf.com, then call \`./selfSignedKeys "*.mySurf.com"\`
"

create_crt() {
	cd $DIR

	if [ -f ${PRIVATE_KEY} ]
	then
		read -p "Private key ${PRIVATE_KEY} was found in this directory. Are you sure you want to continue [y/n]? " -n 1 -r
		if [[ ! $REPLY =~ ^[Yy]$ ]]
		then
		    exit 1
		fi
	fi
	printf "\n"

	# Generates keys and certificates
	openssl genrsa -out ${PRIVATE_KEY} ${NUMBITS}
	openssl req -new -x509 -nodes -key ${PRIVATE_KEY} -days ${DAYS} -subj "/CN=${DOMAIN}" -out ${ROOT_CRT}
	openssl genrsa -out ${SERVER_KEY} ${NUMBITS}
	openssl req -new -key ${SERVER_KEY} -subj "/CN=${DOMAIN}" -out ${SERVER_CSR}
	openssl x509 -req -in ${SERVER_CSR} -CA ${ROOT_CRT} -CAkey ${PRIVATE_KEY} -CAcreateserial -days ${DAYS} -out ${SERVER_CRT}

	cp ${ROOT_CRT} ./assets/client/tls/surfroot.crt
}

for arg in "$@"; do
    shift
    case "$arg" in
        "--help")       set -- "$@" "-h" ;;
        *)              set -- "$@" "$arg"
    esac
done

while getopts "h" option; do
    case "${option}" in
        h)
            printf '%s\n' "$MANUAL"
            exit 0
            ;;
    esac
done
shift $((OPTIND-1))

DOMAIN=$1
if [ -z "$1" ]
then
	echo "Warning: no server was provided; using default $DEFAULT_SERVER"
	DOMAIN=${DEFAULT_SERVER}
fi

create_crt


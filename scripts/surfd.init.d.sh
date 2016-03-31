#!/bin/sh
### BEGIN INIT INFO
# Provides:          surf
# Required-Start:    $all
# Required-Stop:     $local_fs
# Default-Start:     S
# Default-Stop:      0 6
# Short-Description: Starts surf server
# Description:       Starts surf server
### END INIT INFO
PATH="$PATH:/usr/local/bin:/usr/sbin:/sbin:/usr/bin/:/bin"
SERVER="development"
PID_DIR="/var/run"
PID_FILE="$PID_DIR/surfd.pid"
LOG_FILE="/var/log/surfd.log"
LOCK_FILE="/var/lock/surfd.lock"

SURFD=$(which surfd)
KEY="/path/to/key"
CRT="/path/to/crt"
DOMAIN="oceanapp.io"
HTTP_PORT=80
HTTPS_PORT=443

USAGE="Usage requires {start|stop|force-stop|force-reload|restart|status}"

. /lib/init/vars.sh
. /lib/lsb/init-functions

get_pid() {
    echo "$(cat "$PID_FILE")"
}

is_running() {
    if pid_file_exists
    then
        PID=$(get_pid)
        ! [ -z "$(ps aux | awk '{print $2}' | grep "^$PID$")" ]
    else
        ! [ -z "" ]
    fi
}

pid_file_exists() {
    [ -f "$PID_FILE" ]
}

has_lock() {
    [ -f "$LOCK_FILE" ]
}

remove_lock() {
    if has_lock
    then
        rm -f "$LOCK_FILE"
    fi
}

remove_pid_file() {
    if pid_file_exists
    then
        rm -f "$PID_FILE"
    fi
}

get_status() {
    if is_running
    then
        PID=$(get_pid)
        echo "Surfd is running with PID $PID"
    else
        echo "Surfd is not running"
    fi
}


start_process() {
    if has_lock
    then
        log_warning_msg "Surfd lock file exists"
        exit 1
    else
        log_action_msg "Starting Surfd service ($SERVER) ..."

				OCEAN_SERVER=${SERVER} \
		        ${SURFD} \
		            -tlsKey=${KEY} \
		            -tlsCrt=${CRT} \
		            -domain="${DOMAIN}" \
		            -httpAddr=":$HTTP_PORT" \
		            -httpsAddr=":$HTTPS_PORT" \
		            -log="${LOG_FILE}" &

        PID=$!
        echo $PID > "$PID_FILE"
        echo "Started Surfd with pid $PID"
        touch $LOCK_FILE
    fi
}

stop_process() {
    if is_running
    then
        log_action_msg "Stopping Surfd service ..."
        PID=$(get_pid)
        kill -9 $PID
        remove_pid_file
        remove_lock
    else
        log_action_msg "Surfd is already stopped"
    fi
}

force_stop() {
    remove_pid_file
    remove_lock
}

case $1 in
    start)
        if is_running
        then
            PID=$(get_pid)
            echo "Process is running with PID $PID. Use restart|force-reload or stop first"
            log_end_msg 1
            exit 1
        fi

        start_process
        log_end_msg 0
    ;;

    stop)
        stop_process
        log_end_msg 0
    ;;

    force-stop)
        force_stop
        log_end_msg 0
    ;;

    restart|force-reload)
        force_stop
        start_process
        log_end_msg 0
    ;;

    status)
        get_status
    ;;

    *)
        echo $USAGE >&2
        exit 1
    ;;
esac

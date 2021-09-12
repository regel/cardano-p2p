#!/bin/bash -x
# shellcheck disable=SC2086,SC2034
# shellcheck source=/dev/null

# Description : Helper function to validate that input is a number
#             : $1 = number
isNumber() {
  [[ -z $1 ]] && return 1
  [[ $1 =~ ^[0-9]+$ ]] && return 0 || return 1
}

# Description : Helper function to validate IPv4 address
#             : $1 = IP
isValidIPv4() {
  local ip=$1
  [[ -z ${ip} ]] && return 1
  if [[ ${ip} =~ ^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$ || ${ip} =~ ^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$ ]]; then 
    return 0
  fi
  return 1
}

# Description : Helper function to validate IPv6 address, works for normal IPv6 addresses, not dual incl IPv4
#             : $1 = IP
isValidIPv6() {
  local ip=$1
  [[ -z ${ip} ]] && return 1
  ipv6_regex="^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$"
  [[ ${ip} =~ ${ipv6_regex} ]] && return 0
  return 1
}

# Description : Helper function to validate Host name value
#             : $1 = Hostname
isValidHost() {
  local name=$1
  [[ -z ${name} ]] && return 1
  hostname_regex="(?=^.{4,253}$)(^(?:[a-zA-Z0-9](?:(?:[a-zA-Z0-9\-]){0,61}[a-zA-Z0-9])?\.)+([a-zA-Z]{2,}|xn--[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])$)"
  [[ ${name} =~ ${hostname_regex} ]] && return 0
  return 1
}

usage() {
  cat <<-EOF
		Usage: $(basename "$0") [-j number] [-w timeout] pstate.json
		Pool vet. Vet pool metadata-hash and host:port liveness
		-j    Run n jobs in parallel
		-w    Timeout. If a connection and stdin are idle for more than timeout seconds, then the connection is silently closed
		
		EOF
  exit 1
}

N_JOBS=8
TMOUT=1
FILE="pstate.json"

while getopts :j:w: opt; do
  case ${opt} in
    j ) N_JOBS=$OPTARG ;;
    w ) TMOUT=$OPTARG ;;
    \? ) usage ;;
  esac
done
shift $((OPTIND -1))

if [ ! -f "$1" ]; then
  echo "File not found. Can't open file: $1" >&2
  exit 1
fi
FILE=$1

echo_meta() {
  IFS=$'\t' read -r pubkey url hash host port <<< "$1"
  # if bash gets a connection failure prior to the specified timeout, then bash will exit with an exit code of 1 which timeout will also return.
  # if bash isn't able to establish a connection and the specified timeout expires, then timeout will kill bash and exit with a status of 124
  TMOUT=${TMOUT:-1}
  tcpprobe=1
  ! isNumber $port && { echo "null port, ignoring host '$host'" return ;} >&2
  if isValidHost "$host" ; then
    timeout $TMOUT bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
    tcpprobe=$?
  elif isValidIPv4 "$host" ; then
    timeout $TMOUT bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
    tcpprobe=$?
  elif isValidIPv6 "$host" ; then
    timeout $TMOUT bash -c "cat < /dev/null > /dev/tcp/$host/$port" 2>/dev/null
    tcpprobe=$?
  fi
  metadata=$(curl -Ls --connect-timeout 5 "$url")
  echo "$metadata" | jq -e .ticker &> /dev/null
  if [ $? -eq 0 ] && [ -n "$metadata" ]; then
    hashed=$(cardano-cli stake-pool metadata-hash --pool-metadata-file /dev/stdin <<< "$metadata")
    echo "$metadata" | jq \
	    --arg publicKey "$pubkey" \
	    --arg hashed "$hashed" \
	    --arg probe "$tcpprobe" \
	    '{metadata: .} + {publicKey: $publicKey} + {extended: {hash: $hashed, probe:$probe}}'
  fi
}
export -f echo_meta
export -f isNumber
export -f isValidHost
export -f isValidIPv4
export -f isValidIPv6
export TMOUT

(jq -r '
  .[]
    |select(.relays[0]."single host name".dnsName != null)
    |[.publicKey, .metadata.url, .metadata.hash, .relays[0]."single host name".dnsName, .relays[0]."single host name".port]
    | @tsv
 ' "$FILE" ;
 jq -r '
  .[]
    |select(.relays[0]."single host address".IPv4 != null)
    |[.publicKey, .metadata.url, .metadata.hash, .relays[0]."single host address".IPv4, .relays[0]."single host address".port]
    | @tsv
 ' "$FILE";
  jq -r '
  .[]
    |select(.relays[0]."single host address".IPv6 != null)
    |[.publicKey, .metadata.url, .metadata.hash, .relays[0]."single host address".IPv6, .relays[0]."single host address".port]
    | @tsv
 ' "$FILE";
 )|
    parallel -j $N_JOBS echo_meta


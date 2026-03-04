# find gawk
find_gawk_binary() {
  gawk_path="$(which gawk)"
  if [[ $? == 0 ]]; then
    if [ -x "$gawk_path" ]; then
      awk_version_str="$($gawk_path --version)"
      if [[ $? == 0 ]]; then
        if echo "$awk_version_str" | grep -q 'GNU Awk'; then
          echo "$gawk_path"
          exit 0
        fi
      fi
    fi
  fi

  exit 1
}

# detect timezone
detect_timezone() {
  if [[ $TZ != "" ]]; then
    echo "$TZ"
    exit 0
  fi

  if [ -r /etc/timezone ]; then
    host_timezone="$(cat /etc/timezone)"
    if [[ $? == 0 ]]; then
      echo "$host_timezone"
      exit 0
    fi
  fi

  host_timezone="$(timedatectl show --property=Timezone --value)"
  if [[ $? == 0 ]]; then
    echo "$host_timezone"
    exit 0
  fi

  zone_file="$(find /usr/share/zoneinfo -type f -exec cmp -s /etc/localtime '{}' \; -print)"
  if [[ $? == 0 ]]; then
    host_timezone="$(echo "$zone_file" | sed -e 's|^/usr/share/zoneinfo/||' -e '/posix/d')"
    if [[ $? == 0 ]]; then
      echo "$host_timezone"
      exit 0
    fi
  fi

  exit 1
}

# detect os
detect_os_kind() {
  os_kind=""
  case "$(uname -s)" in
    Linux)
      os_kind="linux"
      ;;
    Darwin)
      os_kind="macos"
      ;;
    FreeBSD|OpenBSD|NetBSD|DragonFly)
      os_kind="bsd"
      ;;
    *)
      exit 1      
  esac

  echo "$os_kind"
  exit 0
}

get_file_size() { 
  case "$os_kind" in
    linux)
      stat -c %s "$logfile"
      ;;
    macos|bsd)
      stat -f %z "$logfile"
      ;;
    *)
      echo "error:internal error; invalid os_kind '$os_kind'" 1>&2
      return 1
  esac
}

get_file_modtime() {
  case "$1" in
    linux)
      stat -c %y "$2"
      ;;
    macos|bsd)
      stat -f "%SB" -t "%Y-%m-%d %H:%M:%S" "$2"
      ;;
    *)
      echo "error:internal error; invalid os_kind '$os_kind'" 1>&2
      return 1
  esac
}

# awk funcs
awk_funcs='
  function printPercentage(numCur, numTotal) {
    curPercent = int(numCur/numTotal*20)
    if (curPercent != lastPercent) {
      print "N:p:" curPercent*5 >> "/dev/stderr"
      lastPercent = curPercent
    }
  }

  function printIndexLine(outfile, timestr, linenr, bytenr) {
    print "idx\t" timestr "\t" linenr "\t" bytenr >> outfile;
  }

  function awktime_hhmm() {
    return substr($0, 12, 5);
  }

  function awktime_month() {
    return substr($0, 6, 2);
  }

  function awktime_year(month) {
    return substr($0, 0, 4);
  }

  function awktime_day() {
    return substr($0, 9, 2)
  }

  function validPrefix() {
      return $0 ~ /^[0-9]{4}-[0-9]{2}-[0-9]{2}[ T][0-9]{2}:[0-9]{2}:[0-9]{2},[0-9]{3} /
  }
'
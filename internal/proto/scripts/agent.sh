# trap 'echo "exit_code:$?"' EXIT

source "$(dirname "${BASH_SOURCE[0]}")/agent_lib.sh"

STAGE_INDEX_FULL=1
STAGE_INDEX_APPEND=2
STAGE_QUERYING=3
STAGE_DONE=4

max_num_lines=100

concat_cmds_array() {
  local first=1
  for cmd in "${cmds[@]}"; do
    if [[ $first == 1 ]]; then
      first=0
    else
      echo -n " && "    
    fi
    echo -n "${cmd//\'/\'\"\'\"\'}"
  done
}

while [[ $# -gt 0 ]]; do
  case $1 in
    -c|--index-file) # 对日志进行解析生成的索引文件，保存了日志中的分钟，及该分钟的记录数和字节数
      indexfile="$2"
      shift
      shift
      ;; 
    --logfile) # 待检索的日志名称
      logfile="$2"
      shift
      shift
      ;;
    -f|--from) # 待检索的开始时间
      from="$2"
      shift
      shift
      ;;
    -t|--to) # 待检索的结束时间
      to="$2"
      shift
      shift
      ;;
    -u|--lines-until) # 
      lines_until="$2"
      shift
      shift
      ;;
    --refresh-index)
      refresh_index="1"
      shift
      shift
      ;;
    -l|--max-num-lines)
      max_num_lines="$2"
      shift
      shift
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      positional_args+=("$1")
      shift
      ;;        
  esac
done

set -- "${positional_args[@]}"

gawk_binary="$(find_gawk_binary)"
if [[ $? != 0 ]]; then
  echo "E:1:gawk (GNU Awk) is a requirement, but not found on the system. Please install it, then retry"
  exit 1
fi

os_kind="$(detect_os_kind)"
if [[ $? != 0 ]]; then
  echo "E:1:unknown kernel name $(uname -s)" 2>&1
  exit 1
fi

if [[ "$logfile" == "" ]]; then
  echo "E:1:--logfile is required. Specify the log file manually"
  exit 1
fi

command="$1"
if [[ "${command}" == "" ]]; then
  echo "E:1:command is required"
  exit 1
fi

case "${command}" in 
  query)
    shift
    ;;
  logstream_info)
    host_timezone="$(detect_timezone)"
    if [[ $? == 0 ]]; then
      echo "S:tz:$host_timezone"
    else 
      echo "E:1:failed to detect host timezone"
    fi

    if [ ! -e ${logfile} ]; then
      echo "E:1:${logfile} does not exist"
      exit 1
    fi

    if [ ! -r ${logfile} ]; then
      echo "E:1:${logfile} exists but is not readable, check your permissions"
      exit 1
    fi

    if [ -s ${logfile} ]; then
      last_line="$(tail -n 1 ${logfile})" || exit 1
      first_line="$(head -n 1 ${logfile})" || exit 1
      echo "S:lastLogline:$last_line"
      echo "S:firstLogline:$first_line"
    fi

    exit 0
    ;;
  *)
    echo "E:1:invalid command ${command}"
    exit 1
esac

## queryLogs 必须指定索引文件，此时需要完全重构索引文件
if [[ $indexfile == "" ]]; then
  echo "E:1:-c|--index-file is not set"
  exit 1
fi

source "$(dirname "${BASH_SOURCE[0]}")/agent_index.sh"
source "$(dirname "${BASH_SOURCE[0]}")/agent_search.sh"

user_pattern=$1

logfile_size=$(get_file_size $os_kind $logfile) || exit 1

build_index || exit 1

is_outside_of_range=0
if [[ "$from" != "" || "$to" != "" ]]; then
  refresh_and_retry=0

  if [ -s "$indexfile" ]; then
    if [[ "$from" != "" ]]; then
      read -r from_result from_linenr from_bytenr <<<$(get_linenr_and_bytenr_from_index "$from") || exit 1
      if [[ "$from_result" != "found" ]]; then
        echo "N:the from ${from} isn't found, gonna refresh the index"
        refresh_and_retry=1
      fi
    fi

    if [[ "$to" != "" ]]; then
      read -r to_result to_linenr to_bytenr <<<$(get_linenr_and_bytenr_from_index "$to") || exit 1
      if [[ "$to_result" != "found" ]]; then
        echo "N:the to ${to} isn't found, gonna refresh the index"
        refresh_and_retry=1
      fi
    fi
  else 
    echo "N:index file doesn't exist or is empty, gonna refresh it"  
    refresh_and_retry=1
  fi  

  if [[ "$refresh_and_retry" == 1 ]]; then
    build_index || exit 1

    if [[ "$from" != "" ]]; then
      read -r from_result from_linenr from_bytenr <<<$(get_linenr_and_bytenr_from_index "$from") || exit 1
      if [[ "$from_result" == "before" ]]; then
        echo "N:the from ${from} isn't found, will use the beginning"
      elif [[ "$from_result" == "found" ]]; then
        echo "N:the from ${from} is found: $from_linenr ($from_bytenr)"
        if [[ "$from_linenr" == "" || "$from_bytenr" == "" ]]; then
          echo "E:1:from_result is found but from_bytenr and/or from_linenr is empty"
          exit 1
        fi
      elif [[ "$from_result" == "after" ]]; then
        echo "N:the from ${from} is after the latest log we have, will return nothing"
        is_outside_of_range=1
      else
        echo "error:invalid from_result: $from_result"
        exit 1  
      fi
    fi

    if [[ "$to" != "" ]]; then
      read -r to_result to_linenr to_bytenr <<<$(get_linenr_and_bytenr_from_index "$to") || exit 1
      if [[ "$to_result" == "after" ]]; then
        echo "N:the to ${to} isn't found, will use the end"
      elif [[ "$to_result" == "found" ]]; then
        echo "N:the to ${to} is found: $to_linenr ($to_bytenr)"
        if [[ "$to_linenr" == "" || "$to_bytenr" == "" ]]; then
          echo "E:1:to_result is found but to_bytenr and/or to_linenr is empty"
          exit 1
        fi
      elif [[ "$to_result" == "before" ]]; then
        echo "N:the to ${to} is before the first log we have, will return nothing"
        is_outside_of_range=1
      else
        echo "E:1:invalid to_result: $to_result"
        exit 1    
      fi
    fi
  fi
else 
  if ! [ -s $indexfile ]; then
    echo "N:neither --from or --to are given, but index doesn't exist at all, gonna rebuild"
    build_index || exit 1
  fi
fi

if [[ $is_outside_of_range == 1 ]]; then
  echo "N:stage:$STAGE_DONE:done"
  exit 0
fi

echo "N:stage:$STAGE_QUERYING:querying logs"

from_linenr_int=1
if [[ "$from_linenr" != "" ]]; then
  from_linenr_int=$from_linenr
fi

num_bytes_to_scan=0
if [[ "$from_bytenr" == "" && "$to_bytenr" == "" ]]; then
  num_bytes_to_scan=$logfile_size
elif [[ "$from_bytenr" != "" && "$to_bytenr" == "" ]]; then
  num_bytes_to_scan=$((logfile_size-from_bytenr))
elif [[ "$from_bytenr" == "" && "$to_bytenr" != "" ]]; then    
  num_bytes_to_scan=$((logfile_size-to_bytenr))
else
  num_bytes_to_scan=$((to_bytenr-from_bytenr))  
fi

declare -a cmds
if [[ "$from_bytenr" != "" ]]; then
  if [[ "$to_bytenr" != "" ]]; then
    echo "N:Getting logs from offset $from_bytenr, only $((to_bytenr-from_bytenr)) bytes, all in the $logfile"
    cmds+=("tail -c +$from_bytenr $logfile | head -c $((to_bytenr-from_bytenr))")
  else 
    echo "N:Getting logs from offset $from_bytenr util the end of latest $logfile"  
    cmds+=("tail -c +$from_bytenr $logfile")
  fi
elif [[ "$to_bytenr" != "" ]]; then
  if [[ "$from_bytenr" != "" ]]; then
    echo "N:Getting logs from offset $from_bytenr, only $((to_bytenr-from_bytenr)) bytes, all in the $logfile"
    cmds+=("tail -c +$from_bytenr $logfile | head -c $((to_bytenr-from_bytenr))")
  else
    echo "N:Getting logs from the very beginning to offset $((to_bytenr-1))"
    cmds+=("head -c $((to_bytenr-1)) $logfile")
  fi
else
  echo "N:Getting logs from the very begining in $logfile"
  cmds+=("cat $logfile")
fi

cmds_concatenated="$(concat_cmds_array)"
echo "N:Command to filter logs by time range:"
echo "N:bash -c '$cmds_concatenated'"
echo "N:pattern is $user_pattern"

eval $cmds_concatenated | \
  user_pattern="$user_pattern" \
  max_num_lines="$max_num_lines" \
  num_bytes_to_scan="$num_bytes_to_scan" \
  run_search -

codes=(${PIPESTATUS[@]})
for status in "${codes[@]}"; do
  if [[ $status -ne 0 ]]; then
    exit 1
  fi
done

echo "N:stage:$STAGE_DONE:done"
run_awk_index_script () {
  awk_index_script='
    '$awk_funcs'

    BEGIN {
      if ("'$lastTimeStr'" != "") {
        lastHHMM = substr("'$lastTimeStr'", 12, 5);
      } else {
        lastHHMM = ""
      }
      
      bytenr_next = 1;
      lastPercent = 0;
      size_to_index = '$((total_size-last_bytenr))';
    }

    {
      bytenr_next += length($0) + 1;
      if (validPrefix()) {
        curHHMM = awktime_hhmm();
      }
    }

    (lastHHMM != curHHMM) {
      bytenr_cur = bytenr_next - length($0) - 1;    # 计算截至该行之前的累计字节数
      month = awktime_month();                      # 提取该行的月份 01
      year = awktime_year();                        # 提取该行的年份 2006
      day = awktime_day();                          # 提取该行的天 01
      hhmm = awktime_hhmm();                        # 提取该行的时分 10:22
      curTimestr = year "-" month "-" day "-" hhmm; # 拼接为 2006-01-01-10:22 这种格式
      if (curTimestr < lastTimestr) {               # 如果当前时间小于上一个时间，即代表该日志的时间是降序的，则不处理
        next;
      }

      printIndexLine("'$indexfile'", curTimestr, NR + '$((last_linenr-1))', bytenr_cur + '$((last_bytenr-1))');
      printPercentage(bytenr_cur, size_to_index);
      lastTimestr = curTimestr;
      lastHHMM = curHHMM;
    }
  '

  "$gawk_binary" -b "$awk_index_script" "$@"
}

build_index() {
  if [[ "$refresh_index" == "1" ]]; then
    rm -f $indexfile || exit 1
  fi

  local total_size=$(get_file_size $logfile) || exit 1

  if [ -s $indexfile ]; then # append index
    echo "p:stage:$STAGE_INDEX_APPEND:indexing up" 1>&2

    local lastTimeStr="$(tail -n 1 $indexfile | cut -f2)"
    local last_linenr="$(tail -n 1 $indexfile | cut -f3)"
    local last_bytenr="$(tail -n 1 $indexfile | cut -f4)"
    local size_to_index=$((total_size-last_bytenr))

    eval tail -c +$last_bytenr $logfile | \
      lastTimeStr=$lastTimeStr \
      total_size=$total_size \
      last_bytenr=$last_bytenr \
      last_linenr=$last_linenr \
      run_awk_index_script -

    if [[ "$?" != 0 ]]; then
      echo "debug:failed to index up, removing index file" 1>&2
      rm &indexfile
      exit 1
    fi
  else # rebuild index
    echo "p:stage:$STAGE_INDEX_FULL:indexing from scratch" 1>&2
    
    eval cat $logfile | \
      lastTimeStr="" \
      total_size=$total_size \
      last_bytenr=1 \
      last_linenr=1 \
      run_awk_index_script -

    if [[ "$?" != 0 ]]; then
      echo "debug:failed to index from scratch $logfile, removing index file" 1>&2
      rm $indexfile
      exit 1
    fi
  fi
}

get_linenr_and_bytenr_from_index() {
  awk_script='
    BEGIN {
      isFirstIdx = 1;
      printed = 0;
    }

    $1 == "idx" {
      if ("'$1'" == $2) {
        print "found " $3 " " $4;
        printed = 1;
        exit
      } else if ("'$1'" < $2) {
        if (isFirstIdx) {
          print "before";
        } else {
          print "found " $3 " " $4;
        }
        printed = 1;
        exit
      } else {
        isFirstIdx = 0;
      }
    }

    END {
      if (!printed) {
        print "after";
      }
    }
  '

  "$gawk_binary" -F"\t" "$awk_script" $indexfile
}


run_search() {
  awk_pattern=''
  if [[ "$user_pattern" != "" ]]; then
    awk_pattern="!($user_pattern) {numFilteredOut++; next}"
  fi

  echo "N:awk_pattern is $awk_pattern"

  awk_search_script='
    '$awk_funcs'

    BEGIN {
      byteNr=1;
      curLine=0;
      maxLines='$max_num_lines';
      lastPercent=0;
      numFilteredOut=0;
      num_bytes_to_scan='$num_bytes_to_scan';
      currentLog="";
      print "N:max line is " maxLines
    }

    {
      bytenr += length($0) + 1;
    }

    NR % 100 == 0 {
      printPercentage(byteNr, num_bytes_to_scan)
    }

    '$awk_pattern'

    '$lines_util_check'

    {
      if (validPrefix()) {  
        # 先保存上一条完整日志
        if (currentLog != "") {
          lastLines[curLine] = currentLog;
          lastNRs[curLine] = prevNR;
          curLine++;
          # print "curLine result " curLine
          # 当 curLine 达到 maxLines 时，重置为 0，覆盖最旧的日志
          if (curLine >= maxLines) {
            curLine = 0;
          }
        }

        currentLog=$0;
        prevNR = NR;

        # 开始新的日志行
        curMinKey = substr($0, 1, 16);
        # print "N:current min key is " curMinKey 
        stats[curMinKey]++;
      } else {
        # 续行：追加到当前日志，用空格分离
        currentLog = sprintf("%s%c%s", currentLog, 0, $0);
      }
    }

    END {
      if (currentLog != "") {
        lastLines[curLine] = currentLog;
        lastNRs[curLine] = prevNR;
        curLine++;
      }

      print "N:Filtered out " numFilteredOut " from " NR " lines" > "/dev/stderr"

      for (x in stats) {
        print "T:" x ":" stats[x]
      }

      for (x in lastLines) {
        print "N:" x ":" lastNRs[x] ":" lastLines[x]
      }

      for (i = 0; i < maxLines; i++) {
        # 从当前位置读取值，如果超过 maxLines 则回绕到开头。这样确保在限制maxLines时，只返回最新的日志行。
        # 较旧的日志行可以通过 loadEarlier => lineUtil 来访问
        ln = curLine + i;
        if (ln >= maxLines) {
          ln -= maxLines
        }

        if (!lastLines[ln]) {
          continue;
        }

        curNR = lastNRs[ln];
        line = lastLines[ln];
        print "D:" curNR ":" line
      }
    }
  '

  "$gawk_binary" -b "$awk_search_script" "$@"
  if [[ "$?" != 0 ]]; then
    return 1
  fi
}

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

    {
      if (validPrefix()) {  
        # 先保存上一条完整日志
        if (currentLog != "") {
          lastLines[curLine] = currentLog;
          lastNRs[curLine] = prevNR;
          curLine++;
          # print "curLine result " curLine
          if (curLine >= maxLines) {
            currentLog="";
            exit;
          }
        }

        # 开始新的日志行
        curMinKey = substr($0, 1, 16);
        # print "N:current min key is " curMinKey 
        stats[curMinKey]++;
        currentLog=$0;
        prevNR = NR;
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

      for (i = 0; i < curLine; i++) {
        curNR = lastNRs[i];
        line = lastLines[i];
        print "D:" curNR ":" line
      }
    }
  '

  "$gawk_binary" -b "$awk_search_script" "$@"
  if [[ "$?" != 0 ]]; then
    return 1
  fi
}

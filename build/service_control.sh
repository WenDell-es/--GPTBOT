#!/bin/bash

START_COMMENT='sudo ./gpt_server'
date_str=$(date +%Y-%m-%d)
echo $date_str

function start(){
    echo ""
    echo "start_comment : $START_COMMENT"
    echo ""
    nohup $START_COMMENT >$date_str.log 2>&1 &
    echo ""
}

function stop(){
    echo ""
    echo "comment : $START_COMMENT"
    pid=$(ps -ef | grep "$START_COMMENT" | grep -v grep | awk -F ' ' '{print $2}')
    kill $pid
    echo ""
}

function restart(){
    stop
    start
}

function status(){
    echo "start_comment : $START_COMMENT"
    ps -ef | grep "$START_COMMENT" | grep -v 'grep'
}

function help(){
    echo ""
    echo "./脚本名称.sh start    启动服务"
    echo "./脚本名称.sh stop     停止服务"
    echo "./脚本名称.sh restart  重启服务"
    echo "./脚本名称.sh status   查看服务状态"
    echo ""
}



case $1 in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    *)
        help
        ;;
esac
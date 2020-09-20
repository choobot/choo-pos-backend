#!/bin/sh

source .env

heroku container:login
heroku container:push web --app=$HEROKU_APP
PROD_DATA_SOURCE_NAME="$(heroku config --app $HEROKU_APP | grep CLEARDB_DATABASE_URL | sed -e 's/ //g' | sed -e 's/CLEARDB_DATABASE_URL://g' | sed -e 's/mysql:\/\///g' | sed -e 's/\@/\@tcp(/g' | sed -e 's/\//)\//g' | sed -e 's/reconnect=true/parseTime=true/g')"
heroku config:set LINE_LOGIN_ID=$PROD_LINE_LOGIN_ID LINE_LOGIN_SECRET=$PROD_LINE_LOGIN_SECRET DATA_SOURCE_NAME="$PROD_DATA_SOURCE_NAME" --app=$HEROKU_APP
heroku container:release web --app=$HEROKU_APP
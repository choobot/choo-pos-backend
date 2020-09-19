#!/bin/sh

source .env

heroku container:login
heroku container:push web --app=$HEROKU_APP
heroku config:set LINE_LOGIN_ID=$PROD_LINE_LOGIN_ID LINE_LOGIN_SECRET=$PROD_LINE_LOGIN_SECRET LINE_LOGIN_REDIRECT_URL=$PROD_LINE_LOGIN_REDIRECT_URL --app=$HEROKU_APP
heroku container:release web --app=$HEROKU_APP